/*
 * Copyright (c) 2008-2021, Hazelcast, Inc. All Rights Reserved.
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

package serialization

import "fmt"

const (
	TypeNil                                  = 0
	TypePortable                             = -1
	TypeDataSerializable                     = -2
	TypeByte                                 = -3
	TypeBool                                 = -4
	TypeUInt16                               = -5
	TypeInt16                                = -6
	TypeInt32                                = -7
	TypeInt64                                = -8
	TypeFloat32                              = -9
	TypeFloat64                              = -10
	TypeString                               = -11
	TypeByteArray                            = -12
	TypeBoolArray                            = -13
	TypeUInt16Array                          = -14
	TypeInt16Array                           = -15
	TypeInt32Array                           = -16
	TypeInt64Array                           = -17
	TypeFloat32Array                         = -18
	TypeFloat64Array                         = -19
	TypeStringArray                          = -20
	TypeUUID                                 = -21
	TypeJavaClass                            = -24
	TypeJavaDate                             = -25
	TypeJavaBigInteger                       = -26
	TypeJavaDecimal                          = -27
	TypeJavaArray                            = -28
	TypeJavaArrayList                        = -29
	TypeJavaLinkedList                       = -30
	TypeJavaDefaultTypeCopyOnWriteArrayList  = -31
	TypeJavaDefaultTypeHashMap               = -32
	TypeJavaDefaultTypeConcurrentSkipListMap = -33
	TypeJavaDefaultTypeConcurrentHashMap     = -34
	TypeJavaDefaultTypeLinkedHashMap         = -35
	TypeJavaDefaultTypeTreeMap               = -36
	TypeJavaDefaultTypeHashSet               = -37
	TypeJavaDefaultTypeTreeSet               = -38
	TypeJavaDefaultTypeLinkedHashSet         = -39
	TypeJavaDefaultTypeCopyOnWriteArraySet   = -40
	TypeJavaDefaultTypeConcurrentSkipListSet = -41
	TypeJavaDefaultTypeArrayDeque            = -42
	TypeJavaDefaultTypeLinkedBlockingQueue   = -43
	TypeJavaDefaultTypeArrayBlockingQueue    = -44
	TypeJavaDefaultTypePriorityBlockingQueue = -45
	TypeJavaDefaultTypeDelayQueue            = -46
	TypeJavaDefaultTypeSynchronousQueue      = -47
	TypeJavaDefaultTypeLinkedTransferQueue   = -48
	TypeJavaDefaultTypePriorityQueue         = -49
	TypeJavaDefaultTypeOptional              = -50
	TypeJavaLocalDate                        = -51
	TypeJavaLocalTime                        = -52
	TypeJavaLocalDateTime                    = -53
	TypeJavaOffsetDateTime                   = -54
	TypeCompact                              = -55
	TypeCompactWithSchema                    = -56
	TypeJavaDefaultTypeSerializable          = -100
	TypeJavaDefaultTypeExternalizable        = -101
	TypeCsharpCLRSerializationType           = -110
	TypePythonPickleSerializationType        = -120
	TypeJSONSerialization                    = -130
	TypeGobSerialization                     = -140
	TypeHibernate3TypeHibernateCacheKey      = -200
	TypeHibernate3TypeHibernateCacheEntry    = -201
	TypeHibernate4TypeHibernateCacheKey      = -202
	TypeHibernate4TypeHibernateCacheEntry    = -203
	TypeHibernate5TypeHibernateCacheKey      = -204
	TypeHibernate5TypeHibernateCacheEntry    = -205
	TypeHibernate5TypeHibernateNaturalIDKey  = -206
	TypeJetSerializerFirst                   = -300
	TypeJetSerializerLast                    = -399
)

func TypeToString(t int32) string {
	var s string
	switch t {
	case TypeNil:
		s = "NIL"
	case TypePortable:
		s = "PORTABLE"
	case TypeDataSerializable:
		s = "DATA_SERIALIZABLE"
	case TypeByte:
		s = "BYTE"
	case TypeBool:
		s = "BOOL"
	case TypeUInt16:
		s = "UINT16"
	case TypeInt16:
		s = "INT16"
	case TypeInt32:
		s = "INT32"
	case TypeInt64:
		s = "INT64"
	case TypeFloat32:
		s = "FLOAT32"
	case TypeFloat64:
		s = "FLOAT64"
	case TypeString:
		s = "STRING"
	case TypeByteArray:
		s = "BYTE_ARRAY"
	case TypeBoolArray:
		s = "BOOL_ARRAY"
	case TypeUInt16Array:
		s = "UINT16_ARRAY"
	case TypeInt16Array:
		s = "UINT16_ARRAY"
	case TypeInt32Array:
		s = "INT32_ARRAY"
	case TypeInt64Array:
		s = "INT64_ARRAY"
	case TypeFloat32Array:
		s = "FLOAT32_ARRAY"
	case TypeFloat64Array:
		s = "FLOAT64_ARRAY"
	case TypeStringArray:
		s = "STRING_ARRAY"
	case TypeUUID:
		s = "UUID"
	case TypeJavaClass:
		s = "JAVA_CLASS"
	case TypeJavaDate:
		s = "JAVA_DATA"
	case TypeJavaBigInteger:
		s = "JAVA_BIG_INTEGER"
	case TypeJavaDecimal:
		s = "JAVA_DECIMAL"
	case TypeJavaArray:
		s = "JAVA_ARRAY"
	case TypeJavaArrayList:
		s = "JAVA_ARRAY_LIST"
	case TypeJavaLinkedList:
		s = "JAVA_LINKED_LIST"
	case TypeJavaDefaultTypeCopyOnWriteArrayList:
		s = "JAVA_DEFAULT_TYPE_COPY_ON_WRITE_ARRAY_LIST"
	case TypeJavaDefaultTypeHashMap:
		s = "JAVA_HASHMAP"
	case TypeJavaDefaultTypeConcurrentSkipListMap:
		s = "JAVA_CONCURRENT_SKIP_LIST_MAP"
	case TypeJavaDefaultTypeConcurrentHashMap:
		s = "JAVA_CONCURRENT_HASH_MAP"
	case TypeJavaDefaultTypeLinkedHashMap:
		s = "JAVA_LINKED_HASH_MAP"
	case TypeJavaDefaultTypeTreeMap:
		s = "JAVA_TREE_MAP"
	case TypeJavaDefaultTypeHashSet:
		s = "JAVA_HASH_SET"
	case TypeJavaDefaultTypeTreeSet:
		s = "JAVA_TREE_SET"
	case TypeJavaDefaultTypeLinkedHashSet:
		s = "JAVA_LINKED_HASH_SET"
	case TypeJavaDefaultTypeCopyOnWriteArraySet:
		s = "JAVA_COPY_ON_WRITE_ARRAY_SET"
	case TypeJavaDefaultTypeConcurrentSkipListSet:
		s = "JAVA_CONCURRENT_SKIP_LIST_SET"
	case TypeJavaDefaultTypeArrayDeque:
		s = "JAVA_ARRAY_DEQUE"
	case TypeJavaDefaultTypeLinkedBlockingQueue:
		s = "JAVA_LINKED_BLOCKING_QUEUE"
	case TypeJavaDefaultTypeArrayBlockingQueue:
		s = "JAVA_ARRAY_BLOCKING_QUEUE"
	case TypeJavaDefaultTypePriorityBlockingQueue:
		s = "JAVA_PRIORITY_BLOCKING_QUEUE"
	case TypeJavaDefaultTypeDelayQueue:
		s = "JAVA_DELAY_QUEUE"
	case TypeJavaDefaultTypeSynchronousQueue:
		s = "JAVA_SYNCHRONOUS_QUEUE"
	case TypeJavaDefaultTypeLinkedTransferQueue:
		s = "JAVA_LINKED_TRANSFER_QUEUE"
	case TypeJavaDefaultTypePriorityQueue:
		s = "JAVA_PRIORITY_QUEUE"
	case TypeJavaDefaultTypeOptional:
		s = "JAVA_OPTIONAL"
	case TypeJavaLocalDate:
		s = "JAVA_LOCALDATE"
	case TypeJavaLocalTime:
		s = "JAVA_LOCALTIME"
	case TypeJavaLocalDateTime:
		s = "JAVA_LOCALDATETIME"
	case TypeJavaOffsetDateTime:
		s = "JAVA_OFFSETDATETIME"
	case TypeCompact:
		s = "COMPACT"
	case TypeCompactWithSchema:
		s = "COMPACT_WITH_SCHEMA"
	case TypeJavaDefaultTypeSerializable:
		s = "JAVA_SERIALIZABLE"
	case TypeJavaDefaultTypeExternalizable:
		s = "JAVA_EXTERNALIZABLE"
	case TypeCsharpCLRSerializationType:
		s = "CSHARP_CLR_SERIALIZATION"
	case TypePythonPickleSerializationType:
		s = "PYTHON_PICKLE_SERIALIZATION"
	case TypeJSONSerialization:
		s = "JSON"
	case TypeGobSerialization:
		s = "GO_GOB_SERIALIZATION"
	case TypeHibernate3TypeHibernateCacheKey:
		s = "HIBERNATE3_CACHE_KEY"
	case TypeHibernate3TypeHibernateCacheEntry:
		s = "HIBERNATE3_CACHE_ENTRY"
	case TypeHibernate4TypeHibernateCacheKey:
		s = "HIBERNATE3_CACHE_KEY"
	case TypeHibernate4TypeHibernateCacheEntry:
		s = "HIBERNATE4_CACHE_ENTRY"
	case TypeHibernate5TypeHibernateCacheKey:
		s = "HIBERNATE5_CACHE_KEY"
	case TypeHibernate5TypeHibernateCacheEntry:
		s = "HIBERNATE5_CACHE_ENTRY"
	case TypeHibernate5TypeHibernateNaturalIDKey:
		s = "HIBERNATE5_NATURAL_ID_KEY"
	case TypeJetSerializerFirst:
		s = "JET_SERIALIZER_FIRST"
	case TypeJetSerializerLast:
		s = "JET_SERIALIZER_LAST"
	default:
		s = fmt.Sprintf("UNKNOWN_TYPE:%d", t)
	}
	return fmt.Sprintf("%s", s)
}
