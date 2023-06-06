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
	MultiMapTryLockCodecRequestMessageType  = int32(0x021100)
	MultiMapTryLockCodecResponseMessageType = int32(0x021101)

	MultiMapTryLockCodecRequestThreadIdOffset    = proto.PartitionIDOffset + proto.IntSizeInBytes
	MultiMapTryLockCodecRequestLeaseOffset       = MultiMapTryLockCodecRequestThreadIdOffset + proto.LongSizeInBytes
	MultiMapTryLockCodecRequestTimeoutOffset     = MultiMapTryLockCodecRequestLeaseOffset + proto.LongSizeInBytes
	MultiMapTryLockCodecRequestReferenceIdOffset = MultiMapTryLockCodecRequestTimeoutOffset + proto.LongSizeInBytes
	MultiMapTryLockCodecRequestInitialFrameSize  = MultiMapTryLockCodecRequestReferenceIdOffset + proto.LongSizeInBytes

	MultiMapTryLockResponseResponseOffset = proto.ResponseBackupAcksOffset + proto.ByteSizeInBytes
)

// Tries to acquire the lock for the specified key for the specified lease time. After lease time, the lock will be
// released. If the lock is not available, then the current thread becomes disabled for thread scheduling purposes
// and lies dormant until one of two things happens:the lock is acquired by the current thread, or the specified
// waiting time elapses.

func EncodeMultiMapTryLockRequest(name string, key iserialization.Data, threadId int64, lease int64, timeout int64, referenceId int64) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, MultiMapTryLockCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	EncodeLong(initialFrame.Content, MultiMapTryLockCodecRequestThreadIdOffset, threadId)
	EncodeLong(initialFrame.Content, MultiMapTryLockCodecRequestLeaseOffset, lease)
	EncodeLong(initialFrame.Content, MultiMapTryLockCodecRequestTimeoutOffset, timeout)
	EncodeLong(initialFrame.Content, MultiMapTryLockCodecRequestReferenceIdOffset, referenceId)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(MultiMapTryLockCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	EncodeString(clientMessage, name)
	EncodeData(clientMessage, key)

	return clientMessage
}

func DecodeMultiMapTryLockResponse(clientMessage *proto.ClientMessage) bool {
	frameIterator := clientMessage.FrameIterator()
	initialFrame := frameIterator.Next()

	return DecodeBoolean(initialFrame.Content, MultiMapTryLockResponseResponseOffset)
}
