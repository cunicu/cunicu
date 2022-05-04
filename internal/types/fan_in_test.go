package types_test

import (
	"testing"

	"riasc.eu/wice/internal/types"
)

func TestFanIn(t *testing.T) {
	N := 5

	ch_in := []chan int{}

	for i := 0; i < N; i++ {
		ch_in = append(ch_in, make(chan int))
	}

	ch_out := types.FanIn(ch_in...)

	for i := 0; i < N; i++ {
		ch_in[i] <- i
	}

	for i := 0; i < N; i++ {
		if <-ch_out != i {
			t.Fail()
		}
	}
}
