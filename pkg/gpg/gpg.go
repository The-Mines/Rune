package gpg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"os/exec"
	"strings"
)

// Config holds the configuration for key generation
type Config struct {
	Name       string
	Email      string
	KeyLength  int
	ExpiryDays int
}

// Key represents a key pair
type Key struct {
	GPGPrivateKey string
	GPGPublicKey  string
	SSHPrivateKey string
	SSHPublicKey  string
}

// GenerateKeys generates both GPG and SSH key pairs
func GenerateKeys(config *Config) (*Key, error) {
	gpgKey, err := generateGPGKey(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GPG key: %v", err)
	}

	sshKey, err := generateSSHKey(config.KeyLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SSH key: %v", err)
	}

	return &Key{
		GPGPrivateKey: gpgKey.GPGPrivateKey,
		GPGPublicKey:  gpgKey.GPGPublicKey,
		SSHPrivateKey: sshKey.SSHPrivateKey,
		SSHPublicKey:  sshKey.SSHPublicKey,
	}, nil
}

func generateGPGKey(config *Config) (*Key, error) {
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

	// Export public key in ASCII armor format
	publicKeyCmd := exec.Command("gpg", "--armor", "--export", keyID)
	publicKeyCmd.Env = append(os.Environ(), "GNUPGHOME="+tempDir)
	publicKey, err := publicKeyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to export public GPG key: %v", err)
	}

	return &Key{
		GPGPrivateKey: string(privateKey),
		GPGPublicKey:  string(publicKey),
	}, nil
}

func generateSSHKey(bits int) (*Key, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %v", err)
	}

	// Convert the private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyString := string(pem.EncodeToMemory(privateKeyPEM))

	// Generate the public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create public key: %v", err)
	}

	publicKeyString := string(ssh.MarshalAuthorizedKey(publicKey))

	return &Key{
		SSHPrivateKey: privateKeyString,
		SSHPublicKey:  publicKeyString,
	}, nil
}