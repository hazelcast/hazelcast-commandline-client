package demo

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

const addMapping = `CREATE OR REPLACE MAPPING "{{ .map_name }}"
(
	{{ range $key, $val := .fields -}}
	{{ $key }} {{ $val }},
	{{ end -}}
	__key VARCHAR
)
TYPE IMap
OPTIONS (
    'keyFormat' = 'varchar',
    'valueFormat' = 'json-flat'
);
`

func GenerateMappingQuery(mapName string, fields map[string]any) (string, error) {
	sqlFields := map[string]string{}
	for k, v := range fields {
		sqlFields[k] = findSqlType(v)
	}
	values := map[string]any{
		"map_name": mapName,
		// template sorts map by key
		"fields": sqlFields,
	}
	t, err := template.New("query").Parse(addMapping)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, values)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func findSqlType(v any) string {
	switch t := v.(type) {
	case bool:
		return "BOOLEAN"
	case string:
		return "VARCHAR"
	case int8, byte:
		return "TINYINT"
	case int16:
		return "SMALLINT"
	case int32:
		return "INTEGER"
	case int64:
		return "BIGINT"
	case float32:
		return "REAL"
	case float64:
		return "DOUBLE"
	case time.Time:
		return "TIMESTAMP WITH TIME ZONE"
	default:
		panic(fmt.Sprintf("sql type conversion not supported for type %s", t))
	}
}
