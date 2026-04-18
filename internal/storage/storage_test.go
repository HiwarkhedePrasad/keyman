package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keyman/keyman/internal/storage"
)

func init() {
	// Use a temp dir for tests so they don't pollute the real data dir.
	dir, _ := os.MkdirTemp("", "keyman-test-*")
	storage.DataDir = dir
}

func TestCreateAndLoad(t *testing.T) {
	const name = "test-vault"
	const pass = "testpassword123"

	if err := storage.CreateVault(name, pass); err != nil {
		t.Fatalf("CreateVault() error: %v", err)
	}
	defer storage.DeleteVault(name)

	vault, err := storage.LoadVault(name, pass)
	if err != nil {
		t.Fatalf("LoadVault() error: %v", err)
	}

	if vault.Entries == nil {
		t.Fatal("expected non-nil entries map")
	}
}

func TestSaveAndLoad(t *testing.T) {
	const name = "test-save"
	const pass = "s3cur3P@ss"

	storage.CreateVault(name, pass)
	defer storage.DeleteVault(name)

	vault, _ := storage.LoadVault(name, pass)
	vault.Entries["github"] = "ghp_secret123"
	vault.Entries["aws"] = "AKIAIOSFODNN7EXAMPLE"

	if err := storage.SaveVault(name, pass, vault); err != nil {
		t.Fatalf("SaveVault() error: %v", err)
	}

	loaded, err := storage.LoadVault(name, pass)
	if err != nil {
		t.Fatalf("LoadVault() after save error: %v", err)
	}

	if loaded.Entries["github"] != "ghp_secret123" {
		t.Errorf("github entry mismatch: got %q", loaded.Entries["github"])
	}
	if loaded.Entries["aws"] != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("aws entry mismatch: got %q", loaded.Entries["aws"])
	}
}

func TestWrongPassword(t *testing.T) {
	const name = "test-wrong-pass"
	storage.CreateVault(name, "correct")
	defer storage.DeleteVault(name)

	_, err := storage.LoadVault(name, "incorrect")
	if err == nil {
		t.Fatal("expected error when loading with wrong password")
	}
}

func TestListVaults(t *testing.T) {
	for _, n := range []string{"list-a", "list-b", "list-c"} {
		storage.CreateVault(n, "pass")
		defer storage.DeleteVault(n)
	}

	vaults, err := storage.ListVaults()
	if err != nil {
		t.Fatalf("ListVaults() error: %v", err)
	}

	found := make(map[string]bool)
	for _, v := range vaults {
		found[v] = true
	}

	for _, n := range []string{"list-a", "list-b", "list-c"} {
		if !found[n] {
			t.Errorf("expected vault %q in list", n)
		}
	}
}

func TestDeleteVault(t *testing.T) {
	const name = "test-delete"
	storage.CreateVault(name, "pass")

	if err := storage.DeleteVault(name); err != nil {
		t.Fatalf("DeleteVault() error: %v", err)
	}

	path := filepath.Join(storage.DataDir, name+".vault")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("vault file should not exist after deletion")
	}
}

func TestVaultExists(t *testing.T) {
	const name = "test-exists"
	storage.CreateVault(name, "pass")
	defer storage.DeleteVault(name)

	if !storage.VaultExists(name) {
		t.Fatal("VaultExists should return true for existing vault")
	}
	if storage.VaultExists("no-such-vault-xyz") {
		t.Fatal("VaultExists should return false for non-existent vault")
	}
}
