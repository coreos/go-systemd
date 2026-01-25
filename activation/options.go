package activation

type options struct {
	unsetEnv bool
	method   Method
}

// apply applies option functions to the options struct.
func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}

type option func(*options)

// UnsetEnv controls if the LISTEN_PID, LISTEN_FDS & LISTEN_FDNAMES environment
// variables are unset.
//
// This is useful to avoid clashes in fd usage and to avoid leaking environment
// flags to child processes.
func UnsetEnv(f bool) option {
	return func(o *options) {
		o.unsetEnv = f
	}
}

// UseMethod chooses the [Method] applied to the file descriptor passed in by
// systemd socket activation.
func UseMethod(m Method) option {
	return func(o *options) {
		o.method = m
	}
}
