//go:build std || serializer

package serializer

const validationSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://github.com/hazelcast/hazelcast-client-protocol/blob/master/schema/protocol-schema.json",
  "title": "Hazelcast Compact Serialization Schema",
  "type": "object",
  "definitions": {},
  "additionalProperties": false,
  "properties": {
    "namespace": {
      "type": "string"
    },
    "imports": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "uniqueItems": true
    },
    "classes": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "name": {
            "type": "string"
          },
          "fields": {
            "type": "array",
            "items": {
              "type": "object",
              "additionalProperties": false,
              "properties": {
                "name": {
                  "type": "string"
                },
                "type": {
                  "type": [
                    "string"
                  ]
                },
				"external": {
				  "type": "boolean"
				}
              },
              "required": [
                "name",
                "type"
              ]
            }
          }
        },
        "required": [
          "name",
          "fields"
        ]
      }
    }
  },
  "required": [
    "classes"
  ]
}`
