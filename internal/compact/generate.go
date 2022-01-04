package compact

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Field struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Default string `yaml:"default"`
}

type Class struct {
	Name   string `yaml:"name"`
	Fields []Field
}
type Schemas struct {
	Classes []Class
}

var languageGenerators = map[string]func(schemas Schemas, namespace string, outputDir string) (string, error){
	"java":   JavaGenerate,
	"cpp":    nil,
	"csharp": nil,
}

func Generate(language string, schemaFilePath string, outputDir string, namespace string) error {
	yamlFile, err := ioutil.ReadFile(schemaFilePath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	schemas := Schemas{}
	err = yaml.Unmarshal(yamlFile, &schemas)
	if err != nil {
		log.Fatalf("File Format Error: %v", err)
	}
	err = Validate(schemas)
	if err != nil {
		log.Fatalf("File Format Error: %v", err)
	}
	generator := languageGenerators[language]
	hint, err := generator(schemas, outputDir, namespace)
	//TODO add silent
	log.Printf(hint)
	return err
}

func Validate(schemas Schemas) error {
	return nil
}

/*


builtin_types = {
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

def valid(schemas) -> Optional[str]:
    curr_dir = dirname(realpath(__file__))
    schema_path = join(curr_dir, "validate", "validate.json")
    with open(schema_path, "r") as schema_file:
        schema = json.load(schema_file)
    try:
        jsonschema.validate(schemas, schema)
    except jsonschema.ValidationError as e:
        return e.__str__()

    compacts_by_name = {}
    # detect duplicate classes
    for schema in schemas["classes"]:
        if schema["name"] in compacts_by_name.keys():
            return "A compact with name %s already exists" % schema["name"]
        else:
            compacts_by_name[schema["name"]] = {}

    # detect if all field types are valid
    for item in schemas["classes"]:
        for field in item["fields"]:
            if field["type"] in builtin_types:
                continue
            if field["type"] in compacts_by_name:
                continue

            if field["type"].endswith("[]"):
                component_type = field["type"][0 : len(field["type"]) - 2]
                if component_type in builtin_types:
                    continue
                if component_type in compacts_by_name:
                    continue
            return 'Validation error. field type "%s" is not one of the builtin types and not defined' % field["type"]

    return None

*/
