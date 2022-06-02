package browser

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
)

func execSQL(c *hazelcast.Client, text string) (*sql.Rows, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	lt := strings.ToLower(text)
	if strings.HasPrefix(lt, "select") || strings.HasPrefix(lt, "show") {
		return query(c, text)
	}
	return nil, exec(c, text)
}

func query(c *hazelcast.Client, text string) (*sql.Rows, error) {
	rows, err := c.QuerySQL(context.Background(), text)
	if err != nil {
		return rows, fmt.Errorf("querying: %w", err)
	}
	return rows, nil
}

func exec(db *hazelcast.Client, text string) error {
	r, err := db.ExecSQL(context.Background(), text)
	if err != nil {
		return fmt.Errorf("executing: %w", err)
	}
	ra, err := r.RowsAffected()
	if err != nil {
		return err
	}
	return fmt.Errorf("---\nAffected rows: %d\n\n", ra)
}
