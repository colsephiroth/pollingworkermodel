package client

import (
	"net/http"
	"time"
)

type Options struct {
	GetJobsTickRate         time.Duration
	JobChannelBufferSize    uint
	ResultChannelBufferSize uint
	HttpClient              *http.Client
}

type Option func(o *Options)

func NewOptions(options ...Option) *Options {
	config := &Options{
		GetJobsTickRate:         time.Millisecond * 100,
		JobChannelBufferSize:    100,
		ResultChannelBufferSize: 100,
		HttpClient:              http.DefaultClient,
	}

	for _, option := range options {
		option(config)
	}

	return config
}

func WithGetJobsTickRate(d time.Duration) Option {
	return func(o *Options) {
		o.GetJobsTickRate = d
	}
}

func WithJobChannelBufferSize(s uint) Option {
	return func(o *Options) {
		o.JobChannelBufferSize = s
	}
}

func WithResultChannelBufferSize(s uint) Option {
	return func(o *Options) {
		o.ResultChannelBufferSize = s
	}
}

func WithHttpClient(h *http.Client) Option {
	return func(o *Options) {
		o.HttpClient = h
	}
}
