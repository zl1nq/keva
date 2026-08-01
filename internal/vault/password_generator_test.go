package vault

import "testing"

func TestGeneratePassword(t *testing.T) {
	password, err := GeneratePassword(DefaultPasswordOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(password) != DefaultPasswordOptions().Length {
		t.Fatalf("got length %d", len(password))
	}
}
