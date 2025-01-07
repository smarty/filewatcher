package filewatcher

import "io"

type ListenCloser interface {
	Listen()
	io.Closer
}

type logger interface {
	Printf(string, ...any)
}
