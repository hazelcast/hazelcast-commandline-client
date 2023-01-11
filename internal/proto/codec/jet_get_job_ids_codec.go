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
	iserialization "github.com/hazelcast/hazelcast-go-client"
	proto "github.com/hazelcast/hazelcast-go-client"
)

const (
	JetGetJobIdsCodecRequestMessageType  = int32(0xFE0400)
	JetGetJobIdsCodecResponseMessageType = int32(0xFE0401)

	JetGetJobIdsCodecRequestOnlyJobIdOffset  = proto.PartitionIDOffset + proto.IntSizeInBytes
	JetGetJobIdsCodecRequestInitialFrameSize = JetGetJobIdsCodecRequestOnlyJobIdOffset + proto.LongSizeInBytes
)

func EncodeJetGetJobIdsRequest(onlyName string, onlyJobId int64) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, JetGetJobIdsCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	EncodeLong(initialFrame.Content, JetGetJobIdsCodecRequestOnlyJobIdOffset, onlyJobId)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(JetGetJobIdsCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	EncodeNullableForString(clientMessage, onlyName)

	return clientMessage
}

func DecodeJetGetJobIdsResponse(clientMessage *proto.ClientMessage) (response iserialization.Data, isResponseExists bool) {
	frameIterator := clientMessage.FrameIterator()
	frameIterator.Next()
	if frameIterator.HasNext() {
		response = DecodeData(frameIterator)
		isResponseExists = true
	}
	return response, isResponseExists
}
