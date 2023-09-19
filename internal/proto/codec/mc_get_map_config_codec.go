/*
* Copyright (c) 2008-2022, Hazelcast, Inc. All Rights Reserved.
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
	pubtypes "github.com/hazelcast/hazelcast-go-client/types"
)

const (
	MCGetMapConfigCodecRequestMessageType  = int32(0x200300)
	MCGetMapConfigCodecResponseMessageType = int32(0x200301)

	MCGetMapConfigCodecRequestInitialFrameSize = proto.PartitionIDOffset + proto.IntSizeInBytes

	MCGetMapConfigResponseInMemoryFormatOffset    = proto.ResponseBackupAcksOffset + proto.ByteSizeInBytes
	MCGetMapConfigResponseBackupCountOffset       = MCGetMapConfigResponseInMemoryFormatOffset + proto.IntSizeInBytes
	MCGetMapConfigResponseAsyncBackupCountOffset  = MCGetMapConfigResponseBackupCountOffset + proto.IntSizeInBytes
	MCGetMapConfigResponseTimeToLiveSecondsOffset = MCGetMapConfigResponseAsyncBackupCountOffset + proto.IntSizeInBytes
	MCGetMapConfigResponseMaxIdleSecondsOffset    = MCGetMapConfigResponseTimeToLiveSecondsOffset + proto.IntSizeInBytes
	MCGetMapConfigResponseMaxSizeOffset           = MCGetMapConfigResponseMaxIdleSecondsOffset + proto.IntSizeInBytes
	MCGetMapConfigResponseMaxSizePolicyOffset     = MCGetMapConfigResponseMaxSizeOffset + proto.IntSizeInBytes
	MCGetMapConfigResponseReadBackupDataOffset    = MCGetMapConfigResponseMaxSizePolicyOffset + proto.IntSizeInBytes
	MCGetMapConfigResponseEvictionPolicyOffset    = MCGetMapConfigResponseReadBackupDataOffset + proto.BooleanSizeInBytes
)

// Gets the config of a map on the member it's called on.

func EncodeMCGetMapConfigRequest(mapName string) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, MCGetMapConfigCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(MCGetMapConfigCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	EncodeString(clientMessage, mapName)

	return clientMessage
}

func DecodeMCGetMapConfigResponse(clientMessage *proto.ClientMessage) (inMemoryFormat int32, backupCount int32, asyncBackupCount int32, timeToLiveSeconds int32, maxIdleSeconds int32, maxSize int32, maxSizePolicy int32, readBackupData bool, evictionPolicy int32, mergePolicy string, globalIndexes []pubtypes.IndexConfig) {
	frameIterator := clientMessage.FrameIterator()
	initialFrame := frameIterator.Next()

	inMemoryFormat = DecodeInt(initialFrame.Content, MCGetMapConfigResponseInMemoryFormatOffset)
	backupCount = DecodeInt(initialFrame.Content, MCGetMapConfigResponseBackupCountOffset)
	asyncBackupCount = DecodeInt(initialFrame.Content, MCGetMapConfigResponseAsyncBackupCountOffset)
	timeToLiveSeconds = DecodeInt(initialFrame.Content, MCGetMapConfigResponseTimeToLiveSecondsOffset)
	maxIdleSeconds = DecodeInt(initialFrame.Content, MCGetMapConfigResponseMaxIdleSecondsOffset)
	maxSize = DecodeInt(initialFrame.Content, MCGetMapConfigResponseMaxSizeOffset)
	maxSizePolicy = DecodeInt(initialFrame.Content, MCGetMapConfigResponseMaxSizePolicyOffset)
	readBackupData = DecodeBoolean(initialFrame.Content, MCGetMapConfigResponseReadBackupDataOffset)
	evictionPolicy = DecodeInt(initialFrame.Content, MCGetMapConfigResponseEvictionPolicyOffset)
	mergePolicy = DecodeString(frameIterator)
	globalIndexes = DecodeListMultiFrameForIndexConfig(frameIterator)
	return inMemoryFormat, backupCount, asyncBackupCount, timeToLiveSeconds, maxIdleSeconds, maxSize, maxSizePolicy, readBackupData, evictionPolicy, mergePolicy, globalIndexes
}
