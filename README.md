# Keyman

> A minimal, encrypted credential manager for the terminal — built with Go.

```
⬡  KEYMAN
Secure credential manager

  Vaults

  ⬡ work
    personal
    freelance

  enter open   n new vault   D delete vault   ? help   q quit
```

Keyman stores your API keys, tokens, passwords, and secrets in AES-256-GCM encrypted vault files on your local machine. No cloud, no account, no tracking — just a fast terminal UI and strong encryption.

---

## Features

- **AES-256-GCM encryption** — every vault is encrypted at rest with a key derived from your master password via PBKDF2-SHA256 (100,000 iterations)
- **Multiple vaults** — separate namespaces for work, personal, freelance, and anything else
- **Password generator** — cryptographically secure passwords with configurable length and character classes, with a live strength indicator
- **Clipboard copy** — copy any secret to your clipboard without it ever appearing on screen
- **Masked values** — all values hidden by default; reveal individually on demand
- **Live search** — filter entries as you type with `/`
- **Edit in place** — update existing entries without deleting and re-adding
- **Atomic writes** — vault files are written via temp file + rename so a crash mid-save never corrupts your data

---

## Installation

### From source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/HiwarkhedePrasad/keyman.git
cd keyman
make install
```

This builds the binary and installs it to `/usr/local/bin/keyman`.

### Binary release

Download the prebuilt binary for your platform from the [Releases](https://github.com/HiwarkhedePrasad/keyman/releases) page, make it executable, and place it on your `PATH`.

```bash
# Example for Linux amd64
chmod +x keyman-linux-amd64
sudo mv keyman-linux-amd64 /usr/local/bin/keyman
```

Verify the download against `checksums.txt` before running:

```bash
sha256sum -c checksums.txt --ignore-missing
```

---

## Building

```bash
# Build for the current platform
make build

# Run without installing
make run

# Build release binaries for all platforms (outputs to dist/)
make release

# Run all tests
make test

# Run the linter (requires golangci-lint)
make lint
```

### Supported platforms

| OS      | Architecture |
|---------|-------------|
| Linux   | amd64, arm64 |
| macOS   | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64 |

---

## Usage

```bash
keyman
```

Keyman opens to the vault list. Select a vault with the arrow keys and press `enter` to unlock it with your master password.

### Keyboard shortcuts

**Vault list**

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate vaults |
| `enter` | Open selected vault |
| `n` | Create new vault |
| `D` | Delete selected vault |
| `?` | Help |
| `q` | Quit |

**Key list** (inside a vault)

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate entries |
| `enter` / `c` | Copy value to clipboard |
| `v` | Reveal / hide value |
| `a` | Add new entry |
| `e` | Edit selected entry |
| `d` | Delete selected entry |
| `/` | Search and filter entries |
| `g` | Open password generator |
| `b` / `esc` | Lock vault and go back |

**Password generator**

| Key | Action |
|-----|--------|
| `r` | Regenerate password |
| `+` / `-` | Increase / decrease length |
| `u` | Toggle uppercase |
| `d` | Toggle digits |
| `s` | Toggle symbols |
| `enter` | Copy to clipboard and close |
| `esc` | Close without copying |

**Input fields**

| Key | Action |
|-----|--------|
| `tab` | Toggle password visibility |
| `ctrl+g` | Generate a strong password directly into the field |

---

## Security

Keyman is designed around the principle that your secrets never leave your machine unencrypted.

- **Local only** — vault files live in `keyman_vaults/` at the project root (detected via `go.mod`); nothing is sent anywhere
- **File permissions** — vault files are created with `0600` (readable only by you)
- **Fresh encryption on every save** — each write generates a new random salt and nonce, so saving the same data twice produces a completely different ciphertext
- **Password never stored** — your master password exists only in process memory for the duration of the session and is never written to disk
- **No recovery** — if you forget your master password, the vault cannot be decrypted. Keep it somewhere safe.
- **Atomic writes** — saves go through a `.tmp` file that is renamed into place, so a crash mid-write cannot corrupt an existing vault

### Storage format

Each `.vault` file has the following binary layout:

```
┌──────────────────────────────────────┐
│  32 bytes  │  random PBKDF2 salt     │
├──────────────────────────────────────┤
│  12 bytes  │  AES-GCM nonce          │
├──────────────────────────────────────┤
│  N bytes   │  AES-256-GCM ciphertext │
├──────────────────────────────────────┤
│  16 bytes  │  GCM authentication tag │
└──────────────────────────────────────┘
```

The decrypted plaintext is a JSON object:

```json
{
  "entries": {
    "GitHub API Key": "ghp_xxxxxxxxxxxx",
    "AWS Secret": "xxxxxxxxxxxxxxxxxxxx"
  }
}
```

---

## Contributing

Contributions are welcome. Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Make your changes and add tests where applicable
4. Run `make lint` and `make test` — both must pass
5. Open a pull request with a clear description of what changed and why

Please keep pull requests focused. One feature or fix per PR.

---

## Author

[Prasad Hiwarkhedeprasad](https://github.com/HiwarkhedePrasad) 

---

## License

MIT — see [LICENSE](LICENSE).
