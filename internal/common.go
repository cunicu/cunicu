package internal

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

func SetupRand() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func SetupSignals() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, unix.SIGINT, unix.SIGTERM)

	return ch
}

func SetupPeriodicHeapDumps() {
	logger := zap.L().Named("pprof")

	hn, _ := os.Hostname()
	s := strings.Split(hn, ".")
	prefix := s[0]

	go func() {
		for i := 0; ; i++ {
			fn := fmt.Sprintf("%s_heap.dump", prefix)

			wr, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				logger.Error("Failed to open file for heap profile", zap.Error(err))
				continue
			}

			if err := pprof.WriteHeapProfile(wr); err != nil {
				logger.Error("Failed to write heap profile", zap.Error(err))
			}

			logger.Debug("Wrote heap dump", zap.String("file", fn))

			time.Sleep(1 * time.Second)
		}
	}()
}
