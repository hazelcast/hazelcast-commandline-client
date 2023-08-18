package serializer

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func printFurtherToDoInfo(ec plug.ExecContext, lang string, classes map[ClassInfo]string) {
	printProgrammaticConfig(ec, lang, classes)
	printXMLConfig(ec, lang, classes)
	printYamlConfig(ec, lang, classes)
}

func printProgrammaticConfig(ec plug.ExecContext, lang string, classes map[ClassInfo]string) {
	printHeader(ec, "Programmatic configuration")
	var sb strings.Builder
	if lang == "java" {
		sb.WriteString("compactSerializationConfig.setSerializers(\n")
		n := len(classes)
		c := 0
		for k := range classes {
			if c == n-1 {
				sb.WriteString(fmt.Sprintf("\tnew %s.Serializer()\n);\n", getClassFullName(k.ClassName, k.Namespace)))
				break
			}
			sb.WriteString(fmt.Sprintf("\tnew %s.Serializer(),\n", getClassFullName(k.ClassName, k.Namespace)))
			c++
		}
	}
	ec.PrintlnUnnecessary(sb.String())
}

func printXMLConfig(ec plug.ExecContext, lang string, classes map[ClassInfo]string) {
	if lang != langJava {
		return
	}
	printHeader(ec, "Declarative XML configuration")
	var sb strings.Builder
	if lang == "java" {
		sb.WriteString("<serialization>\n")
		sb.WriteString("\t<compact-serialization>\n")
		sb.WriteString("\t\t<serializers>\n")
		for k := range classes {
			sb.WriteString("\t\t\t<serializer>\n")
			sb.WriteString(fmt.Sprintf("\t\t\t\t%s\n", getClassFullName(k.ClassName, k.Namespace)))
			sb.WriteString("\t\t\t</serializer>\n")
		}
		sb.WriteString("\t\t</serializers>\n")
		sb.WriteString("\t</compact-serialization>\n")
		sb.WriteString("</serialization>\n")
	}
	ec.PrintlnUnnecessary(sb.String())
}

func printYamlConfig(ec plug.ExecContext, lang string, classes map[ClassInfo]string) {
	printHeader(ec, "Declarative YAML configuration")
	var sb strings.Builder
	if lang == "java" {
		sb.WriteString("serialization:\n")
		sb.WriteString("\tcompact-serialization:\n")
		sb.WriteString("\t\tserializers:\n")
		for k := range classes {
			sb.WriteString(fmt.Sprintf("\t\t\t- serializer: %s\n", getClassFullName(k.ClassName, k.Namespace)))
		}
	}
	ec.PrintlnUnnecessary(sb.String())
}

func printHeader(ec plug.ExecContext, header string) {
	ec.PrintlnUnnecessary(fmt.Sprintf("---------%s---------\n\n", header))
}
