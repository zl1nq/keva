package database

type AccountRecord struct {
	ID                string
	EncryptedTitle    []byte
	EncryptedUsername []byte
	EncryptedPassword []byte
	EncryptedURL      []byte
	EncryptedNote     []byte
	CreatedAt         int64
	UpdatedAt         int64
}
