package concurrency

import (
	"context"
	"reflect"
)

func SendToAny[I any](ctx context.Context, value I, channels []chan I) (int, error) {
	length := len(channels)
	cases := make([]reflect.SelectCase, length + 2)
	for i, ch := range channels {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: reflect.ValueOf(value),
		}
	}
	cases[length] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}
	cases[length + 1] = reflect.SelectCase{
		Dir:  reflect.SelectDefault,
	}
	chosen, _, _ := reflect.Select(cases)
	if chosen == length {
		return -1, ctx.Err()
	}
	if chosen == length + 1 {
		return -1, nil
	}
	return chosen, nil
}