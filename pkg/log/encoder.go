// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-FileCopyrightText: 2019 Matthew Sykes
// SPDX-FileCopyrightText: 2017 Jonathan Sternberg
// SPDX-License-Identifier: MIT

package log

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/tty"
)

const (
	// For JSON-escaping; see logfmtEncoder.safeAddString below.
	_hex = "0123456789abcdef"
)

//nolint:gochecknoglobals
var (
	widthLevel atomic.Int32
	widthName  atomic.Int32

	encoderPool = sync.Pool{
		New: func() any {
			e := &encoder{}
			e.reset()
			return e
		},
	}
	bufferPool = buffer.NewPool()

	ColorTime              = tty.Color256(223)
	ColorContext           = tty.Color256(249)
	ColorStacktrace        = tty.Color256(249)
	ColorName              = tty.Color256(145)
	ColorCaller            = tty.Color256(15)
	ColorLevelUnknown      = tty.Color256(1)
	ColorLevelDebugVerbose = tty.Color256(138)
	ColorLevels            = map[Level]string{
		DebugLevel:  tty.Color256(138),
		InfoLevel:   tty.Color256(12),  // Blue
		WarnLevel:   tty.Color256(208), // Yellow
		ErrorLevel:  tty.Color256(124), // Red
		DPanicLevel: tty.Color256(196), // Red
		PanicLevel:  tty.Color256(196), // Red
		FatalLevel:  tty.Color256(196), // Red
	}
)

func ResetWidths() {
	widthLevel.Store(0)
	widthName.Store(0)
}

func ColorLevel(l Level) string {
	if color, ok := ColorLevels[l]; ok {
		return color
	} else if l < DebugLevel {
		return ColorLevelDebugVerbose
	}

	return ColorLevelUnknown
}

type encoderConfig struct {
	ColorTime       string
	ColorContext    string
	ColorStacktrace string
	ColorName       string
	ColorCaller     string
	ColorLevel      func(lvl Level) string

	Function   bool
	Message    bool
	Time       bool
	Context    bool
	Stacktrace bool
	Name       bool
	Caller     bool
	Level      bool

	// Configure the primitive representations of common complex types. For
	// example, some users may want all time.Times serialized as floating-point
	// seconds since epoch, while others may prefer ISO8601 strings.
	EncodeLevel    zapcore.LevelEncoder
	EncodeTime     zapcore.TimeEncoder
	EncodeDuration zapcore.DurationEncoder
	EncodeCaller   zapcore.CallerEncoder

	// Unlike the other primitive type encoders, EncodeName is optional. The
	// zero value falls back to FullNameEncoder.
	EncodeName zapcore.NameEncoder

	// Configures the field separator used by the console encoder. Defaults
	// to tab.
	ConsoleSeparator string

	LineEnding string
}

type encoder struct {
	*encoderConfig

	buf            *buffer.Buffer
	namespaces     []string
	arraySeparator byte
	fieldSeparator byte
	quote          byte
}

func newEncoder(cfg encoderConfig) *encoder {
	e := getEncoder()
	e.encoderConfig = &cfg
	e.buf = bufferPool.Get()

	if len(e.ConsoleSeparator) > 0 {
		e.fieldSeparator = e.ConsoleSeparator[0]
	}

	return e
}

func (e *encoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	e.addKey(key)
	return e.AppendArray(arr)
}

func (e *encoder) AddBinary(key string, value []byte) {
	e.AddString(key, base64.StdEncoding.EncodeToString(value))
}

func (e *encoder) AddBool(key string, value bool) {
	e.addKey(key)
	e.AppendBool(value)
}

func (e *encoder) AddByteString(key string, value []byte) {
	e.addKey(key)
	e.AppendByteString(value)
}

func (e *encoder) AddComplex64(k string, v complex64) {
	e.AddComplex128(k, complex128(v))
}

func (e *encoder) AddComplex128(key string, value complex128) {
	e.addKey(key)
	e.AppendComplex128(value)
}

func (e *encoder) AddDuration(key string, value time.Duration) {
	e.addKey(key)
	e.AppendDuration(value)
}

func (e *encoder) AddFloat32(key string, value float32) {
	e.addKey(key)
	e.AppendFloat32(value)
}

func (e *encoder) AddFloat64(key string, value float64) {
	e.addKey(key)
	e.AppendFloat64(value)
}

func (e *encoder) AddInt(k string, v int) {
	e.AddInt64(k, int64(v))
}

func (e *encoder) AddInt8(k string, v int8) {
	e.AddInt64(k, int64(v))
}

func (e *encoder) AddInt32(k string, v int32) {
	e.AddInt64(k, int64(v))
}

func (e *encoder) AddInt16(k string, v int16) {
	e.AddInt64(k, int64(v))
}

func (e *encoder) AddInt64(key string, value int64) {
	e.addKey(key)
	e.AppendInt64(value)
}

func (e *encoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	e.addKey(key)
	return e.AppendObject(obj)
}

func (e *encoder) AddReflected(key string, value any) error {
	e.addKey(key)
	return e.AppendReflected(value)
}

func (e *encoder) AddString(key, value string) {
	e.addKey(key)
	e.AppendString(value)
}

func (e *encoder) AddTime(key string, value time.Time) {
	e.addKey(key)
	e.AppendTime(value)
}

func (e *encoder) AddUint(k string, v uint) {
	e.AddUint64(k, uint64(v))
}

func (e *encoder) AddUint8(k string, v uint8) {
	e.AddUint64(k, uint64(v))
}

func (e *encoder) AddUint32(k string, v uint32) {
	e.AddUint64(k, uint64(v))
}

func (e *encoder) AddUint16(k string, v uint16) {
	e.AddUint64(k, uint64(v))
}

func (e *encoder) AddUintptr(k string, v uintptr) {
	e.AddUint64(k, uint64(v))
}

func (e *encoder) AddUint64(key string, value uint64) {
	e.addKey(key)
	e.AppendUint64(value)
}

func (e *encoder) AppendBool(value bool) {
	if value {
		e.AppendString("true")
	} else {
		e.AppendString("false")
	}
}

func (e *encoder) AppendByteString(value []byte) {
	e.addSeparator(e.arraySeparator, '[')

	if value == nil {
		e.buf.AppendString("<nil>")
	} else if len(value) > 0 {
		needsQuotes := bytes.IndexFunc(value, needsQuotedValueRune) != -1
		if needsQuotes {
			e.buf.AppendByte('"')
		}
		e.safeAppendByteString(value)
		if needsQuotes {
			e.buf.AppendByte('"')
		}
	}
}

func (e *encoder) AppendComplex64(v complex64) {
	e.AppendComplex128(complex128(v))
}

func (e *encoder) AppendComplex128(value complex128) {
	e.addSeparator(e.arraySeparator, '[')

	// Cast to a platform-independent, fixed-size type.
	r, i := real(value), imag(value)
	e.buf.AppendFloat(r, 64)
	e.buf.AppendByte('+')
	e.buf.AppendFloat(i, 64)
	e.buf.AppendByte('i')
}

func (e *encoder) AppendDuration(value time.Duration) {
	cur := e.buf.Len()
	if e.EncodeDuration != nil {
		e.EncodeDuration(value, e)
	}
	if cur == e.buf.Len() {
		e.AppendInt64(int64(value))
	}
}

func (e *encoder) AppendFloat32(v float32) {
	e.appendFloat(float64(v), 32)
}

func (e *encoder) AppendFloat64(v float64) {
	e.appendFloat(v, 64)
}

func (e *encoder) appendFloat(val float64, bitSize int) {
	e.addSeparator(e.arraySeparator, '[')

	switch {
	case math.IsNaN(val):
		e.buf.AppendString(`NaN`)
	case math.IsInf(val, 1):
		e.buf.AppendString(`+Inf`)
	case math.IsInf(val, -1):
		e.buf.AppendString(`-Inf`)
	default:
		e.buf.AppendFloat(val, bitSize)
	}
}

func (e *encoder) AppendInt(v int) {
	e.AppendInt64(int64(v))
}

func (e *encoder) AppendInt8(v int8) {
	e.AppendInt64(int64(v))
}

func (e *encoder) AppendInt16(v int16) {
	e.AppendInt64(int64(v))
}

func (e *encoder) AppendInt32(v int32) {
	e.AppendInt64(int64(v))
}

func (e *encoder) AppendInt64(value int64) {
	e.addSeparator(e.arraySeparator, '[')
	e.buf.AppendInt(value)
}

func (e *encoder) AppendArray(arr zapcore.ArrayMarshaler) (err error) {
	f := e.clone()
	defer putEncoder(f)

	f.buf.AppendByte('[')

	f.arraySeparator = ','
	if err := arr.MarshalLogArray(f); err != nil {
		return err
	}

	f.buf.AppendByte(']')

	e.addSeparator(e.arraySeparator, '[')
	_, err = e.buf.Write(f.buf.Bytes())
	return err
}

func (e *encoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	f := e.clone()
	defer putEncoder(f)

	f.buf.AppendByte('{')

	if err := obj.MarshalLogObject(f); err != nil {
		return err
	}

	f.buf.AppendByte('}')

	e.addSeparator(e.arraySeparator, '[')
	_, err := e.buf.Write(f.buf.Bytes())
	return err
}

func (e *encoder) AppendReflected(value any) (err error) {
	switch v := value.(type) {
	case nil:
		e.AppendString("<nil>")
	case error:
		e.AppendString(v.Error())
	case []byte:
		e.AppendString(base64.StdEncoding.EncodeToString(v))
	case bool:
		e.AppendBool(v)
	case int:
		e.AppendInt(v)
	case int8:
		e.AppendInt8(v)
	case int16:
		e.AppendInt16(v)
	case int32:
		e.AppendInt32(v)
	case int64:
		e.AppendInt64(v)
	case uint:
		e.AppendUint(v)
	case uint8:
		e.AppendUint8(v)
	case uint16:
		e.AppendUint16(v)
	case uint32:
		e.AppendUint32(v)
	case uint64:
		e.AppendUint64(v)
	case uintptr:
		e.AppendUintptr(v)
	case float32:
		e.AppendFloat32(v)
	case float64:
		e.AppendFloat64(v)
	case string:
		e.AppendString(v)
	case complex64:
		e.AppendComplex64(v)
	case complex128:
		e.AppendComplex128(v)
	default:
		if doString(value) {
			switch v := value.(type) {
			case fmt.Stringer:
				e.AppendString(v.String())
			case encoding.TextMarshaler:
				var b []byte
				if b, err = v.MarshalText(); err == nil {
					e.AppendString(string(b))
				}
			case json.Marshaler:
				var b []byte
				if b, err = v.MarshalJSON(); err == nil {
					e.buf.AppendString(string(b))
				}
			default:
				err = e.appendReflection(value)
			}
		} else {
			err = e.appendReflection(value)
		}
	}

	return err
}

func doString(value any) bool {
	rtype := reflect.TypeOf(value)
	for rtype.Kind() == reflect.Interface || rtype.Kind() == reflect.Pointer {
		rtype = rtype.Elem()
	}

	pkgPath := rtype.PkgPath()

	switch {
	case rtype.Kind() != reflect.Array &&
		rtype.Kind() != reflect.Slice &&
		rtype.Kind() != reflect.Struct:
		return true
	case strings.HasPrefix(pkgPath, "github.com/vishvananda/netlink"):
		return false
	case strings.HasPrefix(pkgPath, "github.com/stv0g/cunicu/pkg/proto"):
		return false
	}

	return true
}

func (e *encoder) appendReflection(value any) error {
	rvalue := reflect.ValueOf(value)
	switch rvalue.Kind() {
	case reflect.Chan, reflect.Func:
		e.AppendString(fmt.Sprintf("%T(%p)", value, value))
	case reflect.Struct:
		return e.appendReflectedStruct(rvalue)
	case reflect.Map:
		return e.appendReflectedMap(rvalue)
	case reflect.Array, reflect.Slice:
		return e.appendReflectedSlice(rvalue)
	case reflect.Interface, reflect.Ptr:
		value := rvalue.Elem().Interface()
		return e.AppendReflected(value)
	default:
	}

	return nil
}

func (e *encoder) appendReflectedStruct(rvalue reflect.Value) error {
	rtype := rvalue.Type()

	return e.AppendObject(zapcore.ObjectMarshalerFunc(func(oe zapcore.ObjectEncoder) error {
		for i := 0; i < rtype.NumField(); i++ {
			ftype := rtype.Field(i)
			fvalue := rvalue.Field(i)
			if !ftype.IsExported() {
				continue
			}

			key, omitempty := getFieldName(ftype)
			if fvalue.IsZero() && omitempty {
				continue
			}

			value := fvalue.Interface()
			if err := oe.AddReflected(key, value); err != nil {
				return err
			}
		}
		return nil
	}))
}

func (e *encoder) appendReflectedMap(rvalue reflect.Value) error {
	return e.AppendObject(zapcore.ObjectMarshalerFunc(func(oe zapcore.ObjectEncoder) error {
		iter := rvalue.MapRange()
		for iter.Next() {
			key := iter.Key().Interface()
			value := iter.Value().Interface()
			if err := oe.AddReflected(fmt.Sprint(key), value); err != nil {
				return err
			}
		}
		return nil
	}))
}

func (e *encoder) appendReflectedSlice(rvalue reflect.Value) error {
	return e.AppendArray(zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for i := 0; i < rvalue.Len(); i++ {
			value := rvalue.Index(i).Interface()
			if err := ae.AppendReflected(value); err != nil {
				return err
			}
		}
		return nil
	}))
}

func (e *encoder) AppendString(value string) {
	e.addSeparator(e.arraySeparator, '[')

	needsQuotes := strings.IndexFunc(value, needsQuotedValueRune) != -1

	if needsQuotes && e.quote != 0 {
		e.buf.AppendByte(e.quote)
	}

	e.safeAppendString(value)

	if needsQuotes && e.quote != 0 {
		e.buf.AppendByte(e.quote)
	}
}

func (e *encoder) AppendTime(value time.Time) {
	cur := e.buf.Len()
	if e.EncodeTime != nil {
		e.EncodeTime(value, e)
	}
	if cur == e.buf.Len() {
		e.AppendInt64(value.UnixNano())
	}
}

func (e *encoder) AppendUint(v uint) {
	e.AppendUint64(uint64(v))
}

func (e *encoder) AppendUint8(v uint8) {
	e.AppendUint64(uint64(v))
}

func (e *encoder) AppendUint16(v uint16) {
	e.AppendUint64(uint64(v))
}

func (e *encoder) AppendUint32(v uint32) {
	e.AppendUint64(uint64(v))
}

func (e *encoder) AppendUintptr(v uintptr) {
	e.AppendUint64(uint64(v))
}

func (e *encoder) AppendUint64(value uint64) {
	e.addSeparator(e.arraySeparator, '[')
	e.buf.AppendUint(value)
}

func (e *encoder) Clone() zapcore.Encoder {
	c := e.clone()
	c.buf.Write(e.buf.Bytes()) //nolint:errcheck
	c.arraySeparator = e.arraySeparator
	c.fieldSeparator = e.fieldSeparator
	return c
}

func (e *encoder) clone() *encoder {
	c := getEncoder()

	c.encoderConfig = e.encoderConfig
	c.buf = bufferPool.Get()
	c.namespaces = e.namespaces
	c.quote = e.quote

	return c
}

func (e *encoder) OpenNamespace(key string) {
	key = strings.Map(keyRuneFilter, key)
	e.namespaces = append(e.namespaces, key)
}

func (e *encoder) colored(color string, cb func()) {
	if color != "" {
		e.buf.AppendString(color)
	}

	cb()

	if color != "" {
		e.buf.AppendString(tty.Reset)
	}
}

func (e *encoder) padded(width int, cb func()) int {
	start := e.buf.Len()

	cb()

	length := e.buf.Len() - start

	for ; length < width; length++ {
		e.buf.AppendByte(' ')
	}

	return length
}

func (e *encoder) aligned(w *atomic.Int32, cb func()) {
	width := w.Load()
	newWidth := e.padded(int(width), cb)
	w.CompareAndSwap(width, int32(newWidth))
}

func (e *encoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) { //nolint:gocognit
	f := e.clone()
	defer putEncoder(f)

	// Time
	if f.Time && f.EncodeTime != nil {
		f.addSeparator(e.fieldSeparator)
		f.colored(e.ColorTime, func() {
			f.EncodeTime(ent.Time, f)
		})
	}

	// Level
	if f.Level && f.EncodeLevel != nil {
		cb := func() {
			f.aligned(&widthLevel, func() {
				f.EncodeLevel(ent.Level, f)
			})
		}

		f.addSeparator(e.fieldSeparator)

		if e.ColorLevel != nil {
			f.colored(e.ColorLevel(Level(ent.Level)), cb)
		} else {
			cb()
		}
	}

	// Name
	if f.Name && ent.LoggerName != "" {
		nameEncoder := f.EncodeName
		if nameEncoder == nil {
			// Fall back to FullNameEncoder for backward compatibility.
			nameEncoder = zapcore.FullNameEncoder
		}

		f.addSeparator(e.fieldSeparator)
		f.colored(e.ColorName, func() {
			f.aligned(&widthName, func() {
				nameEncoder(ent.LoggerName, f)
			})
		})
	}

	if ent.Caller.Defined {
		if f.Caller && f.EncodeCaller != nil {
			f.addSeparator(e.fieldSeparator)
			f.colored(e.ColorCaller, func() {
				f.EncodeCaller(ent.Caller, f)
			})
		}

		if f.Function {
			f.addSeparator(e.fieldSeparator)
			f.colored(e.ColorCaller, func() {
				f.AppendString(ent.Caller.Function)
			})
		}
	}

	// Add the message itself.
	if f.Message {
		f.addSeparator(e.fieldSeparator)
		f.buf.AppendString(ent.Message)
	}

	// Add context
	if len(fields) > 0 || e.buf.Len() > 0 {
		f.addSeparator('\t')
		f.colored(e.ColorContext, func() {
			f.buf.Write(e.buf.Bytes()) //nolint:errcheck

			f.quote = '\''
			if e.buf.Len() > 0 {
				f.fieldSeparator = ' '
			} else {
				f.fieldSeparator = 0
			}

			for i := range fields {
				fields[i].AddTo(f)
				f.fieldSeparator = ' '
			}
		})
	}

	if f.Stacktrace && ent.Stack != "" {
		f.addSeparator(e.fieldSeparator)
		f.colored(e.ColorStacktrace, func() {
			f.buf.AppendString(ent.Stack)
		})
	}

	// Line ending
	if le := e.LineEnding; le == "" {
		f.buf.AppendByte('\n')
	} else {
		f.buf.AppendString(le)
	}

	return f.buf, nil
}

func (e *encoder) lastByte() (byte, bool) {
	last := e.buf.Len() - 1
	if last >= 0 {
		return e.buf.Bytes()[last], true
	}

	return 0, false
}

func (e *encoder) addSeparator(sep byte, not ...byte) {
	if sep == 0 {
		return
	}

	if last, ok := e.lastByte(); ok && !slices.Contains(not, last) {
		e.buf.AppendByte(sep)
	}
}

func (e *encoder) addKey(key string) {
	e.addSeparator(e.fieldSeparator, '{')

	for _, ns := range e.namespaces {
		e.safeAppendString(ns)
		e.buf.AppendByte('.')
	}
	key = strings.Map(keyRuneFilter, key)
	e.safeAppendString(key)
	e.buf.AppendByte('=')
}

// safeAppendString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (e *encoder) safeAppendString(s string) {
	for i := 0; i < len(s); {
		if e.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if e.tryAddRuneError(r, size) {
			i++
			continue
		}
		e.buf.AppendString(s[i : i+size])
		i += size
	}
}

// safeAppendByteString is no-alloc equivalent of safeAddString(string(s)) for s []byte.
func (e *encoder) safeAppendByteString(s []byte) {
	for i := 0; i < len(s); {
		if e.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if e.tryAddRuneError(r, size) {
			i++
			continue
		}
		e.buf.Write(s[i : i+size]) //nolint:errcheck
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (e *encoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		e.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		e.buf.AppendByte('\\')
		e.buf.AppendByte(b)
	case '\n':
		e.buf.AppendByte('\\')
		e.buf.AppendByte('n')
	case '\r':
		e.buf.AppendByte('\\')
		e.buf.AppendByte('r')
	case '\t':
		e.buf.AppendByte('\\')
		e.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		e.buf.AppendString(`\u00`)
		e.buf.AppendByte(_hex[b>>4])
		e.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (e *encoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		e.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}

func (e *encoder) reset() {
	e.encoderConfig = nil
	e.buf = nil
	e.namespaces = nil
	e.arraySeparator = 0
	e.quote = 0
	e.fieldSeparator = ' '
}

func needsQuotedValueRune(r rune) bool {
	return r <= ' ' || r == '=' || r == '"' || r == utf8.RuneError
}

func keyRuneFilter(r rune) rune {
	if needsQuotedValueRune(r) {
		return -1
	}
	return r
}

func getEncoder() *encoder {
	return encoderPool.Get().(*encoder) //nolint:forcetypeassert
}

func putEncoder(e *encoder) {
	e.reset()
	encoderPool.Put(e)
}
