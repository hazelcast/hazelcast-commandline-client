package serialization

import (
	"fmt"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

const snapshotClassID = 32
const snapshotFactoryID = -10002

type Snapshot struct {
	ID            int64
	NumChunks     int64
	NumBytes      int64
	CreationTime  int64
	JobID         int64
	JobName       string
	DagJsonString string
}

func (s *Snapshot) FactoryID() int32 {
	return snapshotFactoryID
}

func (s *Snapshot) ClassID() int32 {
	return snapshotClassID
}

type SnapshotFactory struct{}

func (SnapshotFactory) Create(classID int32) serialization.IdentifiedDataSerializable {
	if classID == snapshotClassID {
		return &Snapshot{}
	}
	panic(fmt.Errorf("classID is not correct, it must be %d", snapshotClassID))
}

func (SnapshotFactory) FactoryID() int32 {
	return snapshotFactoryID
}

func (s *Snapshot) WriteData(output serialization.DataOutput) {
	// not used
}

func (s *Snapshot) ReadData(input serialization.DataInput) {
	s.ID = input.ReadInt64()
	s.NumChunks = input.ReadInt64()
	s.NumBytes = input.ReadInt64()
	s.CreationTime = input.ReadInt64()
	s.JobID = input.ReadInt64()
	s.JobName = input.ReadString()
	s.DagJsonString = input.ReadString()
}
