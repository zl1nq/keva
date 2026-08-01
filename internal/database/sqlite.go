package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)

	db := &DB{conn: conn}
	if err := db.Migrate(context.Background()); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error {
	if db == nil || db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

func (db *DB) Migrate(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    encrypted_title BLOB NOT NULL,
    encrypted_username BLOB,
    encrypted_password BLOB NOT NULL,
    encrypted_url BLOB,
    encrypted_note BLOB,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

INSERT OR IGNORE INTO schema_migrations (version, applied_at) VALUES (1, ?);
`, time.Now().Unix())
	return err
}

func (db *DB) CreateAccount(ctx context.Context, record AccountRecord) error {
	_, err := db.conn.ExecContext(ctx, `
INSERT INTO accounts (
    id,
    encrypted_title,
    encrypted_username,
    encrypted_password,
    encrypted_url,
    encrypted_note,
    created_at,
    updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
`,
		record.ID,
		record.EncryptedTitle,
		record.EncryptedUsername,
		record.EncryptedPassword,
		record.EncryptedURL,
		record.EncryptedNote,
		record.CreatedAt,
		record.UpdatedAt,
	)
	return err
}

func (db *DB) UpdateAccount(ctx context.Context, record AccountRecord) error {
	result, err := db.conn.ExecContext(ctx, `
UPDATE accounts
SET encrypted_title = ?,
    encrypted_username = ?,
    encrypted_password = ?,
    encrypted_url = ?,
    encrypted_note = ?,
    updated_at = ?
WHERE id = ?;
`,
		record.EncryptedTitle,
		record.EncryptedUsername,
		record.EncryptedPassword,
		record.EncryptedURL,
		record.EncryptedNote,
		record.UpdatedAt,
		record.ID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (db *DB) DeleteAccount(ctx context.Context, id string) error {
	result, err := db.conn.ExecContext(ctx, `DELETE FROM accounts WHERE id = ?;`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (db *DB) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	rows, err := db.conn.QueryContext(ctx, `
SELECT id, encrypted_title, encrypted_username, encrypted_password, encrypted_url, encrypted_note, created_at, updated_at
FROM accounts
ORDER BY updated_at DESC;
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []AccountRecord
	for rows.Next() {
		record, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (db *DB) GetAccount(ctx context.Context, id string) (AccountRecord, error) {
	row := db.conn.QueryRowContext(ctx, `
SELECT id, encrypted_title, encrypted_username, encrypted_password, encrypted_url, encrypted_note, created_at, updated_at
FROM accounts
WHERE id = ?;
`, id)
	return scanAccount(row)
}

type accountScanner interface {
	Scan(dest ...any) error
}

func scanAccount(scanner accountScanner) (AccountRecord, error) {
	var record AccountRecord
	if err := scanner.Scan(
		&record.ID,
		&record.EncryptedTitle,
		&record.EncryptedUsername,
		&record.EncryptedPassword,
		&record.EncryptedURL,
		&record.EncryptedNote,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccountRecord{}, sql.ErrNoRows
		}
		return AccountRecord{}, err
	}
	return record, nil
}
