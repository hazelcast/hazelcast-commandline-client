package metric

import (
	"encoding/json"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

func GenerateFirstPingQuery(gm GlobalMetrics, sm SessionMetrics, t time.Time) Query {
	return Query{
		Date:              t.Format(DateFormat),
		ID:                gm.ID,
		Architecture:      gm.Architecture,
		OS:                gm.OS,
		Version:           sm.CLCVersion,
		AcquisitionSource: string(sm.AcquisionSource),
	}
}

type MetricValues map[string]int

func GenerateQueries(db *store.Store, gm GlobalMetrics, dates map[string]struct{}) []Query {
	qs := make([]Query, 0, len(dates))
	for date := range dates {
		entries := make(map[types.Quadruple[string]]MetricValues)
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

			attributes := types.NewQuadruple(k.ClusterID, k.ViridianClusterID,
				string(k.AcquisitionSource), k.CLCVersion)
			if _, ok := entries[attributes]; !ok {
				entries[attributes] = make(MetricValues)
			}
			entries[attributes][k.MetricName] = v
			return true, nil
		})
		for attribs, metrics := range entries {
			cid := attribs.First
			vid := attribs.Second
			acqSource := attribs.Third
			version := attribs.Fourth
			d := Query{
				Date:                       date,
				ID:                         gm.ID,
				Architecture:               gm.Architecture,
				OS:                         gm.OS,
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
