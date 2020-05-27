// Package concurrent provides concurrent primitives/shortcuts.
package concurrent

import (
	"context"
	"sync"
)

// Group is a Close-able thread group.
type Group struct {
	group  sync.WaitGroup
	ctx    context.Context
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
		ctx:    ctx,
		cancel: cancel,
	}
}

// Go runs func in backround.
// Func should return when Group.Done() channel is closed.
func (g *Group) Go(f func()) {
	g.group.Add(1)
	go func() { f(); g.group.Done() }()
}

// Done returns a done channel for this thread group.
// It should be used by goroutines to return when it's closed.
func (g *Group) Done() <-chan struct{} {
	return g.ctx.Done()
}

// Close cancels context and waits for threads to terminate.
func (g *Group) Close() error {
	g.cancel()
	g.group.Wait()
	return nil
}
