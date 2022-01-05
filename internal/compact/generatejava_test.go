package compact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	assert.Equal(t, nil, nil)
}

/*
import unittest
from os.path import dirname, join, realpath

from compactc.java.generate import _generate_class, _hint
from compactc.util import load_schemas


class GenerateTest(unittest.TestCase):
    def test_generate(self):
        curr_dir = dirname(realpath(__file__))
        schema_path = join(curr_dir, "..", "exampleSchemas.yaml")

        schemas = load_schemas(schema_path)

        file = open(
            join(
                curr_dir,
                "TypesWithDefaults.java",
            ),
            "r",
        )
        expected_types_with_defaults = file.read()
        file.close()

        file = open(
            join(
                curr_dir,
                "AllTypes.java",
            ),
            "r",
        )
        expected_all_types = file.read()
        file.close()

        self.assertEqual(expected_types_with_defaults, _generate_class(schemas["classes"][0], "package test;\n\n"))
        self.assertEqual(expected_all_types, _generate_class(schemas["classes"][1], "package test;\n\n"))

    def test_hint(self):
        curr_dir = dirname(realpath(__file__))
        schema_path = join(curr_dir, "..", "exampleSchemas.yaml")

        schemas = load_schemas(schema_path)
        self.assertEqual(
            """Programmatic configuration:
compactSerializationConfig(test.TypesWithDefaults.class, "TypesWithDefaults", test.TypesWithDefaults.HZ_COMPACT_SERIALIZER);
compactSerializationConfig(test.AllTypes.class, "AllTypes", test.AllTypes.HZ_COMPACT_SERIALIZER);

Declarative xml configuration:
<compact-serialization>
   <registered-classes>
        <class type-name="TypesWithDefaults" serializer="test.TypesWithDefaults.Serializer">test.TypesWithDefaults</class>
        <class type-name="AllTypes" serializer="test.AllTypes.Serializer">test.AllTypes</class>
   </registered-classes>

Declarative yaml configuration:
compact-serialization:
   registered-classes:
        - class: test.TypesWithDefaults
          type-name: TypesWithDefaults
          serializer: test.TypesWithDefaults.Serializer
        - class: test.AllTypes
          type-name: AllTypes
          serializer: test.AllTypes.Serializer
""",
            _hint(schemas, "test."),
        )


if __name__ == "__main__":
    unittest.main()
*/