package metrics

import (
	"context"
)

type NopMetricStore struct{}

func (ms *NopMetricStore) Store(key Key, metric string, val int) {}
func (ms *NopMetricStore) Increment(key Key, metric string)      {}
func (ms *NopMetricStore) Send(ctx context.Context) error        { return nil }
