package metric

import (
	"errors"
	"strings"
	"time"
)

const (
	PhonehomeKeyPrefix = "ph"
	DateFormat         = "2006-01-02"
	KeyFieldSeparator  = "\\"
)

type Key struct {
	Datetime          time.Time
	ClusterID         string
	ViridianClusterID string
}

func NewKey(cid, vid string) Key {
	date := time.Now().UTC().Truncate(time.Hour * 24)
	return Key{
		Datetime:          date,
		ClusterID:         cid,
		ViridianClusterID: vid,
	}
}

func NewSimpleKey() Key {
	return NewKey("", "")
}

func datePrefix(date string) string {
	return PhonehomeKeyPrefix + KeyFieldSeparator + date
}

type storageKey struct {
	Key
	KeyPrefix         string
	AcquisitionSource AcquisionSource
	CLCVersion        string
	MetricName        string
}

func newStorageKey(k Key, as AcquisionSource, version string, metric string) storageKey {
	return storageKey{
		Key:               k,
		KeyPrefix:         PhonehomeKeyPrefix,
		AcquisitionSource: as,
		CLCVersion:        version,
		MetricName:        metric,
	}
}

func (c *storageKey) Marshal() ([]byte, error) {
	str := strings.Join([]string{
		c.KeyPrefix,
		c.Date(),
		c.ClusterID,
		c.ViridianClusterID,
		string(c.AcquisitionSource),
		c.CLCVersion,
		c.MetricName,
	}, KeyFieldSeparator)
	return []byte(str), nil
}

func (c *storageKey) Date() string {
	return c.Datetime.Format(DateFormat)
}

func (c *storageKey) Unmarshal(b []byte) error {
	s := strings.Split(string(b), KeyFieldSeparator)
	if len(s) != 7 {
		return errors.New("Key is in an incorrect format")
	}
	date, err := time.Parse(DateFormat, s[1])
	if err != nil {
		return err
	}
	*c = storageKey{
		KeyPrefix: s[0],
		Key: Key{
			Datetime:          date,
			ClusterID:         s[2],
			ViridianClusterID: s[3],
		},
		AcquisitionSource: AcquisionSource(s[4]),
		CLCVersion:        s[5],
		MetricName:        s[6],
	}
	return nil
}
