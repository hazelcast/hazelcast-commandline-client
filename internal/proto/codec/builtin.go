package codec

import (
	"encoding/binary"
	"strings"

	iserialization "github.com/hazelcast/hazelcast-go-client"
	proto "github.com/hazelcast/hazelcast-go-client"
)

// Encoder for ClientMessage and value
type Encoder func(message *proto.ClientMessage, value interface{})

// Decoder creates iserialization.Data
type Decoder func(frameIterator *proto.ForwardFrameIterator) iserialization.Data

func DecodeNullableForString(frameIterator *proto.ForwardFrameIterator) string {
	if NextFrameIsNullFrame(frameIterator) {
		return ""
	}
	return DecodeString(frameIterator)
}

func NextFrameIsNullFrame(frameIterator *proto.ForwardFrameIterator) bool {
	isNullFrame := frameIterator.PeekNext().IsNullFrame()
	if isNullFrame {
		frameIterator.Next()
	}
	return isNullFrame
}

func DecodeString(frameIterator *proto.ForwardFrameIterator) string {
	return string(frameIterator.Next().Content)
}

func EncodeLong(buffer []byte, offset int32, value int64) {
	binary.LittleEndian.PutUint64(buffer[offset:], uint64(value))
}

func EncodeString(message *proto.ClientMessage, value interface{}) {
	message.AddFrame(proto.NewFrame([]byte(value.(string))))
}

func EncodeData(message *proto.ClientMessage, value interface{}) {
	message.AddFrame(proto.NewFrame(value.(iserialization.Data).ToByteArray()))
}

func DecodeData(frameIterator *proto.ForwardFrameIterator) iserialization.Data {
	return frameIterator.Next().Content
}

func DecodeNullableForData(frameIterator *proto.ForwardFrameIterator) iserialization.Data {
	if NextFrameIsNullFrame(frameIterator) {
		return nil
	}
	return DecodeData(frameIterator)
}

func EncodeBoolean(buffer []byte, offset int32, value bool) {
	if value {
		buffer[offset] = 1
	} else {
		buffer[offset] = 0
	}
}

func DecodeBoolean(buffer []byte, offset int32) bool {
	return buffer[offset] == 1
}

func EncodeMapForStringAndString(message *proto.ClientMessage, values map[string]string) {
	message.AddFrame(proto.BeginFrame.Copy())
	for key, value := range values {
		EncodeString(message, key)
		EncodeString(message, value)
	}
	message.AddFrame(proto.EndFrame.Copy())
}

func EncodeNullableMapForStringAndString(message *proto.ClientMessage, values map[string]string) {
	EncodeNullable(message, values, func(message *proto.ClientMessage, _ interface{}) {
		EncodeMapForStringAndString(message, values)
	})
}

func DecodeMapForStringAndString(iterator *proto.ForwardFrameIterator) interface{} {
	result := map[string]string{}
	iterator.Next()
	for !iterator.PeekNext().IsEndFrame() {
		key := DecodeString(iterator)
		value := DecodeString(iterator)
		result[key] = value
	}
	iterator.Next()
	return result
}

func DecodeNullableMapForStringAndString(frameIterator *proto.ForwardFrameIterator) map[string]string {
	if NextFrameIsNullFrame(frameIterator) {
		return nil
	}
	return DecodeMapForStringAndString(frameIterator).(map[string]string)
}

func EncodeInt(buffer []byte, offset, value int32) {
	binary.LittleEndian.PutUint32(buffer[offset:], uint32(value))
}

func DecodeInt(buffer []byte, offset int32) int32 {
	return int32(binary.LittleEndian.Uint32(buffer[offset:]))
}

func DecodeLong(buffer []byte, offset int32) int64 {
	return int64(binary.LittleEndian.Uint64(buffer[offset:]))
}

func EncodeNullable(message *proto.ClientMessage, value interface{}, encoder Encoder) {
	if value == nil {
		message.AddFrame(proto.NullFrame.Copy())
	} else {
		encoder(message, value)
	}
}

func EncodeNullableForString(message *proto.ClientMessage, value string) {
	if strings.TrimSpace(value) == "" {
		message.AddFrame(proto.NullFrame.Copy())
	} else {
		EncodeString(message, value)
	}
}

func FastForwardToEndFrame(frameIterator *proto.ForwardFrameIterator) {
	expectedEndFrames := 1
	for expectedEndFrames != 0 {
		frame := frameIterator.Next()
		if frame.IsEndFrame() {
			expectedEndFrames--
		} else if frame.IsBeginFrame() {
			expectedEndFrames++
		}
	}
}

func EncodeNullableForData(message *proto.ClientMessage, data iserialization.Data) {
	if data == nil {
		message.AddFrame(proto.NullFrame.Copy())
	} else {
		EncodeData(message, data)
	}
}

func EncodeListMultiFrameForString(message *proto.ClientMessage, values []string) {
	message.AddFrame(proto.NewFrameWith([]byte{}, proto.BeginDataStructureFlag))
	for i := 0; i < len(values); i++ {
		EncodeString(message, values[i])
	}
	message.AddFrame(proto.NewFrameWith([]byte{}, proto.EndDataStructureFlag))
}

func DecodeListMultiFrameForString(frameIterator *proto.ForwardFrameIterator) []string {
	result := make([]string, 0)
	frameIterator.Next()
	for !NextFrameIsDataStructureEndFrame(frameIterator) {
		result = append(result, DecodeString(frameIterator))
	}
	frameIterator.Next()
	return result
}

func NextFrameIsDataStructureEndFrame(frameIterator *proto.ForwardFrameIterator) bool {
	return frameIterator.PeekNext().IsEndFrame()
}

func EncodeListMultiFrameForData(message *proto.ClientMessage, values []iserialization.Data) {
	message.AddFrame(NewBeginFrame())
	for i := 0; i < len(values); i++ {
		EncodeData(message, values[i])
	}
	message.AddFrame(NewEndFrame())
}

func NewBeginFrame() proto.Frame {
	return proto.NewFrameWith([]byte{}, proto.BeginDataStructureFlag)
}

func NewEndFrame() proto.Frame {
	return proto.NewFrameWith([]byte{}, proto.EndDataStructureFlag)
}
