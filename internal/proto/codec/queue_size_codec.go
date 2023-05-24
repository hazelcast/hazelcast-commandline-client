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
	QueueSizeCodecRequestMessageType  = int32(0x030300)
	QueueSizeCodecResponseMessageType = int32(0x030301)

	QueueSizeCodecRequestInitialFrameSize = proto.PartitionIDOffset + proto.IntSizeInBytes

	QueueSizeResponseResponseOffset = proto.ResponseBackupAcksOffset + proto.ByteSizeInBytes
)

// Returns the number of elements in this collection.  If this collection contains more than Integer.MAX_VALUE
// elements, returns Integer.MAX_VALUE

func EncodeQueueSizeRequest(ci *proto.ClientInternal, name string) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(false)

	initialFrame := proto.NewFrameWith(make([]byte, QueueSizeCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(QueueSizeCodecRequestMessageType)
	//TODO: handle error?
	pID, _ := stringToPartitionID(ci, name)
	clientMessage.SetPartitionId(pID)

	EncodeString(clientMessage, name)

	return clientMessage
}

func DecodeQueueSizeResponse(clientMessage *proto.ClientMessage) int32 {
	frameIterator := clientMessage.FrameIterator()
	initialFrame := frameIterator.Next()

	return DecodeInt(initialFrame.Content, QueueSizeResponseResponseOffset)
}
