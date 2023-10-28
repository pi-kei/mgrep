package concurrency

import "context"

// Takes a single input channel and sends its values to multiple output channels.
// Sets buffer size of output channels to a specified value.
// When specified context is done it stops listening for input values and closes output channels.
// Returns output channels.
func FanOut[I any](in chan I, n int, bufferSize int, ctx context.Context) []chan I {
	outs := make([]chan I, n)
	for i := 0; i < n; i++ {
		outs[i] = make(chan I, bufferSize)

		go func(out chan I) {
			defer close(out)
			for {
				select {
				case inValue, ok := <-in:
					if ok {
						select {
						case out <- inValue:
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
		}(outs[i])
	}

	return outs
}