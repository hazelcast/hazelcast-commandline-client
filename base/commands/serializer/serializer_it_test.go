//go:build std || serializer

package serializer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/assert"
)

const generatedTestFilesDirectoryName = "generatedTestFiles"

var (
	allTypes,
	nestedCompact,
	allTypesSchema,
	allTypesSchemaPath,
	noNamespace,
	noNamespaceNested,
	noNamespaceSchemaPath,
	noNamespaceSchema,
	external,
	externalSchemaPath,
	externalSchema,
	student,
	classroom,
	school,
	schoolSchemaPath,
	schoolSchema,
	example1,
	example2,
	example3,
	internalClassReferenceSchemaPath,
	internalClassReference string
)

func init() {
	UUIDGenFunc = func() types.UUID {
		return types.NewUUIDWith(10, 10)
	}
	s := types.NewUUIDWith(10, 10).String()
	fmt.Println(s)
	generationTestFilesDir := filepath.Join("testdata", "generationTestFiles")
	generationTestFilesSchemaDir := filepath.Join("testdata", "generationTestFiles", "schema")

	// All types test files
	allTypes = readTestFile(generationTestFilesDir, "AllTypes.java")
	nestedCompact = readTestFile(generationTestFilesDir, "NestedCompact.java")
	allTypesSchemaPath = filepath.Join(generationTestFilesSchemaDir, "allTypes.yaml")
	allTypesSchema = readTestFile(generationTestFilesSchemaDir, "allTypes.yaml")

	// No namespace test files
	noNamespace = readTestFile(generationTestFilesDir, "NoNamespace.java")
	noNamespaceNested = readTestFile(generationTestFilesDir, "NoNamespaceNested.java")
	noNamespaceSchemaPath = filepath.Join(generationTestFilesSchemaDir, "noNamespace.yaml")
	noNamespaceSchema = readTestFile(generationTestFilesSchemaDir, "noNamespace.yaml")

	// External test files
	external = readTestFile(generationTestFilesDir, "External.java")
	externalSchemaPath = filepath.Join(generationTestFilesSchemaDir, "external.yaml")
	externalSchema = readTestFile(generationTestFilesSchemaDir, "external.yaml")

	// Nested import test files
	student = readTestFile(generationTestFilesDir, "Student.java")

	classroom = readTestFile(generationTestFilesDir, "Classroom.java")

	school = readTestFile(generationTestFilesDir, "School.java")
	schoolSchemaPath = filepath.Join(generationTestFilesSchemaDir, "school.yaml")
	schoolSchema = readTestFile(generationTestFilesSchemaDir, "school.yaml")

	// Internal class reference test files
	example1 = readTestFile(generationTestFilesDir, "Example1.java")
	example2 = readTestFile(generationTestFilesDir, "Example2.java")
	example3 = readTestFile(generationTestFilesDir, "Example3.java")
	internalClassReferenceSchemaPath = filepath.Join(generationTestFilesSchemaDir, "internalClassReference.yaml")
	internalClassReference = readTestFile(generationTestFilesSchemaDir, "internalClassReference.yaml")
}

func readTestFile(dir string, fName string) string {
	f, err := os.ReadFile(filepath.Join(dir, fName))
	if err != nil {
		panic(err)
	}
	return string(f)
}

func makeKey(className, fileName string) ClassInfo {
	return ClassInfo{
		Namespace: "",
		FileName:  fileName,
		ClassName: className,
	}
}

func TestGenerate(t *testing.T) {
	tcs := []struct {
		expected          map[ClassInfo]string
		name              string
		lang              string
		compactSchema     string
		compactSchemaPath string
	}{
		{
			name: "AllTypes",
			lang: "java",
			expected: map[ClassInfo]string{
				makeKey("AllTypes", "AllTypes.java"):           allTypes,
				makeKey("NestedCompact", "NestedCompact.java"): nestedCompact,
			},
			compactSchema:     allTypesSchema,
			compactSchemaPath: allTypesSchemaPath,
		},
		{
			name: "NoNamespace",
			lang: "java",
			expected: map[ClassInfo]string{
				makeKey("NoNamespace", "NoNamespace.java"):             noNamespace,
				makeKey("NoNamespaceNested", "NoNamespaceNested.java"): noNamespaceNested,
			},
			compactSchema:     noNamespaceSchema,
			compactSchemaPath: noNamespaceSchemaPath,
		},
		{
			name: "External",
			lang: "java",
			expected: map[ClassInfo]string{
				makeKey("External", "External.java"): external,
			},
			compactSchema:     externalSchema,
			compactSchemaPath: externalSchemaPath,
		},
		{
			name: "NestedImport",
			lang: "java",
			expected: map[ClassInfo]string{
				makeKey("Student", "Student.java"):     student,
				makeKey("Classroom", "Classroom.java"): classroom,
				makeKey("School", "School.java"):       school,
			},
			compactSchema:     schoolSchema,
			compactSchemaPath: schoolSchemaPath,
		},
		{
			name: "InternalClassReference",
			lang: "java",
			expected: map[ClassInfo]string{
				makeKey("Example1", "Example1.java"): example1,
				makeKey("Example2", "Example2.java"): example2,
				makeKey("Example3", "Example3.java"): example3,
			},
			compactSchema:     internalClassReference,
			compactSchemaPath: internalClassReferenceSchemaPath,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			sch, err := parseSchema([]byte(tc.compactSchema))
			if err != nil {
				t.Fatal(err)
			}
			err = processSchema(filepath.Dir(tc.compactSchemaPath), &sch)
			if err != nil {
				t.Fatal(err)
			}
			classes, err := generateCompactClasses(tc.lang, sch)
			if err != nil {
				t.Fatal(err)
			}

			nc := map[ClassInfo]string{}

			for k, v := range classes {
				// We don't care about the namespace for the test and we need this modification for the assertion below to pass
				ci := ClassInfo{
					Namespace: "",
					FileName:  k.FileName,
					ClassName: k.ClassName,
				}
				nc[ci] = v
			}

			isEqual := assert.Equal(t, tc.expected, nc)

			if !isEqual {
				for k, v := range classes {
					err := os.MkdirAll(generatedTestFilesDirectoryName, 0770)
					if err != nil {
						t.Fatal(err)
					}
					f, err := os.Create(filepath.Join(generatedTestFilesDirectoryName, k.FileName))
					if err != nil {
						t.Fatal(err)
					}
					defer f.Close()
					f.WriteString(v)
				}
				t.Fatalf("The generated classes was not equal for test %s and language %s", tc.name, tc.lang)
			}
		})
	}
}
