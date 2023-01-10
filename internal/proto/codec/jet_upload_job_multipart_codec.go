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
	"github.com/hazelcast/hazelcast-go-client/types"
)

const (
	JetUploadJobMultipartCodecRequestMessageType  = int32(0xFE1100)
	JetUploadJobMultipartCodecResponseMessageType = int32(0xFE1101)

	JetUploadJobMultipartCodecRequestSessionIdOffset         = proto.PartitionIDOffset + proto.IntSizeInBytes
	JetUploadJobMultipartCodecRequestCurrentPartNumberOffset = JetUploadJobMultipartCodecRequestSessionIdOffset + proto.UuidSizeInBytes
	JetUploadJobMultipartCodecRequestTotalPartNumberOffset   = JetUploadJobMultipartCodecRequestCurrentPartNumberOffset + proto.IntSizeInBytes
	JetUploadJobMultipartCodecRequestPartSizeOffset          = JetUploadJobMultipartCodecRequestTotalPartNumberOffset + proto.IntSizeInBytes
	JetUploadJobMultipartCodecRequestInitialFrameSize        = JetUploadJobMultipartCodecRequestPartSizeOffset + proto.IntSizeInBytes

	JetUploadJobMultipartResponseResponseOffset = proto.ResponseBackupAcksOffset + proto.ByteSizeInBytes
)

func EncodeJetUploadJobMultipartRequest(sessionId types.UUID, currentPartNumber int32, totalPartNumber int32, partData []byte, partSize int32) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, JetUploadJobMultipartCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	EncodeUUID(initialFrame.Content, JetUploadJobMultipartCodecRequestSessionIdOffset, sessionId)
	EncodeInt(initialFrame.Content, JetUploadJobMultipartCodecRequestCurrentPartNumberOffset, currentPartNumber)
	EncodeInt(initialFrame.Content, JetUploadJobMultipartCodecRequestTotalPartNumberOffset, totalPartNumber)
	EncodeInt(initialFrame.Content, JetUploadJobMultipartCodecRequestPartSizeOffset, partSize)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(JetUploadJobMultipartCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	EncodeByteArray(clientMessage, partData)

	return clientMessage
}

func DecodeJetUploadJobMultipartResponse(clientMessage *proto.ClientMessage) bool {
	frameIterator := clientMessage.FrameIterator()
	initialFrame := frameIterator.Next()

	return DecodeBoolean(initialFrame.Content, JetUploadJobMultipartResponseResponseOffset)
}
