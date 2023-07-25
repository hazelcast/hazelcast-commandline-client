package serializer

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

func generateJavaClasses(schema Schema, classes map[ClassInfo]string) {
	javaClasses := generate(schema)
	for jc := range javaClasses {
		c := ClassInfo{
			Namespace: schema.Namespace,
			FileName:  fmt.Sprintf("%s.java", jc),
			ClassName: jc,
		}
		classes[c] = javaClasses[jc]
	}
}

func generate(schema Schema) map[string]string {
	classes := make(map[string]string)
	for _, cls := range schema.ClassNames {
		var sb strings.Builder
		err := GenerateClass(cls, schema, &sb)
		if err != nil {
			fmt.Println(err)
		}
		classes[cls.Name] = sb.String()
	}
	return classes
}

func GenerateClass(cls Class, sch Schema, w io.Writer) error {
	tmpl, err := template.New("main").Funcs(template.FuncMap{
		"read":              read,
		"toJavaType":        toJavaType,
		"methodName":        methodName,
		"fields":            fields,
		"fieldTypeAndNames": fieldTypeAndNames,
		"getters":           getters,
		"equalsBody":        equalsBody,
		"hashcodeBody":      hashcodeBody,
		"toStringBody":      toStringBody,
		"toStringBodyFirst": toStringBodyFirst,
		"fieldNames":        fieldNames,
		"generateImports":   generateImports,
	}).Parse(bodyTemplate)
	if err != nil {
		return err
	}
	for _, t := range []struct {
		templateName string
		template     string
	}{
		{
			"compactSerDeser",
			compactSerDeserTemplate,
		},
		{
			"imports",
			importsTemplate,
		},
		{
			"constructors",
			constructorsTemplate,
		},
	} {
		tmpl, err := tmpl.New(t.templateName).Parse(t.template)
		tmpl = template.Must(tmpl, err)
	}
	err = tmpl.Execute(w, struct {
		Cls    Class
		Schema Schema
	}{
		cls,
		sch,
	})
	return err
}

type TypeInfo struct {
	Type          string
	IsCustomClass bool
	IsArr         bool
	FullType      string
}

func arrayOf(fieldType string) string {
	if strings.HasSuffix(fieldType, "[]") {
		return fieldType[0 : len(fieldType)-2]
	}
	return fieldType
}

func methodName(field TypeInfo, compactType string) string {
	var mn string
	if field.IsCustomClass {
		mn = "compact"
	} else {
		mn = compactType
	}
	// capitalize fist letter
	mn = strings.ToUpper(string(mn[0])) + mn[1:]
	if field.IsArr {
		return fmt.Sprintf("ArrayOf%s", arrayOf(mn))
	}
	return mn
}

func toJavaType(fieldType string) TypeInfo {
	var ti TypeInfo
	var ok bool
	if strings.HasSuffix(fieldType, "[]") {
		base := arrayOf(fieldType)
		ti.IsArr = true
		ti.Type, ok = javaTypes[base]
		if ok {
			ti.FullType = ti.Type + "[]"
		} else {
			ti.IsCustomClass = true
			ti.Type = base
			ti.FullType = fieldType
		}
	} else {
		ti.Type, ok = javaTypes[fieldType]
		if !ok {
			ti.Type = fieldType
			ti.IsCustomClass = true
		}
		ti.FullType = ti.Type
	}
	return ti
}

func fieldTypeAndNames(cls Class) string {
	var sb strings.Builder
	for _, f := range cls.Fields {
		ti := toJavaType(f.Type)
		sb.WriteString(fmt.Sprintf("%s %s, ", ti.FullType, f.Name))
	}
	s := sb.String()
	if len(s) >= 2 {
		s = s[:len(s)-2]
	}
	return s
}

func fieldNames(class Class) string {
	var sb strings.Builder
	for _, f := range class.Fields {
		sb.WriteString(fmt.Sprintf("%s, ", f.Name))
	}
	content := sb.String()
	return content[:len(content)-2]
}

func hashcodeBody(cls Class) string {
	const indentation = "        "
	var content string
	var isTempDeclared bool
	for _, field := range cls.Fields {
		var line string
		fn := field.Name
		switch field.Type {
		case "boolean":
			line = fmt.Sprintf("result = 31 * result + (%s ? 1 : 0);", fn)
		case "int64":
			line = fmt.Sprintf("result = 31 * result + (int) (%s ^ (%s >>> 32));", fn, fn)
		case "float32":
			line = fmt.Sprintf("result = 31 * result + (%s != +0.0f ? Float.floatToIntBits(%s) : 0);", fn, fn)
		case "float64":
			if !isTempDeclared {
				line = "long temp;\n"
				isTempDeclared = true
			}
			line += fmt.Sprintf(`%stemp = Double.doubleToLongBits(%s);
%sresult = 31 * result + (int) (temp ^ (temp >>> 32));`, indentation, fn, indentation)
		default:
			if strings.HasSuffix(field.Type, "[]") {
				line = fmt.Sprintf("result = 31 * result + Arrays.hashCode(%s);", field.Name)
			} else if _, ok := fixedSizeTypes[field.Type]; ok {
				line = fmt.Sprintf("result = 31 * result + (int) %s;", field.Name)
			} else {
				line = fmt.Sprintf("result = 31 * result + Objects.hashCode(%s);", field.Name)
			}
		}

		content += "        " + line + "\n"
	}
	return content
}

func toStringBody(field Field) string {
	if strings.HasSuffix(field.Type, "[]") {
		return fmt.Sprintf("                + \", %s=\" + Arrays.toString(%s)\n", field.Name, field.Name)
	}
	return fmt.Sprintf("                + \", %s=\" + %s\n", field.Name, field.Name)
}

func toStringBodyFirst(field Field) string {
	if strings.HasSuffix(field.Type, "[]") {
		return fmt.Sprintf("                + \"%s=\" + Arrays.toString(%s)\n", field.Name, field.Name)
	}
	return fmt.Sprintf("                + \"%s=\" + %s\n", field.Name, field.Name)
}

func fields(field Field) string {
	ti := toJavaType(field.Type)
	return fmt.Sprintf("    private %s %s;", ti.FullType, field.Name)
}

func read(field Field) string {
	ti := toJavaType(field.Type)
	if ti.IsArr && ti.IsCustomClass {
		return fmt.Sprintf(`            %s %s = reader.readArrayOfCompact("%s", %s.class);`, ti.FullType, field.Name, field.Name, ti.Type)
	}
	return fmt.Sprintf(`            %s %s = reader.read%s("%s");`, ti.FullType, field.Name, methodName(ti, field.Type), field.Name)
}

func getters(field Field) string {
	ti := toJavaType(field.Type)
	upperName := strings.ToUpper(string(field.Name[0])) + field.Name[1:]
	return fmt.Sprintf(`    public %s get%s() {
        return %s;
    }
`, ti.FullType, upperName, field.Name)
}

func generateImports(clsAndSchema struct {
	Cls    Class
	Schema Schema
}) string {
	var content string
	for _, field := range clsAndSchema.Cls.Fields {
		trimmed := strings.TrimSuffix(field.Type, "[]")
		if field.External || isImportedClass(clsAndSchema.Schema, trimmed) {
			content += fmt.Sprintf("import %s;", trimmed)
		}
	}
	if len(content) > 0 {
		content = "\n" + content + "\n"
	}
	return content
}

func isImportedClass(sch Schema, typ string) bool {
	for _, cls := range sch.Classes {
		fullName := getClassFullName(cls.Name, sch.Namespace)
		if typ == fullName {
			return false
		} else if !strings.Contains(typ, ".") && typ == cls.Name {
			return false
		}
	}
	return !isBuiltInType(typ)
}

func equalsBody(field Field) string {
	ti := toJavaType(field.Type)
	var s string
	if field.Type == "float32" {
		s = fmt.Sprintf("if (Float.compare(%s, that.%s) != 0) return false;", field.Name, field.Name)
	} else if field.Type == "float64" {
		s = fmt.Sprintf("if (Double.compare(%s, that.%s) != 0) return false;", field.Name, field.Name)
	} else if ti.IsArr {
		s = fmt.Sprintf("if (!Arrays.equals(%s, that.%s)) return false;", field.Name, field.Name)
	} else if _, ok := fixedSizeTypes[field.Type]; ok {
		s = fmt.Sprintf("if (%s != that.%s) return false;", field.Name, field.Name)
	} else {
		s = fmt.Sprintf("if (!Objects.equals(%s, that.%s)) return false;", field.Name, field.Name)
	}
	return "        " + s
}
