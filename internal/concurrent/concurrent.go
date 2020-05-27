// Package concurrent provides concurrent primitives/shortcuts.
package concurrent

import (
	"context"
	"sync"
)

// Group is a Close-able thread group.
type Group struct {
	context.Context

	group  sync.WaitGroup
	cancel context.CancelFunc
}

// NewGroup initializes a new Close-able thread group.
//
// Usage:
//
//   threads := concurrent.NewGroup(ctx)
//   threads.Go(func() {
//     <-threads.Done() // "subscribe" for cancellation
//     ...
//   })
//   err := threads.Close() // may be defer-ed etc - blocks till all threads terminate
//
func NewGroup(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{
		Context: ctx,
		cancel:  cancel,
	}
}

// Go runs func in backround.
// Func should return when Group.Context is cancelled/done.
func (g *Group) Go(f func()) {
	g.group.Add(1)
	go func() { f(); g.group.Done() }()
}

// Close cancels context and waits for threads to terminate.
func (g *Group) Close() error {
	g.cancel()
	g.group.Wait()
	return nil
}
