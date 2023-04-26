/*
* Copyright (c) 2008-2023, Hazelcast, Inc. All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License")
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package codec

import (
	proto "github.com/hazelcast/hazelcast-go-client"

	pubcontrol "github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
)

const (
	JobAndSqlSummaryCodecLightJobFieldOffset           = 0
	JobAndSqlSummaryCodecJobIdFieldOffset              = JobAndSqlSummaryCodecLightJobFieldOffset + proto.BooleanSizeInBytes
	JobAndSqlSummaryCodecExecutionIdFieldOffset        = JobAndSqlSummaryCodecJobIdFieldOffset + proto.LongSizeInBytes
	JobAndSqlSummaryCodecStatusFieldOffset             = JobAndSqlSummaryCodecExecutionIdFieldOffset + proto.LongSizeInBytes
	JobAndSqlSummaryCodecSubmissionTimeFieldOffset     = JobAndSqlSummaryCodecStatusFieldOffset + proto.IntSizeInBytes
	JobAndSqlSummaryCodecCompletionTimeFieldOffset     = JobAndSqlSummaryCodecSubmissionTimeFieldOffset + proto.LongSizeInBytes
	JobAndSqlSummaryCodecUserCancelledFieldOffset      = JobAndSqlSummaryCodecCompletionTimeFieldOffset + proto.LongSizeInBytes
	JobAndSqlSummaryCodecUserCancelledInitialFrameSize = JobAndSqlSummaryCodecUserCancelledFieldOffset + proto.BooleanSizeInBytes
)

func EncodeJobAndSqlSummary(clientMessage *proto.ClientMessage, jobAndSqlSummary pubcontrol.JobAndSqlSummary) {
	clientMessage.AddFrame(proto.BeginFrame.Copy())
	initialFrame := proto.NewFrame(make([]byte, JobAndSqlSummaryCodecUserCancelledInitialFrameSize))
	EncodeBoolean(initialFrame.Content, JobAndSqlSummaryCodecLightJobFieldOffset, jobAndSqlSummary.LightJob)
	EncodeLong(initialFrame.Content, JobAndSqlSummaryCodecJobIdFieldOffset, int64(jobAndSqlSummary.JobId))
	EncodeLong(initialFrame.Content, JobAndSqlSummaryCodecExecutionIdFieldOffset, int64(jobAndSqlSummary.ExecutionId))
	EncodeInt(initialFrame.Content, JobAndSqlSummaryCodecStatusFieldOffset, int32(jobAndSqlSummary.Status))
	EncodeLong(initialFrame.Content, JobAndSqlSummaryCodecSubmissionTimeFieldOffset, int64(jobAndSqlSummary.SubmissionTime))
	EncodeLong(initialFrame.Content, JobAndSqlSummaryCodecCompletionTimeFieldOffset, int64(jobAndSqlSummary.CompletionTime))
	EncodeBoolean(initialFrame.Content, JobAndSqlSummaryCodecUserCancelledFieldOffset, jobAndSqlSummary.UserCancelled)
	clientMessage.AddFrame(initialFrame)

	EncodeString(clientMessage, jobAndSqlSummary.NameOrId)
	EncodeNullableForString(clientMessage, jobAndSqlSummary.FailureText)
	EncodeNullableForSqlSummary(clientMessage, &jobAndSqlSummary.SqlSummary)
	EncodeNullableForString(clientMessage, jobAndSqlSummary.SuspensionCause)

	clientMessage.AddFrame(proto.EndFrame.Copy())
}

func DecodeJobAndSqlSummary(frameIterator *proto.ForwardFrameIterator) pubcontrol.JobAndSqlSummary {
	// begin frame
	frameIterator.Next()
	initialFrame := frameIterator.Next()
	lightJob := DecodeBoolean(initialFrame.Content, JobAndSqlSummaryCodecLightJobFieldOffset)
	jobId := DecodeLong(initialFrame.Content, JobAndSqlSummaryCodecJobIdFieldOffset)
	executionId := DecodeLong(initialFrame.Content, JobAndSqlSummaryCodecExecutionIdFieldOffset)
	status := DecodeInt(initialFrame.Content, JobAndSqlSummaryCodecStatusFieldOffset)
	submissionTime := DecodeLong(initialFrame.Content, JobAndSqlSummaryCodecSubmissionTimeFieldOffset)
	completionTime := DecodeLong(initialFrame.Content, JobAndSqlSummaryCodecCompletionTimeFieldOffset)
	var userCancelled bool
	if len(initialFrame.Content) >= JobAndSqlSummaryCodecUserCancelledFieldOffset+proto.BooleanSizeInBytes {
		userCancelled = DecodeBoolean(initialFrame.Content, JobAndSqlSummaryCodecUserCancelledFieldOffset)
	}

	nameOrId := DecodeString(frameIterator)
	failureText := DecodeNullableForString(frameIterator)
	sqlSummary, _ := DecodeNullableForSqlSummary(frameIterator)
	var suspensionCause string
	if !frameIterator.PeekNext().IsEndFrame() {
		suspensionCause = DecodeNullableForString(frameIterator)
	}
	FastForwardToEndFrame(frameIterator)

	return pubcontrol.JobAndSqlSummary{
		LightJob:        lightJob,
		JobId:           jobId,
		ExecutionId:     executionId,
		NameOrId:        nameOrId,
		Status:          status,
		SubmissionTime:  submissionTime,
		CompletionTime:  completionTime,
		FailureText:     failureText,
		SqlSummary:      sqlSummary,
		SuspensionCause: suspensionCause,
		UserCancelled:   userCancelled,
	}
}
