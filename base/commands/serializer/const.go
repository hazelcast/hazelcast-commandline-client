package serializer

const (
	java       = "java"
	python     = "py"
	typescript = "ts"
	cpp        = "cpp"
	golang     = "go"
	cs         = "cs"
)

var supportedLanguages = []string{java, python, typescript, cpp, golang, cs}

var builtinTypes = []string{
	"boolean",
	"int8",
	"int16",
	"int32",
	"int64",
	"float32",
	"float64",
	"string",
	"decimal",
	"time",
	"date",
	"timestamp",
	"timestampWithTimezone",
	"nullableBoolean",
	"nullableInt8",
	"nullableInt16",
	"nullableInt32",
	"nullableInt64",
	"nullableFloat32",
	"nullableFloat64",
}

var javaTypes = map[string]string{
	"boolean":               "boolean",
	"int8":                  "byte",
	"int16":                 "short",
	"int32":                 "int",
	"int64":                 "long",
	"float32":               "float",
	"float64":               "double",
	"string":                "java.lang.String",
	"decimal":               "java.math.BigDecimal",
	"time":                  "java.time.LocalTime",
	"date":                  "java.time.LocalDate",
	"timestamp":             "java.time.LocalDateTime",
	"timestampWithTimezone": "java.time.OffsetDateTime",
	"nullableBoolean":       "Boolean",
	"nullableInt8":          "Byte",
	"nullableInt16":         "Short",
	"nullableInt32":         "Integer",
	"nullableInt64":         "Long",
	"nullableFloat32":       "Float",
	"nullableFloat64":       "Double",
}

var fixedSizeTypes = map[string]string{
	"boolean": "false",
	"int8":    "0",
	"int16":   "0",
	"int32":   "0",
	"int64":   "0",
	"float32": "0.0",
	"float64": "0.0",
}

const validationSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://github.com/hazelcast/hazelcast-client-protocol/blob/master/schema/protocol-schema.json",
  "title": "Hazelcast Client Protocol Definition",
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

const importsTemplate = `{{if .Cls.Namespace }}package {{.Cls.Namespace}};

{{else}}{{end}}import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;
{{generateImports .}}
import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;`

const compactSerDeserTemplate = `public static final class Serializer implements CompactSerializer<{{.Cls.Name}}> {
        @Nonnull
        @Override
        public {{ .Cls.Name }} read(@Nonnull CompactReader reader) {
{{range $field := .Cls.Fields}}{{read $field}}
{{end}}            return new {{ .Cls.Name }}({{ fieldNames .Cls }});
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull {{ .Cls.Name }} object) {
{{range $field := .Cls.Fields}}            writer.write{{methodName (toJavaType $field.Type) $field.Type }}("{{ $field.Name }}", object.{{ $field.Name }});
{{end}}        }
    };

    public static final CompactSerializer<{{ .Cls.Name }}> HZ_COMPACT_SERIALIZER = new Serializer();`

const constructorsTemplate = `public {{ .Cls.Name }}() {
    }

    public {{ .Cls.Name }}({{fieldTypeAndNames .Cls}}) {
{{range $field := .Cls.Fields}}        this.{{$field.Name}} = {{$field.Name}};
{{end}}    }`

const bodyTemplate = `{{ template "imports" $}}

public class {{ .Cls.Name }} {

    {{ template "compactSerDeser" $}}

{{range $field := .Cls.Fields}}{{fields $field}}
{{end}}
    {{ template "constructors" $ }}

{{range $field := .Cls.Fields}}{{getters $field}}
{{end}}    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        {{ .Cls.Name }} that = ({{ .Cls.Name }}) o;
{{range $field := .Cls.Fields}}{{ equalsBody $field }}
{{end}}
        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
{{hashcodeBody .Cls}}
        return result;
    }

    @Override
    public String toString() {
        return "<{{ .Cls.Name }}> {"
{{range $index, $field := .Cls.Fields}}{{ if eq $index 0}}{{ toStringBodyFirst $field }}{{else}}{{ toStringBody $field }}{{end}}{{end}}                + '}';
    }

}`
