package api

type NewAPIFunc func(opts ...Option) API

type Options struct {
	URL         string
	Token 			string
	APIInstance NewAPIFunc
}

type Option func(opts *Options)

func NewOptions(opts ...Option) Options {
	options := NewOptionsDefault()

	for _, o := range opts {
		o(&options)
	}

	return options
}

func NewOptionsDefault() Options {
	return Options{
		URL:         defaultOptions.URL,
		Token:       defaultOptions.Token,
		APIInstance: defaultOptions.APIInstance,
	}
}

var defaultOptions = Options{
	URL:         "",
	Token:		 "",
	APIInstance: nil,
}

func Url(url string) Option {
	return func(options *Options) {
		options.URL = url
	}
}

func Token(t string) Option {
	return func(option *Options) {
		option.Token = t
	}
}

func APIInstance(function NewAPIFunc) Option {
	return func(options *Options) {
		options.APIInstance = function
	}
}
