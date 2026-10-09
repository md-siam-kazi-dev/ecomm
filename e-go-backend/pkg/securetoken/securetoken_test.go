package securetoken

import "testing"

func TestGenerate_UniqueAndHashable(t *testing.T) {
	raw, hash, err := Generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if raw == "" || hash == "" {
		t.Fatal("token and hash must be non-empty")
	}
	if len(hash) != 64 {
		t.Fatalf("expected 64-char hex hash, got %d", len(hash))
	}
	if Hash(raw) != hash {
		t.Fatal("Hash must be deterministic for the same raw token")
	}

	raw2, hash2, err := Generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if raw == raw2 || hash == hash2 {
		t.Fatal("generated tokens must be unique")
	}
}
