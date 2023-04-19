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
)

const (
	JetExportSnapshotCodecRequestMessageType  = int32(0xFE0A00)
	JetExportSnapshotCodecResponseMessageType = int32(0xFE0A01)

	JetExportSnapshotCodecRequestJobIdOffset      = proto.PartitionIDOffset + proto.IntSizeInBytes
	JetExportSnapshotCodecRequestCancelJobOffset  = JetExportSnapshotCodecRequestJobIdOffset + proto.LongSizeInBytes
	JetExportSnapshotCodecRequestInitialFrameSize = JetExportSnapshotCodecRequestCancelJobOffset + proto.BooleanSizeInBytes
)

func EncodeJetExportSnapshotRequest(jobId int64, name string, cancelJob bool) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(false)

	initialFrame := proto.NewFrameWith(make([]byte, JetExportSnapshotCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	EncodeLong(initialFrame.Content, JetExportSnapshotCodecRequestJobIdOffset, jobId)
	EncodeBoolean(initialFrame.Content, JetExportSnapshotCodecRequestCancelJobOffset, cancelJob)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(JetExportSnapshotCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	EncodeString(clientMessage, name)

	return clientMessage
}
