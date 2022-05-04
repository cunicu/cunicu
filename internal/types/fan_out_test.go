package types_test

import (
	"sync"
	"testing"

	"riasc.eu/wice/internal/types"
)

func TestFanOut(t *testing.T) {
	N := 5
	ch_out := []chan int{}

	fo := types.NewFanout[int]()

	for i := 0; i < N; i++ {
		ch_out = append(ch_out, fo.AddChannel())
	}

	fo.C <- 1234

	wg := sync.WaitGroup{}
	wg.Add(N)

	for i := 0; i < N; i++ {
		go func(i int) {
			if <-ch_out[i] != 1234 {
				t.Fail()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
