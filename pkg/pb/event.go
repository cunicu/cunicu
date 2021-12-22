package pb

import "github.com/sirupsen/logrus"

func (e *Event) Log(l logrus.FieldLogger, fmt string, args ...interface{}) {
	f := logrus.Fields{
		"type":  e.Type,
		"state": e.State,
	}

	if e.Time != nil {
		f["time"] = e.Time.Time()
	}

	l.WithFields(f).Infof(fmt, args...)
}

func (e *Event) Match(o *Event) bool {
	if e.Type != o.Type {
		return false
	}

	if e.State != o.State {
		return false
	}

	return true
}
