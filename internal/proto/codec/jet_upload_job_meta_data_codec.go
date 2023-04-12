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
	JetUploadJobMetaDataCodecRequestMessageType  = int32(0xFE1100)
	JetUploadJobMetaDataCodecResponseMessageType = int32(0xFE1101)

	JetUploadJobMetaDataCodecRequestSessionIdOffset   = proto.PartitionIDOffset + proto.IntSizeInBytes
	JetUploadJobMetaDataCodecRequestJarOnMemberOffset = JetUploadJobMetaDataCodecRequestSessionIdOffset + proto.UuidSizeInBytes
	JetUploadJobMetaDataCodecRequestInitialFrameSize  = JetUploadJobMetaDataCodecRequestJarOnMemberOffset + proto.BooleanSizeInBytes
)

func EncodeJetUploadJobMetaDataRequest(sessionId types.UUID, jarOnMember bool, fileName string, sha256Hex string, snapshotName string, jobName string, mainClass string, jobParameters []string) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, JetUploadJobMetaDataCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	EncodeUUID(initialFrame.Content, JetUploadJobMetaDataCodecRequestSessionIdOffset, sessionId)
	EncodeBoolean(initialFrame.Content, JetUploadJobMetaDataCodecRequestJarOnMemberOffset, jarOnMember)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(JetUploadJobMetaDataCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	EncodeString(clientMessage, fileName)
	EncodeString(clientMessage, sha256Hex)
	EncodeNullableForString(clientMessage, snapshotName)
	EncodeNullableForString(clientMessage, jobName)
	EncodeNullableForString(clientMessage, mainClass)
	EncodeListMultiFrameForString(clientMessage, jobParameters)

	return clientMessage
}
