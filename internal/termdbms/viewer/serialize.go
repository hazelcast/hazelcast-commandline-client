package viewer

import (
	"errors"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/termdbms/database"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"
)

var (
	serializationErrorString = fmt.Sprintf("Database driver %s does not support serialization.", database.DriverString)
)

func Serialize(m *TuiModel) (string, error) {
	switch m.Table().Database.(type) {
	case *database.SQLite:
		return SerializeSQLiteDB(m.Table().Database.(*database.SQLite), m), nil
	default:
		return "", errors.New(serializationErrorString)
	}
}

func SerializeOverwrite(m *TuiModel) error {
	t := m.Table()
	switch t.Database.(type) {
	case *database.SQLite:
		SerializeOverwriteSQLiteDB(t.Database.(*database.SQLite), m)
		return nil
	default:
		return errors.New(serializationErrorString)
	}
}

// SQLITE

func SerializeSQLiteDB(db *database.SQLite, m *TuiModel) string {
	db.CloseDatabaseReference()
	source, err := os.ReadFile(db.GetFileName())
	if err != nil {
		panic(err)
	}
	ext := path.Ext(m.InitialFileName)
	newFileName := fmt.Sprintf("%s-%d%s", strings.TrimSuffix(m.InitialFileName, ext), rand.Intn(4), ext)
	err = os.WriteFile(newFileName, source, 0777)
	if err != nil {
		log.Fatal(err)
	}
	db.SetDatabaseReference(db.GetFileName())
	return newFileName
}

func SerializeOverwriteSQLiteDB(db *database.SQLite, m *TuiModel) {
	db.CloseDatabaseReference()
	filename := db.GetFileName()

	source, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(m.InitialFileName, source, 0777)
	if err != nil {
		log.Fatal(err)
	}
	db.SetDatabaseReference(filename)
}
