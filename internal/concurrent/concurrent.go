// Package concurrent provides concurrent primitives/shortcuts.
package concurrent

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Group is a Close-able thread group.
type Group struct {
	*errgroup.Group
	context.Context

	cancel context.CancelFunc
}

// NewGroup initializes a new Close-able thread group.
//
// Usage:
//
//   threads := concurrent.NewGroup(ctx)
//   threads.Go(func() error {
//     <-threads.Done() // "subscribe" for cancellation
//     return nil
//   })
//   err := threads.Close() // may be defer-ed etc - blocks till all threads terminate
//
func NewGroup(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)
	group, ctx := errgroup.WithContext(ctx)
	return &Group{
		Group:   group,
		Context: ctx,
		cancel:  cancel,
	}
}

// Close cancels context and waits for threads to terminate.
func (g *Group) Close() error {
	g.cancel()
	return g.Wait()
}
