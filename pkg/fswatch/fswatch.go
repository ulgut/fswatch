package fswatch

import (
	"context"
	"fmt"
	"sync"
	"syscall"
)

// todo: introduce a timeout

const (
	eventQueueSize = 1
	runningStatus  = 0x0
	exitedStatus   = 0x1
)

type Watcher struct {
	ctx context.Context

	paths map[string]*entry
	mut   sync.Mutex

	stat uint8
}

/*
Returns a new watcher instance.
*/
func New() *Watcher {
	ctx := context.Background()
	w := &Watcher{paths: make(map[string]*entry), mut: sync.Mutex{}, stat: runningStatus, ctx: ctx}

	return w
}

/*
Adds a path to the watcher if it doesn't already exists. Gaurantees no ordering as to when events become available.
*/
func (w *Watcher) Watch(path string, fn Callback) (err error) {
	w.mut.Lock()
	defer w.mut.Unlock()

	if w.stat != runningStatus {
		return fmt.Errorf("watcher is not available")
	}

	if _, ok := w.paths[path]; ok {
		return fmt.Errorf("already watching: '%s'", path)
	}

	fd, err := syscall.Open(path, syscall.O_RDONLY, 0) // todo: move lock acq outside of syscall path
	if err != nil {
		return fmt.Errorf("failed to open '%s': %v", path, err)
	}

	kq, err := syscall.Kqueue() // todo: handwrap a single kqueue
	if err != nil {
		return fmt.Errorf("failed to acquire a kq fd for '%s': %v", path, err)
	}

	ctx, cancel := context.WithCancel(w.ctx)
	entry := &entry{
		fd:       fd,
		kq:       kq,
		ctx:      ctx,
		cancel:   cancel,
		shutdown: make(chan struct{}, 1),
		events:   make(chan Event, eventQueueSize),
		callback: fn,
	}

	w.paths[path] = entry

	go entry.eventProducer(path)
	go entry.eventConsumer(path)

	return
}

/*
Removes a path from the watcher if it exists. Blocks.
*/
func (w *Watcher) Remove(path string) error {
	w.mut.Lock()
	defer w.mut.Unlock()

	entry, ok := w.paths[path]
	if !ok {
		return fmt.Errorf("not watching path: '%s'", path)
	}
	delete(w.paths, path)

	entry.cancel()
	<-entry.shutdown

	return nil
}

func (w *Watcher) handleShutdown() {
	w.mut.Lock()
	defer w.mut.Unlock()

	fmt.Println("here1")

	for _, entry := range w.paths {
		entry.cancel()
		fmt.Println("here2")
		<-entry.shutdown
	}

	w.stat = exitedStatus
}

/*
Blocks until related path watchers are shutdown.
*/
func (w *Watcher) Shutdown() {
	w.handleShutdown()
}
