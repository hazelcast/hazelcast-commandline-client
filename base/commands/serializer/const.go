//go:build std || serializer

package serializer

const (
	langJava = "java"
)

var supportedLanguages = []string{langJava}

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
