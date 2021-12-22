package pb

import "time"

var Ok = Error{
	Ok: true,
}

func TimeNow() *Timestamp {
	t := &Timestamp{}
	t.Set(time.Now())
	return t
}

func (t *Timestamp) Set(s time.Time) {
	t.Nanos = int32(s.Nanosecond())
	t.Seconds = s.Unix()
}

func (t *Timestamp) Time() time.Time {
	return time.Unix(t.Seconds, int64(t.Nanos))
}
