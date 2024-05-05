package activation

type options struct {
	unsetEnv bool
	method   Method
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

// Options takes some [option]s and produces a stuct containing the flags and settings.
func Options(opts ...option) *options {
	o := &options{
		unsetEnv: true,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
