package metrics

import (
	"encoding/json"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
)

func GenerateFirstPingQuery(ga GlobalAttributes, sa SessionAttributes, t time.Time) Query {
	return Query{
		Date:              t.Format(DateFormat),
		ID:                ga.ID,
		Architecture:      ga.Architecture,
		OS:                ga.OS,
		Version:           sa.CLCVersion,
		AcquisitionSource: string(sa.AcquisionSource),
	}
}

type MetricValues map[string]int

func GenerateQueries(db *store.Store, ga GlobalAttributes, dates map[string]struct{}) []Query {
	qs := make([]Query, 0, len(dates))
	for date := range dates {
		entries := make(map[[4]string]MetricValues)
		prefix := datePrefix(date)
		db.RunForeachWithPrefix(prefix, func(keyb, valb []byte) (bool, error) {
			var k storageKey
			if err := k.Unmarshal(keyb); err != nil {
				return false, err
			}
			var v int
			if err := json.Unmarshal(valb, &v); err != nil {
				return false, err
			}
			attributes := [4]string{k.ClusterID, k.ViridianClusterID,
				string(k.AcquisitionSource), k.CLCVersion}
			if _, ok := entries[attributes]; !ok {
				entries[attributes] = make(MetricValues)
			}
			entries[attributes][k.MetricName] = v
			return true, nil
		})
		for attribs, metrics := range entries {
			cid := attribs[0]
			vid := attribs[1]
			acqSource := attribs[2]
			version := attribs[3]
			d := Query{
				Date:                       date,
				ID:                         ga.ID,
				Architecture:               ga.Architecture,
				OS:                         ga.OS,
				Version:                    version,
				AcquisitionSource:          acqSource,
				ClusterUUID:                cid,
				ViridianClusterID:          vid,
				ClusterConfigCount:         metrics["cluster-config-count"],
				SqlRunCount:                metrics["sql"],
				MapRunCount:                metrics["map"],
				TopicRunCount:              metrics["topic"],
				QueueRunCount:              metrics["queue"],
				MultiMapRunCont:            metrics["multimap"],
				ListRunCount:               metrics["list"],
				DemoRunCount:               metrics["demo"],
				ProjectRunCount:            metrics["project"],
				JobRunCount:                metrics["job"],
				ViridianRunCount:           metrics["viridian"],
				SetRunCount:                metrics["set"],
				ShellRunCount:              metrics["shell"],
				ScriptRunCount:             metrics["script"],
				AtomicLongRunCount:         metrics["atomiclong"],
				TotalRunCount:              metrics["total"],
				InteractiveModeRunCount:    metrics["interactive-mode"],
				NoninteractiveModeRunCount: metrics["noninteractive-mode"],
				ScriptModeRunCount:         metrics["script-mode"],
			}
			qs = append(qs, d)
		}
	}
	return qs
}
