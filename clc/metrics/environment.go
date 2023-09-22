package metrics

import (
	"encoding/json"
	"os"
	"runtime"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type GlobalAttributes struct {
	ID           string `json:"id"`
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

func NewGlobalAttributes() GlobalAttributes {
	return GlobalAttributes{
		ID:           types.NewUUID().String(),
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
	}
}

func (g *GlobalAttributes) Marshal() ([]byte, error) {
	return json.Marshal(g)
}

func (g *GlobalAttributes) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, g); err != nil {
		return err
	}
	return nil
}

type SessionAttributes struct {
	CLCVersion      string
	AcquisionSource AcquisionSource
}

func NewSessionMetrics() SessionAttributes {
	return SessionAttributes{
		CLCVersion:      internal.Version,
		AcquisionSource: FindAcquisionSource(),
	}
}

type AcquisionSource string

const (
	EnvAcquisionSource                     = "CLC_ACQUISITION_SOURCE"
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

func (g *SessionAttributes) Marshal() ([]byte, error) {
	return json.Marshal(g)
}

func (g *SessionAttributes) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, g); err != nil {
		return err
	}
	return nil
}
