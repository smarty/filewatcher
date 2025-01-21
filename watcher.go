package filewatcher

import (
	"context"
	"errors"
	"os"
	"strings"
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
	filenames := make([]string, 0, len(config.filenames))
	for _, filename := range config.filenames {
		if filename = strings.TrimSpace(filename); len(filename) > 0 {
			filenames = append(filenames, filename)
		}
	}

	if len(config.filenames) == 0 || config.interval < 0 {
		return nop{}
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
	for i := range this.filenames {
		this.known[i] = this.lastModified(this.filenames[i])
	}

	for this.sleep() {
		if this.update() {
			this.notify()
		}
	}
}
func (this *pollingWatcher) update() bool {
	count := 0

	for index := range this.filenames {
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
		this.known[index] = lastModified
		this.logger.Printf("[INFO] Watched file [%s] was modified on [%s].", this.filenames[index], lastModified.String())
	}

	if count == 0 {
		return false
	}

	this.logger.Printf("[INFO] [%d] watched files modified, notifying subscribers...", count)
	return true
}
func (this *pollingWatcher) lastModified(filename string) time.Time {
	// simple: using last-modification timestamp instead of file hash
	if stat, err := os.Stat(filename); err == nil {
		return stat.ModTime()
	}

	return time.Time{}
}
func (this *pollingWatcher) sleep() bool {
	// simple: polling every interval instead of watching for filesystem changes
	ctx, cancel := context.WithTimeout(this.childContext, this.interval)
	defer cancel()
	<-ctx.Done()
	return errors.Is(ctx.Err(), context.DeadlineExceeded)
}

func (this *pollingWatcher) Close() error {
	this.shutdown()
	return nil
}
