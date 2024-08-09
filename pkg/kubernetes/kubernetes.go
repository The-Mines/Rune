package kubernetes

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

// Config holds the configuration for Kubernetes operations
type Config struct {
	ConfigPath string
	Namespace  string
	OutputFile string
}

// CreateSecret creates a new Kubernetes secret with the given name and data
func CreateSecret(config *Config, secretName, privateKey string) error {
	// Load Kubernetes configuration
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load Kubernetes config: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes clientset: %v", err)
	}

	// Encode the private key
	encodedPrivateKey := base64.StdEncoding.EncodeToString([]byte(privateKey))

	// Create the secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: config.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"private-key": []byte(encodedPrivateKey),
		},
	}

	// If OutputFile is specified, write the secret to a file
	if config.OutputFile != "" {
		return writeSecretToFile(secret, config.OutputFile)
	}

	// Create the secret in the cluster
	_, err = clientset.CoreV1().Secrets(config.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret: %v", err)
	}

	fmt.Printf("Secret '%s' created successfully in namespace '%s'\n", secretName, config.Namespace)
	return nil
}

// writeSecretToFile writes the secret to a YAML file
func writeSecretToFile(secret *corev1.Secret, filename string) error {
	// Convert the secret to YAML
	yamlData, err := yaml.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret to YAML: %v", err)
	}

	// Write the YAML to a file
	err = ioutil.WriteFile(filename, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write secret to file: %v", err)
	}

	fmt.Printf("Secret YAML written to file: %s\n", filename)
	return nil
}