// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-FileCopyrightText: 2019 Matthew Sykes
// SPDX-FileCopyrightText: 2017 Jonathan Sternberg
// SPDX-License-Identifier: MIT

package log

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/stv0g/cunicu/pkg/tty"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var errWelp = errors.New("welp")

type (
	stringer        string
	textMarshaler   string
	jsonMarshaler   string
	failedMarshaler string
)

func (s stringer) String() string                      { return string(s) }
func (t textMarshaler) MarshalText() ([]byte, error)   { return []byte(t), nil }
func (j jsonMarshaler) MarshalJSON() ([]byte, error)   { return []byte(j), nil }
func (t failedMarshaler) MarshalText() ([]byte, error) { return []byte(t), errWelp }

var _ = Context("encoder", func() {
	BeforeEach(func() {
		ResetWidths()
	})

	DescribeTable("key",
		func(
			key string,
			expected string,
		) {
			enc := newEncoder(encoderConfig{})

			enc.AddString(key, "value")
			Expect(enc.buf.String()).To(Equal(expected))

			enc.AddString("x", "y")
			Expect(enc.buf.String()).To(Equal(expected + " x=y"))
		},

		Entry("standard", `k`, `k=value`),
		Entry("back-space", `k\`, `k\\=value`),
		Entry("space", `k `, `k=value`),
		Entry("equal", `k=`, `k=value`),
		Entry("quote", `k"`, `k=value`),
		Entry("UTF-8 rune", `k`+string(utf8.RuneError), `k=value`),
	)

	Describe("namespaces", func() {
		It("simple", func() {
			enc := newEncoder(encoderConfig{})

			enc.AddString("k", "value")
			for _, ns := range []string{"one", "two", "three"} {
				enc.OpenNamespace(ns)
				enc.AddString("k", "value")
			}
			Expect(enc.buf.String()).To(Equal("k=value one.k=value one.two.k=value one.two.three.k=value"))
		})

		DescribeTable("complex",
			func(
				namespaces []string,
				expected string,
			) {
				enc := newEncoder(encoderConfig{})

				for _, ns := range namespaces {
					enc.OpenNamespace(ns)
				}
				enc.AddString("k", "value")
				Expect(enc.buf.String()).To(Equal(expected))
			},
			Entry("backspace", []string{`ns\`}, `ns\\.k=value`),
			Entry("space", []string{`ns `}, `ns.k=value`),
			Entry("equal", []string{`ns=`}, `ns.k=value`),
			Entry("quote", []string{`ns"`}, `ns.k=value`),
			Entry("UTF-8 rune", []string{`ns` + string(utf8.RuneError)}, `ns.k=value`),
		)
	})

	It("can be cloned", func() {
		enc := newEncoder(encoderConfig{})

		enc.AddString("k", "v")

		clone := enc.Clone()
		Expect(clone).To(Equal(enc), "clone should equal original encoder")

		enc.AddString("x", "y")
		Expect(clone).NotTo(Equal(enc), "clone should not equal original encoder")
	})

	Describe("encoder", func() {
		DescribeTable("simple",
			func(
				expected string,
				f func(enc zapcore.Encoder),
			) {
				enc := newEncoder(encoderConfig{})
				enc.quote = '"'

				f(enc)
				Expect(enc.buf.String()).To(Equal(expected))

				enc.AddString("another", "field")
				Expect(enc.buf.String()).To(Equal(expected+" another=field"), "with extra field")
			},
			Entry("binary", "k=YmFzZTY0", func(enc zapcore.Encoder) { enc.AddBinary("k", []byte("base64")) }),
			Entry("bool (true)", "k=true", func(enc zapcore.Encoder) { enc.AddBool("k", true) }),
			Entry("bool (false)", "k=false", func(enc zapcore.Encoder) { enc.AddBool("k", false) }),
			Entry("bytestring", "k=bytes", func(enc zapcore.Encoder) { enc.AddByteString("k", []byte("bytes")) }),
			Entry("bytestring with nil", "k=<nil>", func(enc zapcore.Encoder) { enc.AddByteString("k", nil) }),
			Entry("bytestring with rune", `k=☺`, func(enc zapcore.Encoder) { enc.AddByteString("k", []byte{0xe2, 0x98, 0xba}) }),
			Entry("complex64", "k=1+2i", func(enc zapcore.Encoder) { enc.AddComplex64("k", 1+2i) }),
			Entry("complex128", "k=2+3i", func(enc zapcore.Encoder) { enc.AddComplex128("k", 2+3i) }),
			Entry("float32", "k=3.2", func(enc zapcore.Encoder) { enc.AddFloat32("k", 3.2) }),
			Entry("float32 +Inf", "k=+Inf", func(enc zapcore.Encoder) { enc.AddFloat32("k", float32(math.Inf(1))) }),
			Entry("float32 -Inf", "k=-Inf", func(enc zapcore.Encoder) { enc.AddFloat32("k", float32(math.Inf(-1))) }),
			Entry("float32 NaN", "k=NaN", func(enc zapcore.Encoder) { enc.AddFloat32("k", float32(math.NaN())) }),
			Entry("float64", "k=6.4", func(enc zapcore.Encoder) { enc.AddFloat64("k", 6.4) }),
			Entry("float64 +Inf", "k=+Inf", func(enc zapcore.Encoder) { enc.AddFloat64("k", math.Inf(1)) }),
			Entry("float64 -Inf", "k=-Inf", func(enc zapcore.Encoder) { enc.AddFloat64("k", math.Inf(-1)) }),
			Entry("float64 NaN", "k=NaN", func(enc zapcore.Encoder) { enc.AddFloat64("k", math.NaN()) }),
			Entry("int", "k=-1", func(enc zapcore.Encoder) { enc.AddInt("k", -1) }),
			Entry("int8", "k=-8", func(enc zapcore.Encoder) { enc.AddInt8("k", -8) }),
			Entry("int16", "k=-16", func(enc zapcore.Encoder) { enc.AddInt16("k", -16) }),
			Entry("int32", "k=-32", func(enc zapcore.Encoder) { enc.AddInt32("k", -32) }),
			Entry("int64", "k=-64", func(enc zapcore.Encoder) { enc.AddInt64("k", -64) }),
			Entry("string", "k=string", func(enc zapcore.Encoder) { enc.AddString("k", "string") }),
			Entry("string with spaces", `k="string with spaces"`, func(enc zapcore.Encoder) { enc.AddString("k", "string with spaces") }),
			Entry("string with quotes", `k="\"quoted string\""`, func(enc zapcore.Encoder) { enc.AddString("k", `"quoted string"`) }),
			Entry("string with backslash", `k=\\back\\`, func(enc zapcore.Encoder) { enc.AddString("k", `\back\`) }),
			Entry("string with newline", `k="new\nline"`, func(enc zapcore.Encoder) { enc.AddString("k", "new\nline") }),
			Entry("string with cr", `k="carriage\rreturn"`, func(enc zapcore.Encoder) { enc.AddString("k", "carriage\rreturn") }),
			Entry("string with tab", `k="tab\ttab"`, func(enc zapcore.Encoder) { enc.AddString("k", "tab\ttab") }),
			Entry("string with control char", `k="control\u0000char"`, func(enc zapcore.Encoder) { enc.AddString("k", "control\u0000char") }),
			Entry("string with rune", `k=☺`, func(enc zapcore.Encoder) { enc.AddString("k", "☺") }),
			Entry("string with decode error", `k="\ufffd"`, func(enc zapcore.Encoder) { enc.AddString("k", string([]byte{0xe2})) }),
			Entry("uint", "k=1", func(enc zapcore.Encoder) { enc.AddUint("k", 1) }),
			Entry("uint8", "k=8", func(enc zapcore.Encoder) { enc.AddUint8("k", 8) }),
			Entry("uint16", "k=16", func(enc zapcore.Encoder) { enc.AddUint16("k", 16) }),
			Entry("uint32", "k=32", func(enc zapcore.Encoder) { enc.AddUint32("k", 32) }),
			Entry("uint64", "k=64", func(enc zapcore.Encoder) { enc.AddUint64("k", 64) }),
			Entry("uintptr", "k=128", func(enc zapcore.Encoder) { enc.AddUintptr("k", 128) }),
			Entry("duration", "k=1", func(enc zapcore.Encoder) { enc.AddDuration("k", time.Nanosecond) }),
			Entry("time", "k=0", func(enc zapcore.Encoder) { enc.AddTime("k", time.Unix(0, 0)) }),
		)

		dummyFunc := func(string) {}
		dummyCh := make(chan struct{})

		DescribeTable("reflected",
			func(
				expected string,
				value any,
			) {
				enc := newEncoder(encoderConfig{})

				err := enc.AddReflected("k", value)
				Expect(err).To(Succeed())
				Expect(enc.buf.String()).To(Equal("k=" + expected))
			},
			Entry("nil", "<nil>", nil),
			Entry("error", "welp", errWelp),
			Entry("bytes", "Ynl0ZXM=", []byte("bytes")),
			Entry("stringer", "my-stringer", stringer("my-stringer")),
			Entry("text marshaler", "marshaled-text", textMarshaler("marshaled-text")),
			Entry("json marshaler", `{"json":"data"}`, jsonMarshaler(`{"json":"data"}`)),
			Entry("bool", "true", true),
			Entry("int", "-1", -int(1)),
			Entry("int8", "-8", int8(-8)),
			Entry("int16", "-16", int8(-16)),
			Entry("int32", "-32", int8(-32)),
			Entry("int64", "-64", int8(-64)),
			Entry("uint", "1", uint(1)),
			Entry("uint8", "8", uint8(8)),
			Entry("uint16", "16", uint8(16)),
			Entry("uint32", "32", uint8(32)),
			Entry("uint64", "64", uint8(64)),
			Entry("float32", "3.2", float32(3.2)),
			Entry("float64", "6.4", float64(6.4)),
			Entry("string", "string", "string"),
			Entry("complex64", "1+2i", complex64(1+2i)),
			Entry("complex128", "2+3i", complex128(2+3i)),
			Entry("chan", fmt.Sprintf(`%T(%p)`, dummyCh, dummyCh), dummyCh),
			Entry("func", fmt.Sprintf("%T(%p)", dummyFunc, dummyFunc), dummyFunc),
			Entry("slice", "[0,1,2,3]", []int{0, 1, 2, 3}),
			Entry("map", `{a=0}`, map[string]int{"a": 0}),
			Entry("array", "[one,two]", [2]string{"one", "two"}),
			Entry("ptr", "{}", &struct{}{}),
		)

		It("reflected failed", func() {
			enc := newEncoder(encoderConfig{})

			err := enc.AddReflected("k", failedMarshaler("marshaled"))
			Expect(err).To(MatchError(errWelp))
		})

		Describe("array", func() {
			DescribeTable("simple",
				func(
					expected string,
					f func(enc zapcore.ArrayEncoder),
				) {
					enc := newEncoder(encoderConfig{})

					err := enc.AddArray("x", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
						f(enc)
						return nil
					}))
					Expect(err).To(Succeed())

					err = enc.AddArray("y", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
						f(enc)
						f(enc)
						return nil
					}))
					Expect(err).To(Succeed())

					expected = strings.ReplaceAll("x=[%%] y=[%%,%%]", "%%", expected)
					Expect(enc.buf.String()).To(Equal(expected))
				},
				Entry("bool", "true", func(enc zapcore.ArrayEncoder) { enc.AppendBool(true) }),
				Entry("bytestring", "bytes", func(enc zapcore.ArrayEncoder) { enc.AppendByteString([]byte("bytes")) }),
				Entry("complex64", "1+2i", func(enc zapcore.ArrayEncoder) { enc.AppendComplex64(1 + 2i) }),
				Entry("complex128", "2+3i", func(enc zapcore.ArrayEncoder) { enc.AppendComplex128(2 + 3i) }),
				Entry("float32", "3.2", func(enc zapcore.ArrayEncoder) { enc.AppendFloat32(3.2) }),
				Entry("float64", "6.4", func(enc zapcore.ArrayEncoder) { enc.AppendFloat64(6.4) }),
				Entry("int", "-1", func(enc zapcore.ArrayEncoder) { enc.AppendInt(-1) }),
				Entry("int8", "-8", func(enc zapcore.ArrayEncoder) { enc.AppendInt8(-8) }),
				Entry("int16", "-16", func(enc zapcore.ArrayEncoder) { enc.AppendInt16(-16) }),
				Entry("int32", "-32", func(enc zapcore.ArrayEncoder) { enc.AppendInt32(-32) }),
				Entry("int64", "-64", func(enc zapcore.ArrayEncoder) { enc.AppendInt64(-64) }),
				Entry("string", "string-value", func(enc zapcore.ArrayEncoder) { enc.AppendString("string-value") }),
				Entry("uint", "1", func(enc zapcore.ArrayEncoder) { enc.AppendUint(1) }),
				Entry("uint8", "8", func(enc zapcore.ArrayEncoder) { enc.AppendUint8(8) }),
				Entry("uint16", "16", func(enc zapcore.ArrayEncoder) { enc.AppendUint16(16) }),
				Entry("uint32", "32", func(enc zapcore.ArrayEncoder) { enc.AppendUint32(32) }),
				Entry("uint64", "64", func(enc zapcore.ArrayEncoder) { enc.AppendUint64(64) }),
				Entry("uintptr", "128", func(enc zapcore.ArrayEncoder) { enc.AppendUintptr(128) }),
				Entry("duration", "1", func(enc zapcore.ArrayEncoder) { enc.AppendDuration(time.Nanosecond) }),
				Entry("time", "0", func(enc zapcore.ArrayEncoder) { enc.AppendTime(time.Unix(0, 0)) }),
				Entry("reflected", "{v=v}", func(enc zapcore.ArrayEncoder) { enc.AppendReflected(struct{ V string }{"v"}) }),               //nolint:errcheck
				Entry("reflected 2", "{a=a b=b}", func(enc zapcore.ArrayEncoder) { enc.AppendReflected(struct{ A, B string }{"a", "b"}) }), //nolint:errcheck
			)

			DescribeTable("complex",
				func(
					expected string,
					f zapcore.ArrayMarshalerFunc,
				) {
					enc := newEncoder(encoderConfig{})

					err := enc.AddArray("x", f)
					Expect(err).To(Succeed())
					Expect(enc.buf.String()).To(Equal("x=" + expected))
				},
				Entry("arrays in array",
					"[0,[1,2,3],[4,5,6],7]",
					func(enc zapcore.ArrayEncoder) error {
						enc.AppendInt(0)
						if err := enc.AppendArray(marshalIntArray(1, 3)); err != nil {
							return err
						}
						if err := enc.AppendArray(marshalIntArray(4, 6)); err != nil {
							return err
						}
						enc.AppendInt(7)
						return nil
					},
				),
				Entry("array of objects",
					"[{a=0},{b=1},{c=2}]",
					func(enc zapcore.ArrayEncoder) error {
						for i := 0; i < 3; i++ {
							if err := enc.AppendObject(zapcore.ObjectMarshalerFunc(func(oe zapcore.ObjectEncoder) error {
								oe.AddInt(string(rune('a'+i)), i)
								return nil
							})); err != nil {
								return err
							}
						}
						return nil
					},
				),
			)

			It("error", func() {
				enc := newEncoder(encoderConfig{})

				errBanana := errors.New("banana") //nolint:goerr113

				err := enc.AddArray("x", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
					return errBanana
				}))
				Expect(err).To(MatchError(errBanana))
				Expect(enc.buf.String()).To(Equal("x="))
			})
		})

		Describe("object", func() {
			DescribeTable("simple",
				func(
					expected string,
					f func(enc zapcore.ObjectEncoder),
				) {
					enc := newEncoder(encoderConfig{})

					err := enc.AddObject("x", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
						f(enc)
						return nil
					}))
					Expect(err).To(Succeed())

					expected = strings.ReplaceAll(`x={%%}`, "%%", expected)
					Expect(enc.buf.String()).To(Equal(expected))
				},
				Entry("binary", "k=YmFzZTY0", func(enc zapcore.ObjectEncoder) { enc.AddBinary("k", []byte("base64")) }),
				Entry("bytestring", "k=bytes", func(enc zapcore.ObjectEncoder) { enc.AddByteString("k", []byte("bytes")) }),
				Entry("bool (true)", "k=true", func(enc zapcore.ObjectEncoder) { enc.AddBool("k", true) }),
				Entry("bool (false)", "k=false", func(enc zapcore.ObjectEncoder) { enc.AddBool("k", false) }),
				Entry("complex64", "k=1+2i", func(enc zapcore.ObjectEncoder) { enc.AddComplex64("k", 1+2i) }),
				Entry("complex128", "k=2+3i", func(enc zapcore.ObjectEncoder) { enc.AddComplex128("k", 2+3i) }),
				Entry("duration", "k=1", func(enc zapcore.ObjectEncoder) { enc.AddDuration("k", time.Nanosecond) }),
				Entry("float32", "k=3.2", func(enc zapcore.ObjectEncoder) { enc.AddFloat32("k", 3.2) }),
				Entry("float64", "k=6.4", func(enc zapcore.ObjectEncoder) { enc.AddFloat64("k", 6.4) }),
				Entry("int", "k=-1", func(enc zapcore.ObjectEncoder) { enc.AddInt("k", -1) }),
				Entry("int8", "k=-8", func(enc zapcore.ObjectEncoder) { enc.AddInt8("k", -8) }),
				Entry("int16", "k=-16", func(enc zapcore.ObjectEncoder) { enc.AddInt16("k", -16) }),
				Entry("int32", "k=-32", func(enc zapcore.ObjectEncoder) { enc.AddInt32("k", -32) }),
				Entry("int64", "k=-64", func(enc zapcore.ObjectEncoder) { enc.AddInt64("k", -64) }),
				Entry("string", "k=string", func(enc zapcore.ObjectEncoder) { enc.AddString("k", "string") }),
				Entry("time", "k=0", func(enc zapcore.ObjectEncoder) { enc.AddTime("k", time.Unix(0, 0)) }),
				Entry("uint", "k=1", func(enc zapcore.ObjectEncoder) { enc.AddUint("k", 1) }),
				Entry("uint8", "k=8", func(enc zapcore.ObjectEncoder) { enc.AddUint8("k", 8) }),
				Entry("uint16", "k=16", func(enc zapcore.ObjectEncoder) { enc.AddUint16("k", 16) }),
				Entry("uint32", "k=32", func(enc zapcore.ObjectEncoder) { enc.AddUint32("k", 32) }),
				Entry("uint64", "k=64", func(enc zapcore.ObjectEncoder) { enc.AddUint64("k", 64) }),
				Entry("uintptr", "k=128", func(enc zapcore.ObjectEncoder) { enc.AddUintptr("k", 128) }),
				Entry("reflected", "k={v=v}", func(enc zapcore.ObjectEncoder) { enc.AddReflected("k", struct{ V string }{"v"}) }),               //nolint:errcheck
				Entry("reflected 2", `k={a=a b=b}`, func(enc zapcore.ObjectEncoder) { enc.AddReflected("k", struct{ A, B string }{"a", "b"}) }), //nolint:errcheck
			)

			It("error", func() {
				enc := newEncoder(encoderConfig{})

				errMangoTango := errors.New("mango-tango") //nolint:goerr113

				err := enc.AddObject("x", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
					return errMangoTango
				}))
				Expect(err).To(MatchError(errMangoTango))
				Expect(enc.buf.String()).To(Equal("x="))
			})

			DescribeTable("complex",
				func(
					expected string,
					f zapcore.ObjectMarshalerFunc,
				) {
					enc := newEncoder(encoderConfig{})

					err := enc.AddObject("o", f)
					Expect(err).To(Succeed())
					Expect(enc.buf.String()).To(Equal("o=" + expected))
				},
				Entry("objects with arrays",
					`{a=[1,2,3] b=[4,5]}`,
					func(enc zapcore.ObjectEncoder) error {
						if err := enc.AddArray("a", marshalIntArray(1, 3)); err != nil {
							return fmt.Errorf("failed to add array: %w", err)
						}
						if err := enc.AddArray("b", marshalIntArray(4, 5)); err != nil {
							return fmt.Errorf("failed to add array: %w", err)
						}
						return nil
					},
				),
				Entry("objects with objects ",
					`{0={a=1 b=2 c=3} 1={d=4 e=5 f=6} 2={g=7 h=8 i=9}}`,
					func(enc zapcore.ObjectEncoder) error {
						for i := 0; i < 3; i++ {
							if err := enc.AddObject(strconv.Itoa(i), zapcore.ObjectMarshalerFunc(func(oe zapcore.ObjectEncoder) error {
								for j := 0; j < 3; j++ {
									oe.AddInt(string(rune('a'+i*3+j)), i*3+j+1)
								}
								return nil
							})); err != nil {
								return err
							}
						}
						return nil
					},
				),
			)
		})
	})

	Describe("EncodeEntry", func() {
		DescribeTable("empty config with field", func(color, expected string) {
			enc := newEncoder(encoderConfig{
				ColorContext: color,
			})

			enc.AddString("x", "y")

			buf, err := enc.EncodeEntry(
				zapcore.Entry{},
				[]zapcore.Field{
					zap.String("a", "b"),
					zap.String("c", "d"),
				},
			)

			Expect(err).To(Succeed())
			Expect(buf.String()).To(Equal(expected))
		},
			Entry("with color", ColorContext, tty.Mods("x=y a=b c=d", ColorContext)+"\n"),
			Entry("without color", "", "x=y a=b c=d\n"),
		)

		DescribeTable("with config",
			func(
				expected string,
				ec encoderConfig,
				setup func(enc zapcore.Encoder),
				fields []zapcore.Field,
			) {
				enc := newEncoder(ec)

				if setup != nil {
					setup(enc)
				}

				entry := zapcore.Entry{
					Level:      zapcore.DebugLevel,
					Time:       time.Unix(1, 1),
					LoggerName: "test",
					Message:    "message",
					Caller: zapcore.EntryCaller{
						Defined: true,
						File:    "arthur/philip/dent/h2g2.go",
						Line:    42,
					},
					Stack: "stacktrace\nwith multiple lines\n\tand tabs\n",
				}

				buf, err := enc.EncodeEntry(entry, fields)
				Expect(err).To(Succeed())

				lineEnding := ec.LineEnding
				if lineEnding == "" {
					lineEnding = zapcore.DefaultLineEnding
				}

				Expect(buf.String()).To(Equal(expected + lineEnding))
			},
			Entry("empty",
				"",
				encoderConfig{},
				nil,
				nil,
			),
			Entry("empty with fields",
				"key=value",
				encoderConfig{},
				nil,
				[]zapcore.Field{zap.String("key", "value")},
			),
			Entry("empty with fields and color",
				tty.Mods("key=value", ColorContext),
				encoderConfig{ColorContext: ColorContext},
				nil,
				[]zapcore.Field{zap.String("key", "value")},
			),
			Entry("empty with context",
				"message\tcontext=value field=value",
				encoderConfig{Message: true},
				func(enc zapcore.Encoder) { enc.AddString("context", "value") },
				[]zapcore.Field{zap.String("field", "value")},
			),
			Entry("empty with context and color",
				"message\t"+tty.Mods("context=value field=value", ColorContext),
				encoderConfig{Message: true, ColorContext: ColorContext},
				func(enc zapcore.Encoder) { enc.AddString("context", "value") },
				[]zapcore.Field{zap.String("field", "value")},
			),
			Entry("EncodeTime",
				"1000000001",
				encoderConfig{Time: true, EncodeTime: zapcore.EpochNanosTimeEncoder},
				nil,
				nil,
			),
			Entry("EncodeTime with color",
				tty.Mods("1000000001", ColorTime),
				encoderConfig{Time: true, EncodeTime: zapcore.EpochNanosTimeEncoder, ColorTime: ColorTime},
				nil,
				nil,
			),
			Entry("EncodeLevel",
				"debug",
				encoderConfig{Level: true, EncodeLevel: zapcore.LowercaseLevelEncoder},
				nil,
				nil,
			),
			Entry("EncodeLevel with color",
				tty.Mods("debug", ColorLevels[DebugLevel]),
				encoderConfig{Level: true, EncodeLevel: zapcore.LowercaseLevelEncoder, ColorLevel: ColorLevel},
				nil,
				nil,
			),
			Entry("EncodeName",
				"test",
				encoderConfig{Name: true, EncodeName: zapcore.FullNameEncoder},
				nil,
				nil,
			),
			Entry("EncodeName with color",
				tty.Mods("test", ColorName),
				encoderConfig{Name: true, EncodeName: zapcore.FullNameEncoder, ColorName: ColorName},
				nil,
				nil,
			),
			Entry("EncodeCaller",
				"arthur/philip/dent/h2g2.go:42",
				encoderConfig{Caller: true, EncodeCaller: zapcore.FullCallerEncoder},
				nil,
				nil,
			),
			Entry("EncodeCalle with color",
				tty.Mods("arthur/philip/dent/h2g2.go:42", ColorCaller),
				encoderConfig{Caller: true, EncodeCaller: zapcore.FullCallerEncoder, ColorCaller: ColorCaller},
				nil,
				nil,
			),
			Entry("EncodeMessage",
				"message",
				encoderConfig{Message: true},
				nil,
				nil,
			),
			Entry("EncodeMessage with color",
				"message",
				encoderConfig{Message: true},
				nil,
				nil,
			),
			Entry("Stracktrace",
				"stacktrace\nwith multiple lines\n\tand tabs\n",
				encoderConfig{Stacktrace: true},
				nil,
				nil,
			),
			Entry("Stracktrace with color",
				tty.Mods("stacktrace\nwith multiple lines\n\tand tabs\n", ColorStacktrace),
				encoderConfig{Stacktrace: true, ColorStacktrace: ColorStacktrace},
				nil,
				nil,
			),
			Entry("LineEnding",
				"",
				encoderConfig{LineEnding: "<EOL>"},
				nil,
				nil,
			),
		)
	})

	ts := time.Unix(1, 0)

	DescribeTable("time",
		func(
			expected string,
			timeEncoder zapcore.TimeEncoder,
		) {
			enc := newEncoder(encoderConfig{EncodeTime: timeEncoder})
			enc.quote = '"'

			enc.AddTime("ts", ts)
			Expect(enc.buf.String()).To(Equal("ts=" + expected))
		},
		Entry("nil", "1000000000", nil),
		Entry("epoch millis", "1000", zapcore.EpochMillisTimeEncoder),
		Entry("custom", "custom-time", func(t time.Time, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString("custom-time") }),
		Entry("custom with spaces", `"with spaces"`, func(t time.Time, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString("with spaces") }),
	)

	duration := time.Second

	DescribeTable("duration",
		func(
			expected string,
			durationEncoder zapcore.DurationEncoder,
		) {
			enc := newEncoder(encoderConfig{EncodeDuration: durationEncoder})
			enc.quote = '"'

			enc.AddDuration("duration", duration)
			Expect(enc.buf.String()).To(Equal("duration=" + expected))
		},
		Entry("nil", "1000000000", nil),
		Entry("seconds", "1", zapcore.SecondsDurationEncoder),
		Entry("string", "1s", zapcore.StringDurationEncoder),
		Entry("custom", "custom", func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString("custom") }),
		Entry("custom with spaces", `"with spaces"`, func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString("with spaces") }),
	)
})

func marshalIntArray(start, end int) zapcore.ArrayMarshalerFunc {
	return func(enc zapcore.ArrayEncoder) error {
		for i := start; i <= end; i++ {
			enc.AppendInt(i)
		}
		return nil
	}
}
