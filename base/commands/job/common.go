package job

import (
	"fmt"
	"strconv"
	"strings"
)

func idToString(id int64) string {
	buf := []byte("0000-0000-0000-0000")
	hex := []byte(strconv.FormatInt(id, 16))
	j := 18
	for i := len(hex) - 1; i >= 0; i-- {
		buf[j] = hex[i]
		if j == 15 || j == 10 || j == 5 {
			j--
		}
		j--
	}
	return string(buf[:])
}

func stringToID(s string) (int64, error) {
	// first try whether it's an int
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// otherwise this can be an ID
		s = strings.Replace(s, "-", "", -1)
		i, err = strconv.ParseInt(s, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid ID: %s: %w", s, err)
		}
	}
	return i, nil
}
