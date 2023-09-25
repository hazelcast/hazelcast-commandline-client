package metrics

import (
	"encoding/json"
	"os"
	"runtime"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-go-client/types"
)

func PhoneHomeEnabled() bool {
	return os.Getenv(EnvPhoneHomeEnabled) != "false"
}

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
	CLCVersion        string
	AcquisitionSource AcquisitionSource
}

func NewSessionMetrics() SessionAttributes {
	return SessionAttributes{
		CLCVersion:        internal.Version,
		AcquisitionSource: FindAcquisionSource(),
	}
}

type AcquisitionSource string

const (
	EnvAcquisionSource                       = "CLC_INTERNAL_ACQUISITION_SOURCE"
	AcquisionSourceMC      AcquisitionSource = "MC"
	AcquisionSourceVSCode  AcquisitionSource = "VSCode"
	AcquisionSourceDocker  AcquisitionSource = "Docker"
	AcquisionSourceK8S     AcquisitionSource = "Kubernetes"
	AcquisionSourceUnknown AcquisitionSource = "Unknown"
)

func FindAcquisionSource() AcquisitionSource {
	src := AcquisitionSource(os.Getenv(EnvAcquisionSource))
	if src == "" {
		return AcquisionSourceUnknown
	}
	return AcquisitionSource(src)
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
