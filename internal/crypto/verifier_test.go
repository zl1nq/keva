package crypto

import "testing"

func TestPasswordVerifier(t *testing.T) {
	kek, err := GenerateDEK()
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := CreatePasswordVerifier(kek)
	if err != nil {
		t.Fatal(err)
	}

	if err := VerifyPassword(kek, verifier); err != nil {
		t.Fatalf("expected verifier to pass: %v", err)
	}

	wrong, err := GenerateDEK()
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyPassword(wrong, verifier); err == nil {
		t.Fatal("expected wrong key to fail verifier")
	}
}
