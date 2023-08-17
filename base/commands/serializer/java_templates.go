package serializer

const javaImportsTemplate = `{{if .Cls.Namespace }}package {{.Cls.Namespace}};

{{else}}{{end}}import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;
{{generateImports .}}
import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;`

const javaCompactSerDeserTemplate = `public static final class Serializer implements CompactSerializer<{{.Cls.Name}}> {
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

const javaConstructorsTemplate = `public {{ .Cls.Name }}() {
    }

    public {{ .Cls.Name }}({{fieldTypeAndNames .Cls}}) {
{{range $field := .Cls.Fields}}        this.{{$field.Name}} = {{$field.Name}};
{{end}}    }`

const javaBodyTemplate = `{{ template "imports" $}}

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
