package concurrency

import (
	"context"
)

// Takes an input channel then calls specified proc function on values from input channel and sends values from proc function to a created output channel.
// Sets buffer size of an output channel to specified value.
// When specified context is done it stops listening for input values and closes output channel, or it returns context's error to proc function.
// When proc function is done it stops listening for input values and closes output channel.
// Returns an output channel.
func Pipeline[I any, O any](in chan I, proc func(I, func(O) error), bufferSize int, ctx context.Context) chan O {
	out := make(chan O, bufferSize)

	go func() {
		defer close(out)
		for {
			select {
			case inValue, ok := <-in:
				if ok {
					proc(inValue, func(value O) error {
						select {
						case out <- value:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					})
				} else {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

// Calls Pipeline function on all specified input channels and returns all output channels.
func PipelineMulti[I any, O any](ins []chan I, proc func(I, func(O) error), bufferSize int, ctx context.Context) []chan O {
	outs := make([]chan O, len(ins))
	
	for i := 0; i < len(ins); i++ {
		outs[i] = Pipeline(ins[i], proc, bufferSize, ctx)
	}

	return outs
}