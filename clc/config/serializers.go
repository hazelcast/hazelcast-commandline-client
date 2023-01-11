package config

import (
	"fmt"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

// TODO: move this file to somewhere more appropriate

const (
	jetFactoryID              = -10002
	jetGetJobIdsResultClassID = 19
	jetJobSummaryClassID      = 29
)

type GetJobIdsResult struct {
	IDs   []int64
	Light []bool
}

func (s *GetJobIdsResult) FactoryID() int32 {
	return jetFactoryID
}

func (s *GetJobIdsResult) ClassID() int32 {
	return jetGetJobIdsResultClassID
}

func (s *GetJobIdsResult) WriteData(out serialization.DataOutput) {
	panic("not supported")
}

func (s *GetJobIdsResult) ReadData(in serialization.DataInput) {
	s.IDs = in.ReadInt64Array()
	s.Light = in.ReadBoolArray()
}

type JetJobSummary struct {
	Light          bool
	ID             int64
	ExecutionID    int64
	NameOrID       string
	Status         interface{}
	SubmissionTime int64
	CompletionTime int64
	FailureText    string
}

func (s *JetJobSummary) FactoryID() int32 {
	return jetFactoryID
}

func (s *JetJobSummary) ClassID() int32 {
	return jetJobSummaryClassID
}

func (s *JetJobSummary) WriteData(out serialization.DataOutput) {
	panic("not supported")
}

func (s *JetJobSummary) ReadData(in serialization.DataInput) {
	s.Light = in.ReadBool()
	s.ID = in.ReadInt64()
	s.ExecutionID = in.ReadInt64()
	s.NameOrID = in.ReadString()
	s.Status = in.ReadObject()
	s.SubmissionTime = in.ReadInt64()
	s.CompletionTime = in.ReadInt64()
	s.FailureText = in.ReadString()
}

type JetIdentifiedDataSerializableFactory struct{}

func (f JetIdentifiedDataSerializableFactory) Create(id int32) serialization.IdentifiedDataSerializable {
	switch id {
	case jetGetJobIdsResultClassID:
		return &GetJobIdsResult{}
	case jetJobSummaryClassID:
		return &JetJobSummary{}
	}
	panic(fmt.Errorf("unknown class ID for Jet factory: %d", id))
}

func (f JetIdentifiedDataSerializableFactory) FactoryID() int32 {
	return jetFactoryID
}
