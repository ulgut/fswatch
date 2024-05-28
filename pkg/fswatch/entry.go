package fswatch

import (
	"context"
	"fmt"
	"syscall"
)

const (
	queueWatchMode  = syscall.EVFILT_VNODE
	queueFlagBitmap = syscall.EV_ADD | syscall.EV_ENABLE | syscall.EV_ONESHOT
	watchFlagBitmap = syscall.NOTE_DELETE | syscall.NOTE_WRITE | syscall.NOTE_EXTEND | syscall.NOTE_RENAME
)

type Callback func(Event)

type entry struct {
	ctx      context.Context
	cancel   context.CancelFunc
	shutdown chan struct{}

	fd int // path fd
	kq int // kqueue fd

	events   chan Event
	callback Callback
}

func (e *entry) eventConsumer(path string) {
	fmt.Printf("monitoring: '%s'\n", path)

	for {
		select {
		case <-e.ctx.Done():
			fmt.Printf("shutting consumer wrkr for path: '%s'\n", path)
			return
		case event := <-e.events: // todo: update this
			e.callback(event)
		}
	}
}

func (e *entry) eventProducer(path string) {
	defer syscall.Close(e.fd)
	defer syscall.Close(e.kq)
	defer close(e.shutdown)

	events := make([]syscall.Kevent_t, eventQueueSize)
	config := []syscall.Kevent_t{{
		Ident:  uint64(e.fd),
		Filter: queueWatchMode,
		Flags:  queueFlagBitmap,
		Fflags: watchFlagBitmap,
		Data:   0,
		Udata:  nil,
	}}

	for {
		select {
		case <-e.ctx.Done():
			fmt.Printf("shutting producer wrkr for path: '%s'\n", path)
			return
		default:
			n, err := syscall.Kevent(e.kq, config, events, nil)
			if err != nil { // todo: make more robust
				continue
			}
			for i := 0; i < n; i++ {
				if events[i].Flags&syscall.EV_ERROR != 0 {
					panic(fmt.Sprintf("error: %v", syscall.Errno(events[i].Data)))
				}
				fmt.Println("reokrpofkerofk?")
				e.events <- buildEvent(&events[i], path)
			}
		}
	}
}

func buildEvent(kEvt *syscall.Kevent_t, path string) Event {
	fsEvent := Event{}
	fsEvent.Path = path
	switch {
	case kEvt.Fflags&syscall.NOTE_DELETE != 0:
		fsEvent.Type = DeleteFile
	case kEvt.Fflags&syscall.NOTE_WRITE != 0:
		fsEvent.Type = WriteFile
	case kEvt.Fflags&syscall.NOTE_EXTEND != 0:
		fsEvent.Type = WriteFile // todo: make more robust
	case kEvt.Fflags&syscall.NOTE_RENAME != 0:
		fsEvent.Type = WriteFile // todo: make more robust
	}
	return fsEvent
}
