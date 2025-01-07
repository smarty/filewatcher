package filewatcher

import (
	"context"
	"strings"
	"time"
)

func New(options ...option) ListenCloser {
	var config configuration
	Options.apply(options...)(&config)
	return newSimpleWatcher(config)
}

func (singleton) Context(value context.Context) option {
	return func(this *configuration) { this.context = value }
}
func (singleton) Filenames(values ...string) option {
	return func(this *configuration) {
		for i := range values {
			values[i] = strings.TrimSpace(values[i])
		}
	}
}
func (singleton) Interval(value time.Duration) option {
	return func(this *configuration) { this.interval = value }
}
func (singleton) Notify(value func()) option {
	return func(this *configuration) { this.notify = value }
}

func (singleton) Logger(value logger) option {
	return func(this *configuration) { this.logger = value }
}

func (singleton) apply(options ...option) option {
	return func(this *configuration) {
		for _, item := range Options.defaults(options...) {
			item(this)
		}
	}
}
func (singleton) defaults(options ...option) []option {
	return append([]option{
		Options.Context(context.Background()),
		Options.Filenames(),
		Options.Interval(time.Hour),
		Options.Notify(func() {}),
		Options.Logger(nop{}),
	}, options...)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type configuration struct {
	context   context.Context
	filenames []string
	interval  time.Duration
	notify    func()
	logger    logger
}

type option func(*configuration)
type singleton struct{}

var Options singleton

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type nop struct{}

func (nop) Printf(string, ...any) {}

func (nop) Initialize() error { return nil }
func (nop) Listen()           {}
func (nop) Close() error      { return nil }
