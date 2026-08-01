package vault

import (
	"context"
	"strings"
	"testing"

	"keva/internal/crypto"
	"keva/internal/database"
)

func TestVaultEncryptsCRUDAndSearchesInMemory(t *testing.T) {
	store := newMemoryStore()
	v := New(store)
	dek, err := crypto.GenerateDEK()
	if err != nil {
		t.Fatal(err)
	}

	created, err := v.CreateAccount(context.Background(), dek, AccountInput{
		Title:    "GitHub",
		Username: "octo",
		Password: "plain-secret",
		URL:      "https://github.com",
		Note:     "recovery note",
	})
	if err != nil {
		t.Fatal(err)
	}

	raw := store.records[created.ID]
	rawBytes := string(raw.EncryptedTitle) + string(raw.EncryptedUsername) + string(raw.EncryptedPassword) + string(raw.EncryptedURL) + string(raw.EncryptedNote)
	for _, forbidden := range []string{"GitHub", "octo", "plain-secret", "github.com", "recovery note"} {
		if strings.Contains(rawBytes, forbidden) {
			t.Fatalf("encrypted record contains plaintext %q", forbidden)
		}
	}

	detail, err := v.GetAccount(context.Background(), dek, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Password != "plain-secret" {
		t.Fatalf("got password %q", detail.Password)
	}

	results, err := v.SearchAccounts(context.Background(), dek, "git")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results want 1", len(results))
	}

	updated, err := v.UpdateAccount(context.Background(), dek, created.ID, AccountInput{
		Title:    "Email",
		Username: "me",
		Password: "new-secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Email" {
		t.Fatalf("got title %q", updated.Title)
	}

	if err := v.DeleteAccount(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := v.GetAccount(context.Background(), dek, created.ID); err != ErrAccountNotFound {
		t.Fatalf("got %v want %v", err, ErrAccountNotFound)
	}
}

type memoryStore struct {
	records map[string]database.AccountRecord
}

func newMemoryStore() *memoryStore {
	return &memoryStore{records: map[string]database.AccountRecord{}}
}

func (s *memoryStore) CreateAccount(ctx context.Context, record database.AccountRecord) error {
	s.records[record.ID] = record
	return nil
}

func (s *memoryStore) UpdateAccount(ctx context.Context, record database.AccountRecord) error {
	if _, ok := s.records[record.ID]; !ok {
		return ErrAccountNotFound
	}
	s.records[record.ID] = record
	return nil
}

func (s *memoryStore) DeleteAccount(ctx context.Context, id string) error {
	if _, ok := s.records[id]; !ok {
		return ErrAccountNotFound
	}
	delete(s.records, id)
	return nil
}

func (s *memoryStore) ListAccounts(ctx context.Context) ([]database.AccountRecord, error) {
	records := make([]database.AccountRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}
	return records, nil
}

func (s *memoryStore) GetAccount(ctx context.Context, id string) (database.AccountRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return database.AccountRecord{}, ErrAccountNotFound
	}
	return record, nil
}
