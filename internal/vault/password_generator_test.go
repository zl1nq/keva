package vault

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	password, err := GeneratePassword(DefaultPasswordOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(password) != DefaultPasswordOptions().Length {
		t.Fatalf("got length %d", len(password))
	}

	for name, charset := range map[string]string{
		"uppercase": "ABCDEFGHJKLMNPQRSTUVWXYZ",
		"lowercase": "abcdefghijkmnopqrstuvwxyz",
		"number":    "23456789",
		"symbol":    "!@#$%^&*()-_=+[]{};:,.?",
	} {
		if !strings.ContainsAny(password, charset) {
			t.Errorf("generated password is missing %s characters", name)
		}
	}
}
