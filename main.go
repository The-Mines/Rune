package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/The-Mines/Rune/pkg/github"
	"github.com/The-Mines/Rune/pkg/gitlab"
	"github.com/The-Mines/Rune/pkg/gpg"
	"github.com/The-Mines/Rune/pkg/kubernetes"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "rune",
		Short: "Rune is a CLI tool for setting up VCS and Kubernetes integrations",
		Long:  `Rune simplifies the process of setting up VCS integrations and Kubernetes secrets.`,
	}

	var bootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap VCS and Kubernetes integrations",
		Run:   runBootstrap,
	}

	bootstrapCmd.Flags().String("vcs-type", "github", "VCS type (github or gitlab)")
	bootstrapCmd.Flags().String("vcs-token", "", "VCS Personal Access Token")
	bootstrapCmd.Flags().String("vcs-repo", "", "VCS repository (owner/repo for GitHub, project_id for GitLab)")
	bootstrapCmd.Flags().String("gpg-name", "Rune Bot", "Name for GPG key")
	bootstrapCmd.Flags().String("gpg-email", "rune-bot@example.com", "Email for GPG key")
	bootstrapCmd.Flags().Int("key-length", 4096, "Key length for both GPG and SSH keys")
	bootstrapCmd.Flags().Int("gpg-expiry-days", 0, "GPG key expiry in days (0 for no expiry)")
	bootstrapCmd.Flags().String("kube-config", "", "Path to Kubernetes config file")
	bootstrapCmd.Flags().String("kube-namespace", "default", "Kubernetes namespace for resources")
	bootstrapCmd.Flags().String("output-file", "", "Path to output file for Kubernetes secret YAML")

	bootstrapCmd.MarkFlagRequired("vcs-token")
	bootstrapCmd.MarkFlagRequired("vcs-repo")

	rootCmd.AddCommand(bootstrapCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runBootstrap(cmd *cobra.Command, args []string) {
	vcsType, _ := cmd.Flags().GetString("vcs-type")
	vcsToken, _ := cmd.Flags().GetString("vcs-token")
	vcsRepo, _ := cmd.Flags().GetString("vcs-repo")
	gpgName, _ := cmd.Flags().GetString("gpg-name")
	gpgEmail, _ := cmd.Flags().GetString("gpg-email")
	keyLength, _ := cmd.Flags().GetInt("key-length")
	gpgExpiryDays, _ := cmd.Flags().GetInt("gpg-expiry-days")
	kubeConfig, _ := cmd.Flags().GetString("kube-config")
	kubeNamespace, _ := cmd.Flags().GetString("kube-namespace")
	outputFile, _ := cmd.Flags().GetString("output-file")

	// VCS integration
	var vcsClient interface {
		Authenticate() error
		AddDeployKey(string, string, string, bool) error
	}

	var err error
	switch vcsType {
	case "github":
		vcsClient, err = github.NewClient(vcsToken)
	case "gitlab":
		vcsClient, err = gitlab.NewGitLabClient(vcsToken)
	default:
		log.Fatalf("Unsupported VCS type: %s", vcsType)
	}

	if err != nil {
		log.Fatalf("Failed to create VCS client: %v", err)
	}

	if err := vcsClient.Authenticate(); err != nil {
		log.Fatalf("VCS authentication failed: %v", err)
	}

	// Key generation
	keyConfig := &gpg.Config{
		Name:       gpgName,
		Email:      gpgEmail,
		KeyLength:  keyLength,
		ExpiryDays: gpgExpiryDays,
	}
	keys, err := gpg.GenerateKeys(keyConfig)
	if err != nil {
		log.Fatalf("Key generation failed: %v", err)
	}

	// Create Kubernetes secret
	k8sConfig := kubernetes.Config{
		ConfigPath: kubeConfig,
		Namespace:  kubeNamespace,
		OutputFile: outputFile,
	}

	// If kubeConfig is empty and outputFile is not set, create a default output file
	if kubeConfig == "" && outputFile == "" {
		defaultOutputFile := filepath.Join(".", "rune-secret.yaml")
		k8sConfig.OutputFile = defaultOutputFile
		fmt.Printf("No Kubernetes config or output file specified. Creating secret file: %s\n", defaultOutputFile)
	}

	if err := kubernetes.CreateSecret(&k8sConfig, "rune-keys", keys.GPGPrivateKey, keys.GPGPublicKey, keys.SSHPrivateKey, keys.SSHPublicKey); err != nil {
		log.Fatalf("Failed to create Kubernetes secret: %v", err)
	}

	// Add the SSH public key as a deploy key to the repository
	var deployKeyErr error
	switch vcsType {
	case "github":
		owner, repo, found := strings.Cut(vcsRepo, "/")
		if !found {
			log.Fatalf("Invalid GitHub repository format. Use 'owner/repo'")
		}
		deployKeyErr = vcsClient.AddDeployKey(owner, repo, "Rune Deploy Key", keys.SSHPublicKey, true)
	case "gitlab":
		projectID := vcsRepo // For GitLab, vcsRepo should be the project ID
		deployKeyErr = vcsClient.AddDeployKey(projectID, "Rune Deploy Key", keys.SSHPublicKey, true)
	}

	if deployKeyErr != nil {
		log.Fatalf("Failed to add deploy key: %v", deployKeyErr)
	}

	fmt.Println("VCS integration and Kubernetes secret setup completed successfully!")
}