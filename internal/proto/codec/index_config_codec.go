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
	"github.com/hazelcast/hazelcast-go-client/types"
	pubtypes "github.com/hazelcast/hazelcast-go-client/types"
)

const (
	IndexConfigCodecTypeFieldOffset      = 0
	IndexConfigCodecTypeInitialFrameSize = IndexConfigCodecTypeFieldOffset + proto.IntSizeInBytes
)

func DecodeIndexConfig(frameIterator *proto.ForwardFrameIterator) pubtypes.IndexConfig {
	// begin frame
	frameIterator.Next()
	initialFrame := frameIterator.Next()
	_type := DecodeInt(initialFrame.Content, IndexConfigCodecTypeFieldOffset)

	name := DecodeNullableForString(frameIterator)
	attributes := DecodeListMultiFrameForString(frameIterator)
	bitmapIndexOptions := DecodeNullableForBitmapIndexOptions(frameIterator)
	FastForwardToEndFrame(frameIterator)

	return pubtypes.IndexConfig{
		Name:               name,
		Type:               types.IndexType(_type),
		Attributes:         attributes,
		BitmapIndexOptions: bitmapIndexOptions,
	}
}
