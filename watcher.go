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
	filenames    []string
	interval     time.Duration
	notify       func()
	logger       logger
	known        []time.Time
}

func newSimpleWatcher(config configuration) ListenCloser {
	if len(config.filenames) == 0 || config.interval < 0 {
		return nop{}
	}

	filenames := make([]string, 0, len(config.filenames))
	for _, filename := range config.filenames {
		filenames = append(filenames, filename)
	}

	child, shutdown := context.WithCancel(config.context)
	return &pollingWatcher{
		childContext: child,
		shutdown:     shutdown,
		filenames:    filenames,
		interval:     config.interval,
		notify:       config.notify,
		logger:       config.logger,
		known:        make([]time.Time, len(filenames)),
	}
}

func (this *pollingWatcher) Listen() {
	this.update()

	for this.sleep() {
		if this.update() {
			this.notify()
		}
	}
}
func (this *pollingWatcher) update() (updated bool) {
	count := 0

	for index, _ := range this.filenames {
		lastModified := this.lastModified(this.filenames[index])
		if lastModified.IsZero() {
			continue // no modification time
		}

		if this.known[index].IsZero() {
			continue // no last modified time; don't notify subscribers
		}

		if !lastModified.After(this.known[index]) {
			continue // file hasn't been modified since last check
		}

		count++
		updated = true
		this.known[index] = lastModified
		this.logger.Printf("[INFO] Watched file [%s] was modified on [%s].", this.filenames[index], lastModified.String())
	}

	this.logger.Printf("[INFO] [%d] watched files modified, notifying subscribers...", count)
	return updated
}
func (this *pollingWatcher) lastModified(filename string) time.Time {
	// simple: using last-modification timestamp instead of file hash
	stat, _ := os.Stat(filename)
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

func (this *pollingWatcher) Initialize() error { return nil }
