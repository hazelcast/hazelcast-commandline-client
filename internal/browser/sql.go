package browser

import (
	"database/sql"
	"fmt"
	"strings"
)

func execSQL(db *sql.DB, text string) (*sql.Rows, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	lt := strings.ToLower(text)
	if strings.HasPrefix(lt, "select") || strings.HasPrefix(lt, "show") {
		return query(db, text)
	}
	return nil, exec(db, text)
}

func query(db *sql.DB, text string) (*sql.Rows, error) {
	rows, err := db.Query(text)
	if err != nil {
		return rows, fmt.Errorf("querying: %w", err)
	}
	return rows, nil
}

func exec(db *sql.DB, text string) error {
	r, err := db.Exec(text)
	if err != nil {
		return fmt.Errorf("executing: %w", err)
	}
	ra, err := r.RowsAffected()
	if err != nil {
		return err
	}
	return fmt.Errorf("---\nAffected rows: %d\n\n", ra)
}
