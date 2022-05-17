package viewer

import (
	"errors"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/internal/termdbms/database"
)

var (
	serializationErrorString = fmt.Sprintf("Database driver %s does not support serialization.", database.DriverString)
)

func Serialize(m *TuiModel) (string, error) {
	return "", errors.New(serializationErrorString)
}

func SerializeOverwrite(m *TuiModel) error {
	return errors.New(serializationErrorString)
}
