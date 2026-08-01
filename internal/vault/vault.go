package vault

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	vaultcrypto "keva/internal/crypto"
	"keva/internal/database"
)

var ErrAccountNotFound = errors.New("account not found")

type Store interface {
	CreateAccount(ctx context.Context, record database.AccountRecord) error
	UpdateAccount(ctx context.Context, record database.AccountRecord) error
	DeleteAccount(ctx context.Context, id string) error
	ListAccounts(ctx context.Context) ([]database.AccountRecord, error)
	GetAccount(ctx context.Context, id string) (database.AccountRecord, error)
}

type Vault struct {
	store Store
}

func New(store Store) *Vault {
	return &Vault{store: store}
}

func (v *Vault) CreateAccount(ctx context.Context, dek []byte, input AccountInput) (AccountSummary, error) {
	if err := validateAccountInput(input); err != nil {
		return AccountSummary{}, err
	}

	now := time.Now().Unix()
	detail := AccountDetail{
		ID:        uuid.NewString(),
		Title:     strings.TrimSpace(input.Title),
		Username:  input.Username,
		Password:  input.Password,
		URL:       input.URL,
		Note:      input.Note,
		CreatedAt: now,
		UpdatedAt: now,
	}

	record, err := encryptDetail(dek, detail)
	if err != nil {
		return AccountSummary{}, err
	}
	if err := v.store.CreateAccount(ctx, record); err != nil {
		return AccountSummary{}, err
	}
	return summaryFromDetail(detail), nil
}

func (v *Vault) UpdateAccount(ctx context.Context, dek []byte, id string, input AccountInput) (AccountSummary, error) {
	if id == "" {
		return AccountSummary{}, ErrAccountNotFound
	}
	if err := validateAccountInput(input); err != nil {
		return AccountSummary{}, err
	}

	existing, err := v.GetAccount(ctx, dek, id)
	if err != nil {
		return AccountSummary{}, err
	}

	detail := AccountDetail{
		ID:        id,
		Title:     strings.TrimSpace(input.Title),
		Username:  input.Username,
		Password:  input.Password,
		URL:       input.URL,
		Note:      input.Note,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: time.Now().Unix(),
	}

	record, err := encryptDetail(dek, detail)
	if err != nil {
		return AccountSummary{}, err
	}
	if err := v.store.UpdateAccount(ctx, record); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccountSummary{}, ErrAccountNotFound
		}
		return AccountSummary{}, err
	}
	return summaryFromDetail(detail), nil
}

func (v *Vault) DeleteAccount(ctx context.Context, id string) error {
	if id == "" {
		return ErrAccountNotFound
	}
	if err := v.store.DeleteAccount(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountNotFound
		}
		return err
	}
	return nil
}

func (v *Vault) ListAccounts(ctx context.Context, dek []byte) ([]AccountSummary, error) {
	records, err := v.store.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]AccountSummary, 0, len(records))
	for _, record := range records {
		detail, err := decryptRecord(dek, record)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summaryFromDetail(detail))
	}
	return summaries, nil
}

func (v *Vault) SearchAccounts(ctx context.Context, dek []byte, keyword string) ([]AccountSummary, error) {
	summaries, err := v.ListAccounts(ctx, dek)
	if err != nil {
		return nil, err
	}

	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return summaries, nil
	}

	filtered := make([]AccountSummary, 0)
	for _, summary := range summaries {
		haystack := strings.ToLower(summary.Title + "\n" + summary.Username + "\n" + summary.URL)
		if strings.Contains(haystack, keyword) {
			filtered = append(filtered, summary)
		}
	}
	return filtered, nil
}

func (v *Vault) GetAccount(ctx context.Context, dek []byte, id string) (AccountDetail, error) {
	if id == "" {
		return AccountDetail{}, ErrAccountNotFound
	}
	record, err := v.store.GetAccount(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccountDetail{}, ErrAccountNotFound
		}
		return AccountDetail{}, err
	}
	return decryptRecord(dek, record)
}

func validateAccountInput(input AccountInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return errors.New("title is required")
	}
	if input.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func encryptDetail(dek []byte, detail AccountDetail) (database.AccountRecord, error) {
	title, err := vaultcrypto.Encrypt(dek, []byte(detail.Title))
	if err != nil {
		return database.AccountRecord{}, err
	}
	username, err := vaultcrypto.Encrypt(dek, []byte(detail.Username))
	if err != nil {
		return database.AccountRecord{}, err
	}
	password, err := vaultcrypto.Encrypt(dek, []byte(detail.Password))
	if err != nil {
		return database.AccountRecord{}, err
	}
	url, err := vaultcrypto.Encrypt(dek, []byte(detail.URL))
	if err != nil {
		return database.AccountRecord{}, err
	}
	note, err := vaultcrypto.Encrypt(dek, []byte(detail.Note))
	if err != nil {
		return database.AccountRecord{}, err
	}

	return database.AccountRecord{
		ID:                detail.ID,
		EncryptedTitle:    title,
		EncryptedUsername: username,
		EncryptedPassword: password,
		EncryptedURL:      url,
		EncryptedNote:     note,
		CreatedAt:         detail.CreatedAt,
		UpdatedAt:         detail.UpdatedAt,
	}, nil
}

func decryptRecord(dek []byte, record database.AccountRecord) (AccountDetail, error) {
	title, err := decryptString(dek, record.EncryptedTitle)
	if err != nil {
		return AccountDetail{}, err
	}
	username, err := decryptString(dek, record.EncryptedUsername)
	if err != nil {
		return AccountDetail{}, err
	}
	password, err := decryptString(dek, record.EncryptedPassword)
	if err != nil {
		return AccountDetail{}, err
	}
	url, err := decryptString(dek, record.EncryptedURL)
	if err != nil {
		return AccountDetail{}, err
	}
	note, err := decryptString(dek, record.EncryptedNote)
	if err != nil {
		return AccountDetail{}, err
	}

	return AccountDetail{
		ID:        record.ID,
		Title:     title,
		Username:  username,
		Password:  password,
		URL:       url,
		Note:      note,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
}

func decryptString(dek, sealed []byte) (string, error) {
	if len(sealed) == 0 {
		return "", nil
	}
	plaintext, err := vaultcrypto.Decrypt(dek, sealed)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func summaryFromDetail(detail AccountDetail) AccountSummary {
	return AccountSummary{
		ID:        detail.ID,
		Title:     detail.Title,
		Username:  detail.Username,
		URL:       detail.URL,
		CreatedAt: detail.CreatedAt,
		UpdatedAt: detail.UpdatedAt,
	}
}
