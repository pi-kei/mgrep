package concurrency

import (
	"context"
	"sync"
)

// Takes multiple input channels and sends their values to a single output channel.
// Sets buffer size of an output channel to a specified value.
// When specified context is done it stops listening for input values and closes output channel.
// Returns an output channel.
func FanIn[I any](ins []chan I, bufferSize int, ctx context.Context) chan I {
	out := make(chan I, bufferSize)

	var wg sync.WaitGroup

	wg.Add(len(ins))
	for _, in := range ins {
		go func(ch <-chan I) {
			defer wg.Done()
			for {
				select {
				case value, ok := <-ch:
					if ok {
						select {
						case out <- value:
						case <-ctx.Done():
							return
						}
					} else {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}(in)
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}