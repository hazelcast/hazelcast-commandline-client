package metric

import (
	"encoding/json"
	"os"
	"runtime"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type GlobalMetrics struct {
	ID           string `json:"id"`
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

func NewGlobalMetrics() GlobalMetrics {
	return GlobalMetrics{
		ID:           types.NewUUID().String(),
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
	}
}

func (g *GlobalMetrics) Marshal() ([]byte, error) {
	return json.Marshal(g)
}

func (g *GlobalMetrics) Unmarshal(b []byte) error {
	gm := GlobalMetrics{}
	err := json.Unmarshal(b, &gm)
	if err != nil {
		return err
	}
	*g = gm
	return nil
}

type SessionMetrics struct {
	CLCVersion      string
	AcquisionSource AcquisionSource
}

func NewSessionMetrics() SessionMetrics {
	return SessionMetrics{
		CLCVersion:      internal.Version,
		AcquisionSource: FindAcquisionSource(),
	}
}

type AcquisionSource string

const (
	EnvAcquisionSource                     = "CLC_ACQUISION_SOURCE"
	AcquisionSourceMC      AcquisionSource = "MC"
	AcquisionSourceVSCode  AcquisionSource = "VSCode"
	AcquisionSourceDocker  AcquisionSource = "Docker"
	AcquisionSourceK8S     AcquisionSource = "Kubernetes"
	AcquisionSourceUnknown AcquisionSource = "Unknown"
)

func FindAcquisionSource() AcquisionSource {
	src := AcquisionSource(os.Getenv(EnvAcquisionSource))
	if src == "" {
		return AcquisionSourceUnknown
	}
	return AcquisionSource(src)
}

func (g *SessionMetrics) Marshal() ([]byte, error) {
	return json.Marshal(g)
}

func (g *SessionMetrics) Unmarshal(b []byte) error {
	gm := SessionMetrics{}
	err := json.Unmarshal(b, &gm)
	if err != nil {
		return err
	}
	*g = gm
	return nil
}
