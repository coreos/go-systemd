package activation

type options struct {
	unsetEnv bool
	method   Method
}

type option func(*options)

func UnsetEnv(f bool) option {
	return func(o *options) {
		o.unsetEnv = f
	}
}

func UseMethod(m Method) option {
	return func(o *options) {
		o.method = m
	}
}

func Options(opts ...option) *options {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
