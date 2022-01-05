package compact

import "fmt"

func JavaGenerate(schemas Schemas, namespace string, outputDir string) (string, error) {
	packageName := ""
	packagePrefix := ""
	if namespace == "" {
		packageName = "package $;\n\n"
		packagePrefix = "$."
	}
	fmt.Println(packageName)
	fmt.Println(packagePrefix)
	for _, class := range schemas.Classes {
		fmt.Println(class.Name)
	}
	return "", nil
}

/*

def generate(schemas: typing.Dict[str, typing.Any], namespace: typing.Optional[str], output_dir: str) -> str:
    package = ""
    package_prefix = ""
    if namespace is not None:
        package = f"package {namespace};\n\n"
        package_prefix = f"{namespace}."

    for schema in schemas["classes"]:
        name = schema["name"]
        content = _generate_class(schema, package)
        save_file(join(output_dir, name + ".java"), content)

    hint = _hint(schemas, package_prefix)
    return hint


def _hint(schemas, package_prefix: str):
    hint = "Programmatic configuration:\n"
    for schema in schemas["classes"]:
        hint += f"""compactSerializationConfig({package_prefix}{schema["name"]}.class, "{schema["name"]}", {package_prefix}{schema["name"]}.HZ_COMPACT_SERIALIZER);\n"""
    hint += "\n"
    hint += "Declarative xml configuration:\n"
    hint += "<compact-serialization>\n"
    hint += "   <registered-classes>\n"
    for schema in schemas["classes"]:
        hint += f"""        <class type-name="{schema["name"]}" serializer="{package_prefix}{schema["name"]}.Serializer">{package_prefix}{schema["name"]}</class>\n"""
    hint += "   </registered-classes>\n\n"
    hint += "Declarative yaml configuration:\n"
    hint += "compact-serialization:\n"
    hint += "   registered-classes:\n"
    for schema in schemas["classes"]:
        hint += f"""        - class: {package_prefix}{schema["name"]}\n"""
        hint += f"""          type-name: {schema["name"]}\n"""
        hint += f"""          serializer: {package_prefix}{schema["name"]}.Serializer\n"""
    return hint


_fixed_size_types = {
    "boolean": "false",
    "int8": "0",
    "int16": "0",
    "int32": "0",
    "int64": "0",
    "float32": "0.0",
    "float64": "0.0",
}

_java_types = {
    "boolean": "boolean",
    "int8": "byte",
    "int16": "short",
    "int32": "int",
    "int64": "long",
    "float32": "float",
    "float64": "double",
    "string": "java.lang.String",
    "decimal": "java.math.BigDecimal",
    "time": "java.time.LocalTime",
    "date": "java.time.LocalDate",
    "timestamp": "java.time.LocalDateTime",
    "timestampWithTimezone": "java.time.OffsetDateTime",
    "nullableBoolean": "Boolean",
    "nullableInt8": "Byte",
    "nullableInt16": "Short",
    "nullableInt32": "Integer",
    "nullableInt64": "Long",
    "nullableFloat32": "Float",
    "nullableFloat64": "Double",
}


def _component_type(field):
    field_type = field["type"]
    return field_type[0 : len(field_type) - 2]


def _default_value(field):
    if field["type"] in _fixed_size_types:
        if "default" not in field:
            return _fixed_size_types[field["type"]]
        return _stringify(field["default"])
    if "default" not in field:
        return "null"
    raise RuntimeError(field["type"] + "can not have a default section defined in schema")


def _stringify(value):
    if value == True:
        return "true"
    if value == False:
        return "false"
    return str(value)


def _types(field):
    field_type = field["type"]
    try:
        if field_type.endswith("[]"):
            return _java_types[field_type[0 : len(field_type) - 2]] + "[]"
        return _java_types[field_type]
    except KeyError:
        return field_type


def _method_name(field):
    field_type = field["type"]
    if field_type.endswith("[]"):
        if field_type[0 : len(field_type) - 2] in _java_types:
            return "ArrayOf" + field_type[0].capitalize() + field_type[1 : len(field_type) - 2]
        else:
            return "ArrayOfCompact"
    if field_type in _java_types:
        return field_type[0].capitalize() + field_type[1:]
    else:
        return "Compact"


def _field_names(schema):
    content = ""
    for field in schema["fields"]:
        content += field["name"] + ", "
    return content[0 : len(content) - 2]


def _write(schema):
    content = ""
    for field in schema["fields"]:
        content += "            "
        content += "writer.write" + _method_name(field) + '("' + field["name"] + '", object.' + field["name"] + ");\n"
    return content[0 : len(content) - 1]


def _read(schema):
    content = ""
    for field in schema["fields"]:
        content += "            "
        if _method_name(field) == "arrayOfCompact":
            content += (
                _types(field)
                + " "
                + field["name"]
                + " = reader.readArrayOfCompact("
                + field["name"]
                + ", "
                + _component_type(field)
                + ".class , null);\n"
            )
            content += (
                _types(field)
                + " "
                + field["name"]
                + " = reader.readArrayOfCompact("
                + field["name"]
                + ", "
                + _component_type(field)
                + ".class , null);\n"
            )
        elif _types(field) == "byte":
            content += (
                _types(field)
                + " "
                + field["name"]
                + " = reader.read"
                + _method_name(field)
                + '("'
                + field["name"]
                + '", (byte) '
                + _default_value(field)
                + ");\n"
            )
        elif _types(field) == "short":
            content += (
                _types(field)
                + " "
                + field["name"]
                + " = reader.read"
                + _method_name(field)
                + '("'
                + field["name"]
                + '", (short) '
                + _default_value(field)
                + ");\n"
            )
        elif _types(field) == "float":
            content += (
                _types(field)
                + " "
                + field["name"]
                + " = reader.read"
                + _method_name(field)
                + '("'
                + field["name"]
                + '", (float) '
                + _default_value(field)
                + ");\n"
            )
        else:
            content += (
                _types(field)
                + " "
                + field["name"]
                + " = reader.read"
                + _method_name(field)
                + '("'
                + field["name"]
                + '", '
                + _default_value(field)
                + ");\n"
            )
    content += "            "
    content += "return new " + schema["name"] + "(" + _field_names(schema) + ");"
    return content


def _fields(schema):
    content = ""
    for field in schema["fields"]:
        if _types(field) == "float":
            content += (
                "    private " + _types(field) + " " + field["name"] + " = (float) " + _default_value(field) + ";\n"
            )
        else:
            content += "    private " + _types(field) + " " + field["name"] + " = " + _default_value(field) + ";\n"
    return content[0 : len(content) - 1]


def _field_type_and_names(schema):
    content = ""
    for field in schema["fields"]:
        content += _types(field) + " " + field["name"] + ", "
    return content[0 : len(content) - 2]


def _constructor_body(schema):
    content = ""
    for field in schema["fields"]:
        content += "        "
        content += "this." + field["name"] + " = " + field["name"] + ";\n"
    return content[0 : len(content) - 1]


def _getters(schema):
    content = ""
    for field in schema["fields"]:
        content += "    public " + _types(field) + " get" + field["name"].capitalize() + "() {\n "
        content += "       return " + field["name"] + ";\n"
        content += "    }\n\n"
    return content[0 : len(content) - 2]


def _equals_body(schema):
    content = ""
    for field in schema["fields"]:
        content += "        "
        if field["type"] == "float32":
            content += "if (Float.compare(" + field["name"] + ", that." + field["name"] + ") != 0) return false;"
        elif field["type"] == "float64":
            content += "if (Double.compare(" + field["name"] + ", that." + field["name"] + ") != 0) return false;"
        elif field["type"].endswith("[]"):
            content += "if (!Arrays.equals(" + field["name"] + ", that." + field["name"] + ")) return false;"
        elif field["type"] in _fixed_size_types:
            content += "if (" + field["name"] + " != that." + field["name"] + ") return false;"
        else:
            content += "if (!Objects.equals(" + field["name"] + ", that." + field["name"] + ")) return false;"
        content += "\n"
    return content[0 : len(content) - 1]


def _hashcode_body(schema):
    content = ""
    is_temp_declared = False
    for field in schema["fields"]:
        content += "        "
        if field["type"] == "boolean":
            content += "result = 31 * result + (" + field["name"] + " ? 1 : 0);"
        elif field["type"] == "int64":
            content += "result = 31 * result + (int) (" + field["name"] + " ^ (" + field["name"] + " >>> 32));"
        elif field["type"] == "float32":
            content += (
                "result = 31 * result + ("
                + field["name"]
                + " != +0.0f ? Float.floatToIntBits("
                + field["name"]
                + ") : 0);"
            )
        elif field["type"] == "float64":
            if not is_temp_declared:
                content += "long temp;"
                content += "\n        "
            content += "temp = Double.doubleToLongBits(" + field["name"] + ");"
            content += "\n        "
            content += "result = 31 * result + (int) (temp ^ (temp >>> 32));"
        elif field["type"].endswith("[]"):
            content += "result = 31 * result + Arrays.hashCode(" + field["name"] + ");"
        elif field["type"] in _fixed_size_types:
            content += "result = 31 * result + (int) " + field["name"] + ";"
        else:
            content += "result = 31 * result + Objects.hashCode(" + field["name"] + ");"
        content += "\n"
    return content[0 : len(content) - 1]


def _to_string_body(schema):
    content = ""
    for field in schema["fields"]:
        content += "               "
        if field["type"].endswith("[]"):
            content += ' + ", + ' + field["name"] + '=" + Arrays.toString(' + field["name"] + ")"
        else:
            content += ' + ", + ' + field["name"] + '=" + ' + field["name"]
        content += "\n"
    return content[0 : len(content) - 1]


def _generate_class(schema: typing.Dict[str, typing.Any], package: str) -> str:
    content = f"""{package}import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class {schema["name"]} {{

    public static final class Serializer implements CompactSerializer<{schema["name"]}> {{
        @Nonnull
        @Override
        public {schema["name"]} read(@Nonnull CompactReader reader) {{
{_read(schema)}
        }}

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull {schema["name"]} object) {{
{_write(schema)}
        }}
    }};


    public static final CompactSerializer<{schema["name"]}> HZ_COMPACT_SERIALIZER = new Serializer();

{_fields(schema)}

    public {schema["name"]}() {{
    }}

    public {schema["name"]}({_field_type_and_names(schema)}) {{
{_constructor_body(schema)}
    }}

{_getters(schema)}

    @Override
    public boolean equals(Object o) {{
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        {schema["name"]} that = ({schema["name"]}) o;
{_equals_body(schema)}
        return true;
    }}

    @Override
    public int hashCode() {{
        int result = 0;
{_hashcode_body(schema)}
        return result;
    }}

    @Override
    public String toString() {{
        return "<{schema["name"]}> {{"
{_to_string_body(schema)}
                + '}}';
    }}

}}"""
    return content

*/
