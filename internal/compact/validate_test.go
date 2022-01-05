package compact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	schemas := Schemas{Classes: []Class{{Name: "Employee",
		Fields: []Field{
			{"age", "int32[]", ""},
			{"name", "string", ""},
			{"id", "int64", "1231231"},
		}}}}
	assert.NoError(t, Validate(schemas))
}
func TestInvalidFieldType(t *testing.T) {
	schemas := Schemas{Classes: []Class{{Name: "Employee",
		Fields: []Field{
			{"age", "Work", ""},
		}}}}
	assert.EqualError(t, Validate(schemas), "field type Work is not one of the builtin types and or defined")
}

func TestValid_NewFieldTypeIsDefinedInSchema(t *testing.T) {
	schemas := Schemas{Classes: []Class{
		{Name: "Employee",
			Fields: []Field{
				{"age", "Work", ""}, {"ages", "Work[]", ""},
			}},
		{Name: "Work",
			Fields: []Field{}},
	}}
	assert.NoError(t, Validate(schemas))
}

func TestDuplicateClassNames(t *testing.T) {
	schemas := Schemas{Classes: []Class{{Name: "Employee",
		Fields: []Field{
			{"age", "int32[]", ""},
			{"name", "string", ""},
			{"id", "int64", "1231231"},
		}}, {Name: "Employee",
		Fields: []Field{
			{"age", "int32[]", ""},
			{"name", "string", ""},
			{"id", "int64", "1231231"},
		}}}}
	assert.EqualError(t, Validate(schemas), "a compact with name Employee already exists")
}

func TestDuplicateFieldNames(t *testing.T) {
	schemas := Schemas{Classes: []Class{{Name: "Employee",
		Fields: []Field{
			{"age", "int32[]", ""},
			{"age", "string", ""},
			{"id", "int64", "1231231"},
		}}}}
	assert.EqualError(t, Validate(schemas), "a field with name age already exists in Employee")
}

func TestDefaultSectionOnlyForPrimitives(t *testing.T) {
	schemas := Schemas{Classes: []Class{{Name: "Employee",
		Fields: []Field{
			{"age", "int32[]", "313"},
			{"id", "int64", "1231231"},
		}}}}
	assert.EqualError(t, Validate(schemas), "default section is not allowed for field type int32[]")
	schemas = Schemas{Classes: []Class{
		{Name: "Employee",
			Fields: []Field{
				{"age", "Work", ""}, {"ages", "Work", "Work{}"},
			}},
		{Name: "Work",
			Fields: []Field{}},
	}}
	assert.EqualError(t, Validate(schemas), "default section is not allowed for field type Work")

}
