package agent

import (
	"crypto/ecdsa"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func privKeyHex(k *ecdsa.PrivateKey) string {
	return hex.EncodeToString(crypto.FromECDSA(k))
}

func TestGenerateLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.key")

	a, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if a.Address == "" || a.PrivateKey() == nil {
		t.Fatalf("Generate returned empty agent")
	}
	if err := a.Save(path, false); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %o, want 0600", info.Mode().Perm())
	}

	b, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Address != a.Address {
		t.Errorf("address mismatch: %s vs %s", b.Address, a.Address)
	}
}

func TestSave_RefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.key")
	a, _ := Generate()
	if err := a.Save(path, false); err != nil {
		t.Fatal(err)
	}
	b, _ := Generate()
	err := b.Save(path, false)
	if err == nil {
		t.Errorf("Save: expected refusal when file exists, got nil")
	}
	if err := b.Save(path, true); err != nil {
		t.Errorf("Save with overwrite=true: %v", err)
	}
	loaded, _ := Load(path)
	if loaded.Address != b.Address {
		t.Errorf("after overwrite, loaded != b")
	}
}

func TestLoad_RejectsInsecureMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.key")
	a, _ := Generate()
	_ = a.Save(path, false)
	// Make it world-readable.
	_ = os.Chmod(path, 0o644)
	if _, err := Load(path); err == nil {
		t.Errorf("Load did NOT reject world-readable key file")
	}
}

func TestLoadOrGenerate_GeneratesAndPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "agent.key")
	a, created, err := LoadOrGenerate(path)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Errorf("expected created=true on first call")
	}
	b, created, err := LoadOrGenerate(path)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Errorf("expected created=false on second call")
	}
	if a.Address != b.Address {
		t.Errorf("addresses diverged across calls: %s vs %s", a.Address, b.Address)
	}
}

func TestLoadFromHexEnv(t *testing.T) {
	a, _ := Generate()
	priv := a.PrivateKey()
	hexKey := "0x" + privKeyHex(priv)
	t.Setenv("TEST_HL_KEY", hexKey)
	b, err := LoadFromHexEnv("TEST_HL_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if b.Address != a.Address {
		t.Errorf("recovered address %s, want %s", b.Address, a.Address)
	}

	t.Setenv("TEST_HL_KEY", "deadbeef") // too short
	if _, err := LoadFromHexEnv("TEST_HL_KEY"); err == nil {
		t.Error("expected error on short hex")
	}
}
