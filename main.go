package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/The-Mines/Rune/pkg/github"
	"github.com/The-Mines/Rune/pkg/gpg"
	"github.com/The-Mines/Rune/pkg/kubernetes"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "rune",
		Short: "Rune is a CLI tool for setting up GitHub and Kubernetes integrations",
		Long:  `Rune simplifies the process of setting up GitHub integrations and Kubernetes secrets.`,
	}

	var bootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap GitHub and Kubernetes integrations",
		Run:   runBootstrap,
	}

	bootstrapCmd.Flags().String("github-token", "", "GitHub Personal Access Token")
	bootstrapCmd.Flags().String("github-repo", "", "GitHub repository (owner/repo)")
	bootstrapCmd.Flags().String("gpg-name", "Rune Bot", "Name for GPG key")
	bootstrapCmd.Flags().String("gpg-email", "rune-bot@example.com", "Email for GPG key")
	bootstrapCmd.Flags().Int("gpg-key-length", 4096, "GPG key length")
	bootstrapCmd.Flags().Int("gpg-expiry-days", 0, "GPG key expiry in days (0 for no expiry)")
	bootstrapCmd.Flags().String("kube-config", "", "Path to Kubernetes config file")
	bootstrapCmd.Flags().String("kube-namespace", "default", "Kubernetes namespace for resources")
	bootstrapCmd.Flags().String("output-file", "", "Path to output file for Kubernetes secret YAML")

	bootstrapCmd.MarkFlagRequired("github-token")
	bootstrapCmd.MarkFlagRequired("github-repo")

	rootCmd.AddCommand(bootstrapCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runBootstrap(cmd *cobra.Command, args []string) {
	githubToken, _ := cmd.Flags().GetString("github-token")
	githubRepo, _ := cmd.Flags().GetString("github-repo")
	gpgName, _ := cmd.Flags().GetString("gpg-name")
	gpgEmail, _ := cmd.Flags().GetString("gpg-email")
	gpgKeyLength, _ := cmd.Flags().GetInt("gpg-key-length")
	gpgExpiryDays, _ := cmd.Flags().GetInt("gpg-expiry-days")
	kubeConfig, _ := cmd.Flags().GetString("kube-config")
	kubeNamespace, _ := cmd.Flags().GetString("kube-namespace")
	outputFile, _ := cmd.Flags().GetString("output-file")

	// GitHub integration
	githubClient, err := github.NewClient(githubToken)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	if err := githubClient.Authenticate(); err != nil {
		log.Fatalf("GitHub authentication failed: %v", err)
	}

	// GPG key generation
	gpgConfig := &gpg.Config{
		Name:       gpgName,
		Email:      gpgEmail,
		KeyLength:  gpgKeyLength,
		ExpiryDays: gpgExpiryDays,
	}
	key, err := gpg.GenerateKey(gpgConfig)
	if err != nil {
		log.Fatalf("GPG key generation failed: %v", err)
	}

	// Create Kubernetes secret for private key
	k8sConfig := kubernetes.Config{
		ConfigPath: kubeConfig,
		Namespace:  kubeNamespace,
		OutputFile: outputFile,
	}
	if err := kubernetes.CreateSecret(&k8sConfig, "privatekey", key.PrivateKey); err != nil {
		log.Fatalf("Failed to create Kubernetes secret: %v", err)
	}

	// Add the public key as a deploy key to the repository
	owner, repo, found := strings.Cut(githubRepo, "/")
	if !found {
		log.Fatalf("Invalid GitHub repository format. Use 'owner/repo'")
	}
	_, err = githubClient.AddDeployKey(owner, repo, "Rune Deploy Key", key.PublicKey, true)
	if err != nil {
		log.Fatalf("Failed to add deploy key: %v", err)
	}

	fmt.Println("GitHub integration and Kubernetes secret setup completed successfully!")
}