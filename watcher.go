package filewatcher

import (
	"context"
	"errors"
	"os"
	"time"
)

type pollingWatcher struct {
	childContext context.Context
	shutdown     context.CancelFunc
	filename     string
	interval     time.Duration
	notify       func()
	known        time.Time
}

func newSimpleWatcher(parent context.Context, filename string, interval time.Duration, notify func()) ListenCloser {
	if len(filename) == 0 || interval < 0 {
		return nop{}
	}

	child, shutdown := context.WithCancel(parent)
	return &pollingWatcher{
		childContext: child,
		shutdown:     shutdown,
		filename:     filename,
		interval:     interval,
		notify:       notify,
	}
}

func (this *pollingWatcher) Listen() {
	this.known = this.lastModified()
	for this.sleep() {
		if lastModified := this.lastModified(); lastModified.After(this.known) {
			this.known = lastModified
			this.notify()
		}
	}
}
func (this *pollingWatcher) lastModified() time.Time {
	// simple: using last-modification timestamp instead of file hash
	stat, _ := os.Stat(this.filename)
	return stat.ModTime()
}
func (this *pollingWatcher) sleep() bool {
	// simple: polling every interval instead of watching for filesystem changes
	ctx, cancel := context.WithTimeout(this.childContext, this.interval)
	defer cancel()
	<-ctx.Done()
	return errors.Is(ctx.Err(), context.Canceled)
}

func (this *pollingWatcher) Close() error {
	this.shutdown()
	return nil
}
