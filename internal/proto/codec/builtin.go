package codec

import (
	"encoding/binary"
	"strings"

	iserialization "github.com/hazelcast/hazelcast-go-client"
	proto "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"

	control "github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
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

func EncodeNullableForSqlSummary(message *proto.ClientMessage, summary *control.SqlSummary) {
	if summary == nil {
		message.AddFrame(proto.NullFrame.Copy())
	} else {
		EncodeSqlSummary(message, *summary)
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

func DecodeEntryListForDataAndData(frameIterator *proto.ForwardFrameIterator) []proto.Pair {
	result := make([]proto.Pair, 0)
	frameIterator.Next()
	for !NextFrameIsDataStructureEndFrame(frameIterator) {
		key := DecodeData(frameIterator)
		value := DecodeData(frameIterator)
		result = append(result, proto.NewPair(key, value))
	}
	frameIterator.Next()
	return result
}

func EncodeUUID(buffer []byte, offset int32, uuid types.UUID) {
	isNullEncode := uuid.Default()
	EncodeBoolean(buffer, offset, isNullEncode)
	if isNullEncode {
		return
	}
	bufferOffset := offset + proto.BooleanSizeInBytes
	EncodeLong(buffer, bufferOffset, int64(uuid.MostSignificantBits()))
	EncodeLong(buffer, bufferOffset+proto.LongSizeInBytes, int64(uuid.LeastSignificantBits()))
}

func DecodeUUID(buffer []byte, offset int32) types.UUID {
	isNull := DecodeBoolean(buffer, offset)
	if isNull {
		return types.UUID{}
	}
	mostSignificantOffset := offset + proto.BooleanSizeInBytes
	leastSignificantOffset := mostSignificantOffset + proto.LongSizeInBytes
	mostSignificant := uint64(DecodeLong(buffer, mostSignificantOffset))
	leastSignificant := uint64(DecodeLong(buffer, leastSignificantOffset))
	return types.NewUUIDWith(mostSignificant, leastSignificant)
}

func EncodeByteArray(message *proto.ClientMessage, value []byte) {
	message.AddFrame(proto.NewFrame(value))
}

func DecodeNullableForSqlSummary(it *proto.ForwardFrameIterator) (control.SqlSummary, bool) {
	if NextFrameIsNullFrame(it) {
		return control.SqlSummary{}, false
	}
	ss := DecodeSqlSummary(it)
	return ss, true
}

func DecodeListMultiFrameForJobAndSqlSummary(frameIterator *proto.ForwardFrameIterator) []control.JobAndSqlSummary {
	result := []control.JobAndSqlSummary{}
	frameIterator.Next()
	for !NextFrameIsDataStructureEndFrame(frameIterator) {
		result = append(result, DecodeJobAndSqlSummary(frameIterator))
	}
	frameIterator.Next()
	return result
}

func DecodeListMultiFrameForData(frameIterator *proto.ForwardFrameIterator) []*iserialization.Data {
	result := make([]*iserialization.Data, 0)
	frameIterator.Next()
	for !NextFrameIsDataStructureEndFrame(frameIterator) {
		d := DecodeData(frameIterator)
		result = append(result, &d)
	}
	frameIterator.Next()
	return result
}
