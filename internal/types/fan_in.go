package types

func FanIn[T any](chans ...chan T) chan T {
	nch := make(chan T)

	for _, ch := range chans {
		go func(ch chan T) {
			for m := range ch {
				nch <- m
			}
		}(ch)
	}

	return nch
}
