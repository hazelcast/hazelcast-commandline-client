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
	TypeSimpleEntry                          = -22
	TypeSimpleImmutableEntry                 = -23
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
	// TypeUnknown is the type of values with unknown types
	// does not exist in the reference implementation
	TypeUnknown = -2022
	// TypeSkip indicates that this value should be ignored
	TypeSkip = -2023
	// TypeNotDecoded indicates that this value was not decodec
	TypeNotDecoded = -2024
)

var typeToString = map[int32]string{
	TypeNil:                                  "NIL",
	TypePortable:                             "PORTABLE",
	TypeDataSerializable:                     "DATA_SERIALIZABLE",
	TypeByte:                                 "BYTE",
	TypeBool:                                 "BOOL",
	TypeUInt16:                               "UINT16",
	TypeInt16:                                "INT16",
	TypeInt32:                                "INT32",
	TypeInt64:                                "INT64",
	TypeFloat32:                              "FLOAT32",
	TypeFloat64:                              "FLOAT64",
	TypeString:                               "STRING",
	TypeByteArray:                            "BYTE_ARRAY",
	TypeBoolArray:                            "BOOL_ARRAY",
	TypeUInt16Array:                          "UINT16_ARRAY",
	TypeInt16Array:                           "UINT16_ARRAY",
	TypeInt32Array:                           "INT32_ARRAY",
	TypeInt64Array:                           "INT64_ARRAY",
	TypeFloat32Array:                         "FLOAT32_ARRAY",
	TypeFloat64Array:                         "FLOAT64_ARRAY",
	TypeStringArray:                          "STRING_ARRAY",
	TypeUUID:                                 "UUID",
	TypeSimpleEntry:                          "SIMPLE_ENTRY",
	TypeSimpleImmutableEntry:                 "SIMPLE_IMMUTABLE_ENTRY",
	TypeJavaClass:                            "JAVA_CLASS",
	TypeJavaDate:                             "JAVA_DATE",
	TypeJavaBigInteger:                       "JAVA_BIG_INTEGER",
	TypeJavaDecimal:                          "JAVA_DECIMAL",
	TypeJavaArray:                            "JAVA_ARRAY",
	TypeJavaArrayList:                        "JAVA_ARRAY_LIST",
	TypeJavaLinkedList:                       "JAVA_LINKED_LIST",
	TypeJavaDefaultTypeCopyOnWriteArrayList:  "JAVA_COPY_ON_WRITE_ARRAY_LIST",
	TypeJavaDefaultTypeHashMap:               "JAVA_HASHMAP",
	TypeJavaDefaultTypeConcurrentSkipListMap: "JAVA_CONCURRENT_SKIP_LIST_MAP",
	TypeJavaDefaultTypeConcurrentHashMap:     "JAVA_CONCURRENT_HASH_MAP",
	TypeJavaDefaultTypeLinkedHashMap:         "JAVA_LINKED_HASH_MAP",
	TypeJavaDefaultTypeTreeMap:               "JAVA_TREE_MAP",
	TypeJavaDefaultTypeHashSet:               "JAVA_HASH_SET",
	TypeJavaDefaultTypeTreeSet:               "JAVA_TREE_SET",
	TypeJavaDefaultTypeLinkedHashSet:         "JAVA_LINKED_HASH_SET",
	TypeJavaDefaultTypeCopyOnWriteArraySet:   "JAVA_COPY_ON_WRITE_ARRAY_SET",
	TypeJavaDefaultTypeConcurrentSkipListSet: "JAVA_CONCURRENT_SKIP_LIST_SET",
	TypeJavaDefaultTypeArrayDeque:            "JAVA_ARRAY_DEQUE",
	TypeJavaDefaultTypeLinkedBlockingQueue:   "JAVA_LINKED_BLOCKING_QUEUE",
	TypeJavaDefaultTypeArrayBlockingQueue:    "JAVA_ARRAY_BLOCKING_QUEUE",
	TypeJavaDefaultTypePriorityBlockingQueue: "JAVA_PRIORITY_BLOCKING_QUEUE",
	TypeJavaDefaultTypeDelayQueue:            "JAVA_DELAY_QUEUE",
	TypeJavaDefaultTypeSynchronousQueue:      "JAVA_SYNCHRONOUS_QUEUE",
	TypeJavaDefaultTypeLinkedTransferQueue:   "JAVA_LINKED_TRANSFER_QUEUE",
	TypeJavaDefaultTypePriorityQueue:         "JAVA_PRIORITY_QUEUE",
	TypeJavaDefaultTypeOptional:              "JAVA_OPTIONAL",
	TypeJavaLocalDate:                        "JAVA_LOCALDATE",
	TypeJavaLocalTime:                        "JAVA_LOCALTIME",
	TypeJavaLocalDateTime:                    "JAVA_LOCALDATETIME",
	TypeJavaOffsetDateTime:                   "JAVA_OFFSETDATETIME",
	TypeCompact:                              "COMPACT",
	TypeCompactWithSchema:                    "COMPACT_WITH_SCHEMA",
	TypeJavaDefaultTypeSerializable:          "JAVA_SERIALIZABLE",
	TypeJavaDefaultTypeExternalizable:        "JAVA_EXTERNALIZABLE",
	TypeCsharpCLRSerializationType:           "CSHARP_CLR_SERIALIZATION",
	TypePythonPickleSerializationType:        "PYTHON_PICKLE_SERIALIZATION",
	TypeJSONSerialization:                    "JSON",
	TypeGobSerialization:                     "GO_GOB_SERIALIZATION",
	TypeHibernate3TypeHibernateCacheKey:      "HIBERNATE3_CACHE_KEY",
	TypeHibernate3TypeHibernateCacheEntry:    "HIBERNATE3_CACHE_ENTRY",
	TypeHibernate4TypeHibernateCacheKey:      "HIBERNATE3_CACHE_KEY",
	TypeHibernate4TypeHibernateCacheEntry:    "HIBERNATE4_CACHE_ENTRY",
	TypeHibernate5TypeHibernateCacheKey:      "HIBERNATE5_CACHE_KEY",
	TypeHibernate5TypeHibernateCacheEntry:    "HIBERNATE5_CACHE_ENTRY",
	TypeHibernate5TypeHibernateNaturalIDKey:  "HIBERNATE5_NATURAL_ID_KEY",
	TypeJetSerializerFirst:                   "JET_SERIALIZER_FIRST",
	TypeJetSerializerLast:                    "JET_SERIALIZER_LAST",
}

func TypeToString(t int32) string {
	s, ok := typeToString[t]
	if !ok {
		return fmt.Sprintf("UNKNOWN_TYPE:%d", t)
	}
	return s
}
