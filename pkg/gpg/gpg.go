package gpg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/ssh"
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

	// Export public key in OpenSSH format
	publicKeyCmd := exec.Command("gpg", "--export", keyID)
	publicKeyCmd.Env = append(os.Environ(), "GNUPGHOME="+tempDir)
	publicKeyData, err := publicKeyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to export public GPG key: %v", err)
	}

	// Convert PGP public key to OpenSSH format
	publicKey, err := convertToOpenSSH(publicKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert public key to OpenSSH format: %v", err)
	}

	return &Key{
		PrivateKey: string(privateKey),
		PublicKey:  publicKey,
	}, nil
}

func convertToOpenSSH(pgpPublicKey []byte) (string, error) {
	// Read the PGP public key
	entityList, err := openpgp.ReadKeyRing(bytes.NewReader(pgpPublicKey))
	if err != nil {
		return "", fmt.Errorf("failed to read PGP key: %v", err)
	}

	if len(entityList) == 0 {
		return "", fmt.Errorf("no PGP keys found")
	}

	// Get the first key
	entity := entityList[0]
	publicKey := entity.PrimaryKey

	// Convert to SSH public key
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH public key: %v", err)
	}

	// Marshal the SSH public key to authorized_keys format
	sshPublicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)

	return string(sshPublicKeyBytes), nil
}