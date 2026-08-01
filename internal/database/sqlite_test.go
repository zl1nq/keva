package database

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestAccountCRUD(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "keva.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	record := AccountRecord{
		ID:                "id-1",
		EncryptedTitle:    []byte("encrypted title"),
		EncryptedUsername: []byte("encrypted username"),
		EncryptedPassword: []byte("encrypted password"),
		EncryptedURL:      []byte("encrypted url"),
		EncryptedNote:     []byte("encrypted note"),
		CreatedAt:         1,
		UpdatedAt:         1,
	}

	if err := db.CreateAccount(ctx, record); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetAccount(ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.EncryptedTitle) != "encrypted title" {
		t.Fatalf("got title %q", got.EncryptedTitle)
	}

	record.EncryptedTitle = []byte("updated")
	record.UpdatedAt = 2
	if err := db.UpdateAccount(ctx, record); err != nil {
		t.Fatal(err)
	}

	list, err := db.ListAccounts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || string(list[0].EncryptedTitle) != "updated" {
		t.Fatalf("unexpected list: %#v", list)
	}

	if err := db.DeleteAccount(ctx, record.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetAccount(ctx, record.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("got %v want %v", err, sql.ErrNoRows)
	}
}
