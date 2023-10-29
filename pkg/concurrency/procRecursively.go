package concurrency

import (
	"context"
	"sync"
)

// Process values recursively on specified number of concurrent routines.
// Processing a single value can produce more values of the same type to process.
// Specified proc function receives another function to send new values back to processing.
// This send function returns index of chosen channel if value was sent or -1 and an error if occured.
// If value was not sent to avoid deadlocks and waits proc function should continue processing in the same routine.
// Returns output channels.
func ProcRecursively[I any, O any](root I, proc func(I, func(I) (int, error), func(O) error), n int, bufferSize int, ctx context.Context) []chan O {
	var wg sync.WaitGroup
	ins := make([]chan I, n)

	for i := 0; i < n; i++ {
		ins[i] = make(chan I, bufferSize)
	}

	send := func(value I) (int, error) {
		chosen, err := SendToAny(value, ins, ctx)
		if chosen >= 0 {
			wg.Add(1)
		}
		return chosen, err
	}

	outs := PipelineMulti(ins, func(child I, innerProc func(O) error) {
		defer wg.Done()
		proc(child, send, innerProc)
	}, bufferSize, ctx)

	send(root)

	go func() {
		defer func() {
			for _, ch := range ins {
				close(ch)
			}
		}()
		wg.Wait()
	}()

	return outs
}