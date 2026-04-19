// Package storage handles reading and writing encrypted Keyman vault files.
// Each vault is a JSON document encrypted with AES-256-GCM. The file layout is:
//
//	[32-byte salt][encrypted JSON payload]
//
// The JSON payload has the structure:
//
//	{ "entries": { "name": "value", ... } }
package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/keyman/keyman/internal/crypto"
)

var (
	DataDir = "keyman_vaults"
)

const ext = ".vault"

// Vault holds the decrypted entries for a folder.
type Vault struct {
	Entries map[string]string `json:"entries"`
}

// EnsureDataDir creates the data directory if it does not exist.
func EnsureDataDir() error {
	return os.MkdirAll(resolvedDataDir(), 0o700)
}

// ListVaults returns the names of all vaults on disk (without extension).
func ListVaults() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(resolvedDataDir(), "*"+ext))
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(files))
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, ext)
		names = append(names, name)
	}
	return names, nil
}

// VaultExists reports whether a vault with the given name exists.
func VaultExists(name string) bool {
	_, err := os.Stat(vaultPath(name))
	return err == nil
}

// LoadVault decrypts and deserialises a vault using the master password.
func LoadVault(name, password string) (*Vault, error) {
	data, err := os.ReadFile(vaultPath(name))
	if err != nil {
		return nil, err
	}

	if len(data) < 32 {
		return nil, errors.New("vault file is corrupted")
	}

	salt := data[:32]
	ciphertext := data[32:]

	key := crypto.DeriveKey(password, salt)
	plaintext, err := crypto.Decrypt(key, ciphertext)
	if err != nil {
		return nil, err
	}

	var vault Vault
	if err := json.Unmarshal(plaintext, &vault); err != nil {
		return nil, err
	}

	if vault.Entries == nil {
		vault.Entries = make(map[string]string)
	}

	return &vault, nil
}

// SaveVault serialises and encrypts a vault, writing it to disk.
func SaveVault(name, password string, vault *Vault) error {
	if vault.Entries == nil {
		vault.Entries = make(map[string]string)
	}

	plaintext, err := json.Marshal(vault)
	if err != nil {
		return err
	}

	salt, err := crypto.NewSalt()
	if err != nil {
		return err
	}

	key := crypto.DeriveKey(password, salt)
	ciphertext, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		return err
	}

	// Write salt then ciphertext atomically via temp file.
	tmp := vaultPath(name) + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	if _, err := f.Write(salt); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if _, err := f.Write(ciphertext); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, vaultPath(name))
}

// CreateVault creates a new empty vault with the given name and password.
func CreateVault(name, password string) error {
	vault := &Vault{Entries: make(map[string]string)}
	return SaveVault(name, password, vault)
}

// DeleteVault removes a vault file from disk.
func DeleteVault(name string) error {
	return os.Remove(vaultPath(name))
}

func vaultPath(name string) string {
	return filepath.Join(resolvedDataDir(), name+ext)
}

func resolvedDataDir() string {
	if filepath.IsAbs(DataDir) {
		return DataDir
	}

	wd, err := os.Getwd()
	if err != nil {
		return DataDir
	}

	if root := projectRootFrom(wd); root != "" {
		return filepath.Join(root, DataDir)
	}

	return filepath.Join(wd, DataDir)
}

func projectRootFrom(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
