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
	"encoding/json"

	pubcontrol "github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
	proto "github.com/hazelcast/hazelcast-go-client"
)

const (
	MCGetTimedMemberStateCodecRequestMessageType  = int32(0x200B00)
	MCGetTimedMemberStateCodecResponseMessageType = int32(0x200B01)

	MCGetTimedMemberStateCodecRequestInitialFrameSize = proto.PartitionIDOffset + proto.IntSizeInBytes
)

// Gets the latest TimedMemberState of the member it's called on.

func EncodeMCGetTimedMemberStateRequest() *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, MCGetTimedMemberStateCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(MCGetTimedMemberStateCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	return clientMessage
}

func DecodeMCGetTimedMemberStateResponse(clientMessage *proto.ClientMessage) string {
	frameIterator := clientMessage.FrameIterator()
	// empty initial frame
	frameIterator.Next()

	return DecodeNullableForString(frameIterator)
}

func DecodeTimedMemberStateJsonString(jsonString string) (*pubcontrol.TimedMemberStateWrapper, error) {
	state := &pubcontrol.TimedMemberStateWrapper{}
	err := json.Unmarshal([]byte(jsonString), state)
	if err != nil {
		return nil, err
	}
	return state, nil
}
