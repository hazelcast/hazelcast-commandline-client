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

	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
)

const (
	JetGetJobAndSqlSummaryListCodecRequestMessageType  = int32(0xFE0F00)
	JetGetJobAndSqlSummaryListCodecResponseMessageType = int32(0xFE0F01)

	JetGetJobAndSqlSummaryListCodecRequestInitialFrameSize = proto.PartitionIDOffset + proto.IntSizeInBytes
)

func EncodeJetGetJobAndSqlSummaryListRequest() *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, JetGetJobAndSqlSummaryListCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(JetGetJobAndSqlSummaryListCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	return clientMessage
}

func DecodeJetGetJobAndSqlSummaryListResponse(clientMessage *proto.ClientMessage) []control.JobAndSqlSummary {
	frameIterator := clientMessage.FrameIterator()
	// empty initial frame
	frameIterator.Next()

	return DecodeListMultiFrameForJobAndSqlSummary(frameIterator)
}
