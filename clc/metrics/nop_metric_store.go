package metrics

type NopMetricStore struct{}

func (ms *NopMetricStore) Store(key Key, metric string, val int) {}
func (ms *NopMetricStore) Increment(key Key, metric string)      {}
