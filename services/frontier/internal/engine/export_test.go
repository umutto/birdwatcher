package engine

import "context"

// DispatchNext exports the internal dispatchNext method for testing.
func (e *Engine) DispatchNext(ctx context.Context) (bool, error) {
	return e.dispatchNext(ctx)
}

// QueueLen returns the current number of items in the engine's queue for testing.
func (e *Engine) QueueLen() int {
	return e.queue.Len()
}
