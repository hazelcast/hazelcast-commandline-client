package metrics

import (
	"context"
)

type MetricStorer interface {
	// Store stores the value in the specified key/metric
	// metric is in the form of "metric1.metric2.metric3".
	Store(key Key, metric string, val int)
	// Increment increments the value in the specified key/metric by one.
	// metric is in the form of "metric1.metric2.metric3"
	Increment(key Key, metric string)
}

type metricSender interface {
	Send(ctx context.Context) error
}
