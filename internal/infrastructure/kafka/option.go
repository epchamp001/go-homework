package kafka

import (
	"io"

	"github.com/twmb/franz-go/pkg/kgo"
)

type ClientOption func([]kgo.Opt) []kgo.Opt

func WithBasicLogger(dst io.Writer, lvl kgo.LogLevel, prefix string) ClientOption {
	return func(o []kgo.Opt) []kgo.Opt {
		l := kgo.BasicLogger(dst, lvl, func() string { return prefix })
		return append(o, kgo.WithLogger(l))
	}
}

type RecordOption func(*kgo.Record)

func WithHeaders(h ...kgo.RecordHeader) RecordOption {
	return func(r *kgo.Record) {
		r.Headers = append(r.Headers, h...)
	}
}
