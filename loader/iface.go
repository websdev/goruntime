package loader

import "github.com/lyft/goruntime/snapshot"

type IFace interface {
	// @return Snapshot the current snapshot. This reference is safe to use forever, but will grow
	//         stale so should not be stored beyond when it is immediately needed.
	Snapshot() snapshot.IFace

	// Add a channel that will be written to when a new snapshot is available. "1" will be written
	// to the channel as a sentinal.
	// @param callback supplies the callback to add.
	AddUpdateCallback(callback chan<- int)
}

// Implementation of Loader with no backing store.
type Nil struct {
	snapshot snapshot.IFace
}

func (n *Nil) Snapshot() snapshot.IFace { return n.snapshot }

func (n *Nil) AddUpdateCallback(callback chan<- int) {}
