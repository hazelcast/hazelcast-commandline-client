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
	MultiMapUnlockCodecRequestMessageType  = int32(0x021300)
	MultiMapUnlockCodecResponseMessageType = int32(0x021301)

	MultiMapUnlockCodecRequestThreadIdOffset    = proto.PartitionIDOffset + proto.IntSizeInBytes
	MultiMapUnlockCodecRequestReferenceIdOffset = MultiMapUnlockCodecRequestThreadIdOffset + proto.LongSizeInBytes
	MultiMapUnlockCodecRequestInitialFrameSize  = MultiMapUnlockCodecRequestReferenceIdOffset + proto.LongSizeInBytes
)

// Releases the lock for the specified key regardless of the lock owner. It always successfully unlocks the key,
// never blocks and returns immediately.

func EncodeMultiMapUnlockRequest(name string, key iserialization.Data, threadId int64, referenceId int64) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, MultiMapUnlockCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	EncodeLong(initialFrame.Content, MultiMapUnlockCodecRequestThreadIdOffset, threadId)
	EncodeLong(initialFrame.Content, MultiMapUnlockCodecRequestReferenceIdOffset, referenceId)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(MultiMapUnlockCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	EncodeString(clientMessage, name)
	EncodeData(clientMessage, key)

	return clientMessage
}
