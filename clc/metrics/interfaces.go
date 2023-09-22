package metrics

import (
	"context"
)

type MetricStoreSender interface {
	MetricStorer
	MetricSender
}

type MetricStorer interface {
	// metric is in the form of "metric1.metric2.metric3"
	Store(key Key, metric string, val int)
	// metric is in the form of "metric1.metric2.metric3"
	Increment(key Key, metric string)
}

type MetricSender interface {
	Send(ctx context.Context) error
}
