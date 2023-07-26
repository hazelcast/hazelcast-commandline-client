package serializer

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateJSONSchemaString(t *testing.T) {
	tcs := []struct {
		name        string
		schema      string
		isErr       bool
		noSchemaErr bool
		errString   string
	}{
		{
			name:   "non-json string",
			schema: "",
			isErr:  true,
		},
		{
			name: "valid",
			schema: `{ 
				  "classes":[
					 {
						"name":"Employee",
						"fields":[
						   {
							  "name":"age",
							  "type":"int32[]"
						   },
						   {
							  "name":"name",
							  "type":"string"
						   },
						   {
							  "name":"id",
							  "type":"int64"
						   }
						]
					 }
				  ]
			   }`,
			noSchemaErr: true,
		},
		{
			name:        "valid custom field type defined in schema",
			schema:      ValidNewFieldTypeDefined,
			noSchemaErr: true,
		},
		{
			name:      "mandatory class field is missing",
			schema:    "{}",
			errString: "classes is required",
		},
		{
			name: "mandatory 'fields' field of class is missing",
			schema: `{ "classes":[
                     {
                        "name":"Employee"
                     }
                  ]
           }`,
			errString: "fields is required",
		},
		{
			name: "mandatory 'name' field in 'fields' field is missing",
			schema: `{ "classes":[
                     {
                        "name":"Employee",
                        "fields":[
                           {
                              "type":"Work"
                           }
                        ]
                     }
                  ]
           }`,
			errString: "name is required",
		},
		{
			name: "mandatory 'type' field in 'fields' field is missing",
			schema: `{ "classes":[
                     {
                        "name":"Employee",
                        "fields":[
                           {
                              "name":"age"
                           }
                        ]
                     }
                  ]
           }`,
			errString: "type is required",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			valid, errors, err := validateJSONSchemaString(tc.schema)
			assert.Equal(t, tc.isErr, err != nil)
			if tc.isErr {
				return
			}
			assert.Nil(t, err)
			if tc.noSchemaErr {
				assert.Empty(t, errors)
				assert.True(t, valid)
				return
			}
			assert.False(t, valid)
			assert.Contains(t, strings.Join(errors, ","), tc.errString)
		})
	}
}

func TestValidateSchemaSemantics(t *testing.T) {
	tcs := []struct {
		name               string
		schema             string
		isErr              bool
		err                string
		expectedClassNames map[string]Class
	}{
		{
			name:   "valid field type defined in schema",
			schema: ValidNewFieldTypeDefined,
			expectedClassNames: map[string]Class{
				"Employee": {
					Name: "Employee",
					Fields: []Field{
						{Name: "age", Type: "Work"},
					},
					Namespace: "",
				},
				"Work": {
					Name:      "Work",
					Fields:    []Field{},
					Namespace: "",
				},
			},
		},
		{
			name: "namespace should apply to all classes defined in the same schema",
			schema: `{
				"namespace": "com.respectme",
				"classes":[
				   {
					  "name":"Employee",
					  "fields":[
						 {
							"name":"age",
							"type":"int16"
						 }
					  ]
				   },
				   {
					  "name":"Student",
					  "fields":[
						{
							"name":"number",
							"type":"int32"
						}
					  ]
				   }
				]
		    }`,
			expectedClassNames: map[string]Class{
				"com.respectme.Employee": {
					Name: "Employee",
					Fields: []Field{
						{Name: "age", Type: "int16"},
					},
					Namespace: "com.respectme",
				},
				"com.respectme.Student": {
					Name: "Student",
					Fields: []Field{
						{Name: "number", Type: "int32"},
					},
					Namespace: "com.respectme",
				},
			},
		},
		{
			name: "invalid field type",
			// Field is not defined nor imported nor external
			schema: `{
				  "classes":[
					 {
						"name":"Employee",
						"fields":[
						   {
							  "name":"age",
							  "type":"Work"
						   }
						]
					 }
				  ]
			}`,
			isErr: true,
			err:   "not one of the builtin types or not defined",
		},
		{
			name: "duplicate compact class",
			schema: `{
		          "classes":[
		             {
		                "name":"Employee",
		                "fields":[
		                   {
		                      "name":"age",
		                      "type":"Work"
		                   }
		                ]
		             },
		             {
		                "name":"Employee",
		                "fields":[]
		             }
		          ]
		    }`,
			isErr: true,
			err:   "already exist",
		},
		{
			name: "duplicate field name same field type",
			schema: `{
		          "classes":[
		             {
		                "name":"Employee",
		                "fields":[
		                   {
		                      "name":"age",
		                      "type":"int8"
		                   },
		                   {
		                      "name":"age",
		                      "type":"int8"
		                   }
		                ]
		             }
		          ]
		    }`,
			isErr: true,
			err:   "field is defined more than once in class",
		},
		{
			name: "duplicate field name different field type",
			schema: `{
		          "classes":[
		             {
		                "name":"Employee",
		                "fields":[
		                   {
		                      "name":"age",
		                      "type":"int8"
		                   },
		                   {
		                      "name":"age",
		                      "type":"string"
		                   }
		                ]
		             }
		          ]
		    }`,
			isErr: true,
			err:   "field is defined more than once in class",
		},
		{
			name: "valid array field type",
			schema: `{
		          "classes":[
		             {
		                "name":"Employee",
		                "fields":[
		                   {
		                      "name":"age",
		                      "type":"nullableInt16[]"
		                   }
		                ]
		             }
		          ]
		    }`,
			expectedClassNames: map[string]Class{
				"Employee": {
					Name: "Employee",
					Fields: []Field{
						{Name: "age", Type: "nullableInt16[]"},
					},
					Namespace: "",
				},
			},
		},
		{
			name: "invalid array field type",
			schema: `{
		          "classes":[
		             {
		                "name":"Employee",
		                "fields":[
		                   {
		                      "name":"age",
		                      "type":"[]"
		                   }
		                ]
		             }
		          ]
		    }`,
			isErr: true,
			err:   "not one of the builtin types or not defined",
		},
		{
			name: "can import another yaml file and use its class",
			schema: `{
				"imports":["testdata/xyz.yaml"],
				"classes":[
					{
					"name":"Employee",
					"fields":[
						{
							"name":"age",
							"type":"com.xyz.Work"
						}
					]
					}
				]
			}`,
			expectedClassNames: map[string]Class{
				"Employee": {
					Name: "Employee",
					Fields: []Field{
						{Name: "age", Type: "com.xyz.Work"},
					},
					Namespace: "",
				},
				"com.xyz.Work": {
					Name: "Work",
					Fields: []Field{
						{Name: "name", Type: "string"},
					},
					Namespace: "com.xyz",
				},
			},
		},
		{
			name: "can import multiple other yaml files and use types with the same class name",
			schema: `{
				"imports":["testdata/xyz.yaml", "testdata/zyx.yaml"],
				"classes":[
					{
					"name":"Employee",
					"fields":[
						{
							"name":"work",
							"type":"com.xyz.Work"
						},
						{
							"name":"work2",
							"type":"com.zyx.Work"
						}
					]
					}
				]
			}`,
			expectedClassNames: map[string]Class{
				"Employee": {
					Name: "Employee",
					Fields: []Field{
						{Name: "work", Type: "com.xyz.Work"},
						{Name: "work2", Type: "com.zyx.Work"},
					},
					Namespace: "",
				},
				"com.xyz.Work": {
					Name: "Work",
					Fields: []Field{
						{Name: "name", Type: "string"},
					},
					Namespace: "com.xyz",
				},
				"com.zyx.Work": {
					Name: "Work",
					Fields: []Field{
						{Name: "name2", Type: "string"},
					},
					Namespace: "com.zyx",
				},
			},
		},
		{
			name: "Defining a class as external makes it usable even when not imported",
			// Adding external makes this work
			schema: `{
				"classes":[
					{
					"name":"Employee",
					"fields":[
						{
							"name":"age",
							"type":"Work",
							"external":true
						}
					]
					}
				]
			}`,
			expectedClassNames: map[string]Class{
				"Employee": {
					Name: "Employee",
					Fields: []Field{
						{Name: "age", Type: "Work", External: true},
					},
					Namespace: "",
				},
			},
		},
		{
			name: "defining and importing the same class causes error",
			// In this, imported class and defined class are exactly the same.
			schema: `{
				  "imports":["testdata/example.yml"],
				  "classes":[
					 {
						"name":"Employee",
						"fields":[
						   {
							  "name":"name",
							  "type":"string"
						   }
						]
					 }
				  ]
			   }`,
			isErr: true,
			err:   "class defined more than once",
		},
		{
			name: "importing and defining the same named and namespaced class causes error",
			// In this, the type is different.
			schema: `{
				  "imports":["testdata/example.yml"],
				  "classes":[
					 {
						"name":"Employee",
						"fields":[
						   {
							  "name":"work",
							  "type":"string"
						   }
						]
					 }
				  ]
			}`,
			isErr: true,
			err:   "class defined more than once",
		},
		{
			name: "already imported files should be skipped",
			schema: `{
				  "imports":["testdata/nestedImport1.yml"], 
				  "classes":[]
			   }`,
			isErr: false,
			expectedClassNames: map[string]Class{
				"com.nestedImport1.Foo": {
					Name: "Foo",
					Fields: []Field{
						{Name: "fieldFoo", Type: "string"},
					},
					Namespace: "com.nestedImport1",
				},
				"com.nestedImport2.Bar": {
					Name: "Bar",
					Fields: []Field{
						{Name: "fieldBar", Type: "string"},
					},
					Namespace: "com.nestedImport2",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var schema map[string]interface{}
			require.Nil(t, json.Unmarshal([]byte(tc.schema), &schema))
			sch, err := convertMapToSchema(schema)
			require.Nil(t, err)
			workingDir, err := os.Getwd()
			require.Nil(t, err)
			err = processSchema(workingDir, &sch)
			if !tc.isErr {
				assert.Nil(t, err)
				require.Equal(t, tc.expectedClassNames, sch.ClassNames)
				return
			}
			if err == nil {
				t.Fatal("expected error but got none")
			}
			assert.Contains(t, err.Error(), tc.err)
		})
	}
}

const (
	ValidNewFieldTypeDefined = `{
      "classes":[
         {
            "name":"Employee",
            "fields":[
               {
                  "name":"age",
                  "type":"Work"
               }
            ]
         },
         {
            "name":"Work",
            "fields":[]
         }
      ]
   }`
)
