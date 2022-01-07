package util_test

import (
	"testing"
	"time"

	"riasc.eu/wice/internal/util"
)

func TestPrettyDuration(t *testing.T) {
	if util.PrettyDuration(5*time.Hour+15*time.Minute+time.Second, false) != "5 hours, 15 minutes, 1 second" {
		t.Fail()
	}

	if util.PrettyDuration(0, false) != "" {
		t.Fail()
	}
}

func TestAgo(t *testing.T) {
	s := time.Now()

	if util.Ago(s, false) != "Now" {
		t.Fail()
	}

	s = s.Add(-time.Hour)

	if util.Ago(s, false) != "1 hour ago" {
		t.Errorf("%s", util.Ago(s, false))
	}

	s = s.Add(-10 * time.Minute)

	if util.Ago(s, false) != "1 hour, 10 minutes ago" {
		t.Errorf("%s", util.Ago(s, false))
	}
}

func TestPrettyBytes(t *testing.T) {
	if util.PrettyBytes(500, false) != "500 B" {
		t.Fail()
	}

	if util.PrettyBytes(1536, false) != "1.50 KiB" {
		t.Fail()
	}

	if util.PrettyBytes(1572864, false) != "1.50 MiB" {
		t.Fail()
	}
}

func TestEvery(t *testing.T) {
	if util.Every(5*time.Hour, false) != "every 5 hours" {
		t.Fail()
	}

	if util.Every(time.Hour, false) != "every 1 hour" {
		t.Fail()
	}
}
