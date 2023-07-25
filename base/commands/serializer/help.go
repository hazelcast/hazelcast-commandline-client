package serializer

import (
	"fmt"
	"strings"
)

func printFurtherToDoInfo(lang string, classes map[ClassInfo]string) {
	fmt.Printf("WARNING: Don't forget to register your generated serializers to Hazelcast configuration.\n\n")
	printProgrammaticConfig(lang, classes)
	printXMLConfig(lang, classes)
	printYamlConfig(lang, classes)
}

func printProgrammaticConfig(lang string, classes map[ClassInfo]string) {
	printHeader("Programmatic configuration")
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
	fmt.Println(sb.String())
}

func printXMLConfig(lang string, classes map[ClassInfo]string) {
	printHeader("Declarative XML configuration")
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
	fmt.Println(sb.String())
}

func printYamlConfig(lang string, classes map[ClassInfo]string) {
	printHeader("Declarative YAML configuration")
	var sb strings.Builder
	if lang == "java" {
		sb.WriteString("serialization:\n")
		sb.WriteString("\tcompact-serialization:\n")
		sb.WriteString("\t\tserializers:\n")
		for k := range classes {
			sb.WriteString(fmt.Sprintf("\t\t\t- serializer: %s\n", getClassFullName(k.ClassName, k.Namespace)))
		}
	}
	fmt.Println(sb.String())
}

func printHeader(header string) {
	fmt.Printf("---------%s---------\n\n", header)
}
