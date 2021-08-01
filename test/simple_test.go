package hornet_test

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os/exec"
	"syscall"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func Killall(cmds ...*exec.Cmd) error {
	for _, cmd := range cmds {
		err := cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			return err
		}

		err = cmd.Wait()
		if err != nil {
			return err
		}
	}

	return nil
}

func SlicePrefix(prefix string, stream *io.ReadCloser) {
	scanner := bufio.NewScanner(*stream)
	for scanner.Scan() {
		fmt.Println(scanner.Text()) // Println will add back the final '\n'
	}
	if err := scanner.Err(); err != nil {
		log.WithError(err).Error("Reading stream")
	}
}

func RunWice(h *gont.Host, args ...string) (*exec.Cmd, error) {
	cmd := append([]string{"../../../cmd/wice/main.go"}, args...)
	w, stdout, stderr, err := h.GoRunAsync(cmd...)
	if err != nil {
		return nil, err
	}

	go SlicePrefix("Wice "+h.Name+": ", stdout)
	go SlicePrefix("Wice "+h.Name+": ", stderr)

	return w, nil
}

func ConfigureWireguard(h *gont.Host) error {

	return nil
}

func TestWice(t *testing.T) {
	n := gont.NewNetwork("test")
	defer n.Close()

	sw, err := n.AddSwitch("sw")
	if err != nil {
		t.Fail()
	}

	// h1, err := n.AddHost("h1", nil, &gont.Interface{"eth0", net.IPv4(10, 0, 0, 1), mask(), sw})
	// if err != nil {
	// 	t.Fail()
	// }

	// h2, err := n.AddHost("h2", nil, &gont.Interface{"eth0", net.IPv4(10, 0, 0, 2), mask(), sw})
	// if err != nil {
	// 	t.Fail()
	// }

	h3, err := n.AddHost("h3", nil, &gont.Interface{"eth0", net.IPv4(10, 0, 0, 3), mask(), sw})
	if err != nil {
		t.Fail()
	}

	b, stdout, stderr, err := h3.GoRunAsync("../../../cmd/wice-signal-http")
	if err != nil {
		t.Fail()
	}

	go SlicePrefix("Backend: ", stdout)
	go SlicePrefix("Backend: ", stderr)

	// w1, err := RunWice(h1)
	// if err != nil {
	// 	t.Fail()
	// }

	// w2, err := RunWice(h2)
	// if err != nil {
	// 	t.Fail()
	// }

	h3.Run("curl", "http://h3:8080/")

	time.Sleep(2 * time.Second)

	if err = Killall(b); err != nil {
		t.Fail()
	}
}
