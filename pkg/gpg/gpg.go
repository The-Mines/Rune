package gpg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Config holds the configuration for GPG key generation
type Config struct {
	Name       string
	Email      string
	KeyLength  int
	ExpiryDays int
}

// Key represents a GPG key pair
type Key struct {
	PrivateKey string
	PublicKey  string
}

// GenerateKey generates a new GPG key pair based on the provided configuration
func GenerateKey(config *Config) (*Key, error) {
	// Generate key pair
	cmd := exec.Command("gpg", "--batch", "--gen-key")
	stdin := bytes.NewBufferString(fmt.Sprintf(`
Key-Type: RSA
Key-Length: %d
Name-Real: %s
Name-Email: %s
Expire-Date: %d
%%no-protection
%%commit
`, config.KeyLength, config.Name, config.Email, config.ExpiryDays))
	cmd.Stdin = stdin

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to generate GPG key: %v", err)
	}

	// Get the key ID of the newly generated key
	keyIDCmd := exec.Command("gpg", "--list-secret-keys", "--with-colons")
	output, err := keyIDCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list GPG keys: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var keyID string
	for _, line := range lines {
		if strings.HasPrefix(line, "sec:") {
			fields := strings.Split(line, ":")
			if len(fields) > 4 {
				keyID = fields[4]
				break
			}
		}
	}

	if keyID == "" {
		return nil, fmt.Errorf("failed to find generated GPG key ID")
	}

	// Export private key
	privateKeyCmd := exec.Command("gpg", "--export-secret-keys", "--armor", keyID)
	privateKey, err := privateKeyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to export private GPG key: %v", err)
	}

	// Export public key
	publicKeyCmd := exec.Command("gpg", "--export", "--armor", keyID)
	publicKey, err := publicKeyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to export public GPG key: %v", err)
	}

	return &Key{
		PrivateKey: string(privateKey),
		PublicKey:  string(publicKey),
	}, nil
}