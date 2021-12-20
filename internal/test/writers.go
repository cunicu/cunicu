//go:build linux
// +build linux

package test

import (
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func LogWriter(logger *log.Entry, stdout io.Reader, stderr io.Reader) {
	go io.Copy(logger.Writer(), stdout)
	go io.Copy(logger.Writer(), stderr)
}

func StdWriter(stdout io.Reader, stderr io.Reader) {
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
}

func FileWriter(fn string, stdout io.Reader, stderr io.Reader) (*os.File, error) {
	dir := filepath.Dir(fn)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	wr, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	go io.Copy(wr, stdout)
	go io.Copy(wr, stderr)

	return wr, err
}
