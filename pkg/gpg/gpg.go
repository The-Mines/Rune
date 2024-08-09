package gpg

import (
	"fmt"
	"os"
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
	// Create a temporary GPG home directory
	tempDir, err := os.MkdirTemp("", "gpg-home")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary GPG home: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate key pair
	cmd := exec.Command("gpg", "--batch", "--passphrase", "", "--quick-generate-key", config.Email, "rsa"+fmt.Sprint(config.KeyLength), "sign", fmt.Sprint(config.ExpiryDays)+"d")
	cmd.Env = append(os.Environ(), "GNUPGHOME="+tempDir)
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to generate GPG key: %v, output: %s", err, output)
	}

	// Get the key ID of the newly generated key
	keyIDCmd := exec.Command("gpg", "--list-secret-keys", "--with-colons")
	keyIDCmd.Env = append(os.Environ(), "GNUPGHOME="+tempDir)
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
	privateKeyCmd := exec.Command("gpg", "--batch", "--passphrase", "", "--pinentry-mode", "loopback", "--export-secret-keys", "--armor", keyID)
	privateKeyCmd.Env = append(os.Environ(), "GNUPGHOME="+tempDir)
	privateKey, err := privateKeyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to export private GPG key: %v", err)
	}

	// Export public key
	publicKeyCmd := exec.Command("gpg", "--export", "--armor", keyID)
	publicKeyCmd.Env = append(os.Environ(), "GNUPGHOME="+tempDir)
	publicKey, err := publicKeyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to export public GPG key: %v", err)
	}

	return &Key{
		PrivateKey: string(privateKey),
		PublicKey:  string(publicKey),
	}, nil
}