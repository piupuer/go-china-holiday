package holiday

type Options struct {
	filename string
}

func WithFilename(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).filename = s
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			filename: defaultFilename,
		}
	}
	return options
}
