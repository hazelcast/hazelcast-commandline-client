package serializer

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

const (
	boolType    = "boolean"
	int64Type   = "int64"
	float32Type = "float32"
	float64Type = "float64"
)

var indent4 = strings.Repeat(" ", 4)
var indent8 = strings.Repeat(" ", 8)
var indent12 = strings.Repeat(" ", 12)
var indent16 = strings.Repeat(" ", 16)

func generateJavaClasses(schema Schema) (map[ClassInfo]string, error) {
	classes := make(map[ClassInfo]string)
	javaClasses, err := generate(schema)
	if err != nil {
		return nil, err
	}
	for jc := range javaClasses {
		c := ClassInfo{
			Namespace: schema.Namespace,
			FileName:  fmt.Sprintf("%s.java", jc),
			ClassName: jc,
		}
		classes[c] = javaClasses[jc]
	}
	return classes, nil
}

func generate(schema Schema) (map[string]string, error) {
	classes := make(map[string]string)
	for _, cls := range schema.ClassNames {
		var sb strings.Builder
		err := GenerateClass(cls, schema, &sb)
		if err != nil {
			return nil, fmt.Errorf("generating class: %s: %w", cls.Name, err)
		}
		classes[cls.Name] = sb.String()
	}
	return classes, nil
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"read":              generateReadMethodString,
		"toJavaType":        convertFieldTypeToJavaType,
		"methodName":        methodName,
		"fields":            generateFieldString,
		"fieldTypeAndNames": generateFieldTypeAndNamesString,
		"getters":           generateGetterString,
		"equalsBody":        generateEqualsMethodBody,
		"hashcodeBody":      hashcodeBody,
		"toStringBody":      toStringBody,
		"toStringBodyFirst": toStringBodyFirst,
		"fieldNames":        generateFieldNamesString,
		"generateImports":   generateImportsString,
	}
}

type temp struct {
	templateName string
	template     string
}

type classSchema struct {
	Cls Class
	Sch Schema
}

func GenerateClass(cls Class, sch Schema, w io.Writer) error {
	tmpl, err := template.New("main").Funcs(funcMap()).Parse(javaBodyTemplate)
	if err != nil {
		return err
	}
	temps := []temp{
		{
			"compactSerDeser",
			javaCompactSerDeserTemplate,
		},
		{
			"imports",
			javaImportsTemplate,
		},
		{
			"constructors",
			javaConstructorsTemplate,
		},
	}
	for _, t := range temps {
		tmpl, err := tmpl.New(t.templateName).Parse(t.template)
		tmpl = template.Must(tmpl, err)
	}
	err = tmpl.Execute(w, classSchema{
		Cls: cls,
		Sch: sch,
	})
	return err
}

type TypeInfo struct {
	Type          string
	IsCustomClass bool
	IsArray       bool
	FullType      string
}

func hasArraySuffix(s string) bool {
	return strings.HasSuffix(s, "[]")
}

func trimArraySuffix(fieldType string) string {
	if hasArraySuffix(fieldType) {
		return fieldType[0 : len(fieldType)-2]
	}
	return fieldType
}

func methodName(field TypeInfo, compactType string) string {
	mn := compactType
	if field.IsCustomClass {
		mn = "compact"
	}
	// capitalize fist letter
	mn = strings.ToUpper(string(mn[0])) + mn[1:]
	if field.IsArray {
		return fmt.Sprintf("ArrayOf%s", trimArraySuffix(mn))
	}
	return mn
}

func convertFieldTypeToJavaType(fieldType string) TypeInfo {
	var ti TypeInfo
	var ok bool
	if hasArraySuffix(fieldType) {
		base := trimArraySuffix(fieldType)
		ti.IsArray = true
		ti.Type, ok = javaTypes[base]
		if ok {
			ti.FullType = ti.Type + "[]"
			return ti
		}
		ti.IsCustomClass = true
		ti.Type = base
		ti.FullType = fieldType
		return ti
	}
	ti.Type, ok = javaTypes[fieldType]
	if !ok {
		ti.Type = fieldType
		ti.IsCustomClass = true
	}
	ti.FullType = ti.Type
	return ti
}

func generateFieldTypeAndNamesString(cls Class) string {
	var sb strings.Builder
	for _, f := range cls.Fields {
		ti := convertFieldTypeToJavaType(f.Type)
		sb.WriteString(fmt.Sprintf("%s %s, ", ti.FullType, f.Name))
	}
	s := sb.String()
	if len(s) >= 2 {
		s = s[:len(s)-2]
	}
	return s
}

func generateFieldNamesString(class Class) string {
	var sb strings.Builder
	for _, f := range class.Fields {
		sb.WriteString(fmt.Sprintf("%s, ", f.Name))
	}
	content := sb.String()
	return content[:len(content)-2]
}

func hashcodeBody(cls Class) string {
	var content strings.Builder
	var isTempDeclared bool
	for _, field := range cls.Fields {
		content.WriteString(indent8)
		fn := field.Name
		switch field.Type {
		case boolType:
			fmt.Fprintf(&content, "result = 31 * result + (%s ? 1 : 0);", fn)
		case int64Type:
			fmt.Fprintf(&content, "result = 31 * result + (int) (%s ^ (%s >>> 32));", fn, fn)
		case float32Type:
			fmt.Fprintf(&content, "result = 31 * result + (%s != +0.0f ? Float.floatToIntBits(%s) : 0);", fn, fn)
		case float64Type:
			if !isTempDeclared {
				content.WriteString("long temp;\n")
				isTempDeclared = true
			}
			fmt.Fprintf(&content, fmt.Sprintf(`%stemp = Double.doubleToLongBits(%s);
%sresult = 31 * result + (int) (temp ^ (temp >>> 32));`, indent8, fn, indent8))
		default:
			if hasArraySuffix(field.Type) {
				fmt.Fprintf(&content, fmt.Sprintf("result = 31 * result + Arrays.hashCode(%s);", field.Name))
			} else if _, ok := fixedSizeTypes[field.Type]; ok {
				fmt.Fprintf(&content, "result = 31 * result + (int) %s;", field.Name)
			} else {
				fmt.Fprintf(&content, "result = 31 * result + Objects.hashCode(%s);", field.Name)
			}
		}
		content.WriteString("\n")
	}
	return content.String()
}

func toStringBody(field Field) string {
	if hasArraySuffix(field.Type) {
		return fmt.Sprintf("%s+ \", %s=\" + Arrays.toString(%s)\n", indent16, field.Name, field.Name)
	}
	return fmt.Sprintf("%s+ \", %s=\" + %s\n", indent16, field.Name, field.Name)
}

func toStringBodyFirst(field Field) string {
	if hasArraySuffix(field.Type) {
		return fmt.Sprintf("%s+ \"%s=\" + Arrays.toString(%s)\n", indent16, field.Name, field.Name)
	}
	return fmt.Sprintf("%s+ \"%s=\" + %s\n", indent16, field.Name, field.Name)
}

func generateFieldString(field Field) string {
	ti := convertFieldTypeToJavaType(field.Type)
	return fmt.Sprintf("%sprivate %s %s;", indent4, ti.FullType, field.Name)
}

func generateReadMethodString(field Field) string {
	ti := convertFieldTypeToJavaType(field.Type)
	if ti.IsArray && ti.IsCustomClass {
		return fmt.Sprintf(`%s%s %s = reader.readArrayOfCompact("%s", %s.class);`, indent12, ti.FullType, field.Name, field.Name, ti.Type)
	}
	return fmt.Sprintf(`%s%s %s = reader.read%s("%s");`, indent12, ti.FullType, field.Name, methodName(ti, field.Type), field.Name)
}

func generateGetterString(field Field) string {
	ti := convertFieldTypeToJavaType(field.Type)
	upperName := strings.ToUpper(string(field.Name[0])) + field.Name[1:]
	return fmt.Sprintf(`%spublic %s get%s() {
        return %s;
    }
`, indent4, ti.FullType, upperName, field.Name)
}

func generateImportsString(clsAndSchema classSchema) string {
	var content strings.Builder
	for _, field := range clsAndSchema.Cls.Fields {
		trimmed := trimArraySuffix(field.Type)
		if field.External || isImportedClass(clsAndSchema.Sch, trimmed) {
			fmt.Fprintf(&content, "import %s;", trimmed)
		}
	}
	if content.Len() > 0 {
		c := content.String()
		content.Reset()
		fmt.Fprintf(&content, "\n%s\n", c)
	}
	return content.String()
}

func isImportedClass(sch Schema, typ string) bool {
	for _, cls := range sch.Classes {
		fullName := getClassFullName(cls.Name, sch.Namespace)
		if typ == fullName {
			return false
		}
		if !strings.Contains(typ, ".") && typ == cls.Name {
			return false
		}
	}
	return !isBuiltInType(typ)
}

func generateEqualsMethodBody(field Field) string {
	ti := convertFieldTypeToJavaType(field.Type)
	var s string
	if field.Type == "float32" {
		s = fmt.Sprintf("if (Float.compare(%s, that.%s) != 0) return false;", field.Name, field.Name)
	} else if field.Type == "float64" {
		s = fmt.Sprintf("if (Double.compare(%s, that.%s) != 0) return false;", field.Name, field.Name)
	} else if ti.IsArray {
		s = fmt.Sprintf("if (!Arrays.equals(%s, that.%s)) return false;", field.Name, field.Name)
	} else if _, ok := fixedSizeTypes[field.Type]; ok {
		s = fmt.Sprintf("if (%s != that.%s) return false;", field.Name, field.Name)
	} else {
		s = fmt.Sprintf("if (!Objects.equals(%s, that.%s)) return false;", field.Name, field.Name)
	}
	return fmt.Sprintf("%s%s", indent8, s)
}
