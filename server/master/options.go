package master

type Options struct {
	addr string
}

type MasterOption func(*Options)

// WithAddr
func WithAddr(addr string) MasterOption {
	return func(options *Options) {
		options.addr = addr
	}
}
