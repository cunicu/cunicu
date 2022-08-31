package nodes

import (
	"fmt"
	"os/exec"
	"time"

	g "github.com/stv0g/gont/pkg"
	"golang.org/x/sys/unix"
)

const (
	KillTimeout = 10 * time.Minute
)

type Node interface {
	g.Node

	Start(binary, dir string, args ...any) error
	Stop() error
	Close() error
}

func GracefullyTerminate(cmd *exec.Cmd) error {
	if err := cmd.Process.Signal(unix.SIGTERM); err != nil {
		return err
	}

	// Forcefully kill agent if it did not terminate after 10secs
	timer := time.AfterFunc(KillTimeout, func() {
		if err := cmd.Process.Kill(); err != nil {
			panic(fmt.Errorf("failed to kill process: %w", err))
		}
	})
	defer timer.Stop()

	return cmd.Wait()
}
