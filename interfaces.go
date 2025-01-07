package filewatcher

import "io"

type ListenCloser interface {
	Initialize() error
	Listen()
	io.Closer
}

type logger interface {
	Printf(string, ...any)
}
