//go:build std || serializer

package serializer

const javaImportsTemplate = `{{if .Class.Namespace }}package {{.Class.Namespace}};

{{else}}{{end}}import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;
{{generateImports .}}
import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;`

const javaCompactSerializerTemplate = `public static final class Serializer implements CompactSerializer<{{.Class.Name}}> {
        @Nonnull
        @Override
        public {{ .Class.Name }} read(@Nonnull CompactReader reader) {
{{range $field := .Class.Fields}}{{read $field}}
{{end}}            return new {{ .Class.Name }}({{ fieldNames .Class }});
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull {{ .Class.Name }} object) {
{{range $field := .Class.Fields}}            writer.write{{methodName (toJavaType $field.Type) $field.Type }}("{{ $field.Name }}", object.{{ $field.Name }});
{{end}}        }
    };

    public static final CompactSerializer<{{ .Class.Name }}> HZ_COMPACT_SERIALIZER = new Serializer();`

const javaConstructorsTemplate = `public {{ .Class.Name }}() {
    }

    public {{ .Class.Name }}({{fieldTypeAndNames .Class}}) {
{{range $field := .Class.Fields}}        this.{{$field.Name}} = {{$field.Name}};
{{end}}    }`

const javaBodyTemplate = `{{ template "imports" $}}

public class {{ .Class.Name }} {

    {{ template "compactSerDeser" $}}

{{range $field := .Class.Fields}}{{fields $field}}
{{end}}
    {{ template "constructors" $ }}

{{range $field := .Class.Fields}}{{getters $field}}
{{end}}    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        {{ .Class.Name }} that = ({{ .Class.Name }}) o;
{{range $field := .Class.Fields}}{{ equalsBody $field }}
{{end}}
        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
{{hashcodeBody .Class}}
        return result;
    }

    @Override
    public String toString() {
        return "<{{ .Class.Name }}> {"
{{range $index, $field := .Class.Fields}}{{ if eq $index 0}}{{ toStringBodyFirst $field }}{{else}}{{ toStringBody $field }}{{end}}{{end}}                + '}';
    }

}`
