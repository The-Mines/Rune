package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/The-Mines/Rune/pkg/github"
	"github.com/The-Mines/Rune/pkg/gpg"
	"github.com/The-Mines/Rune/pkg/kubernetes"
// 	"github.com/The-Mines/Rune/pkg/tekton"
// )
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "rune",
		Short: "Rune is a CLI tool for bootstrapping Tekton event listeners",
		Long:  `Rune simplifies the process of setting up Tekton event listeners with GitHub integration.`,
	}

	var bootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap Tekton event listeners",
		Run:   runBootstrap,
	}

	bootstrapCmd.Flags().String("github-token", "", "GitHub Personal Access Token")
	bootstrapCmd.Flags().String("cluster-name", "", "Kubernetes cluster name")
	bootstrapCmd.Flags().String("gpg-name", "Rune Tekton Bot", "Name for GPG key")
	bootstrapCmd.Flags().String("gpg-email", "rune-tekton-bot@example.com", "Email for GPG key")
	bootstrapCmd.Flags().Int("gpg-key-length", 4096, "GPG key length")
	bootstrapCmd.Flags().Int("gpg-expiry-days", 0, "GPG key expiry in days (0 for no expiry)")
	bootstrapCmd.Flags().String("kube-config", "", "Path to Kubernetes config file")
	bootstrapCmd.Flags().String("kube-namespace", "default", "Kubernetes namespace for the secret")
	bootstrapCmd.Flags().String("output-file", "", "Path to output file for Kubernetes secret YAML")

	bootstrapCmd.MarkFlagRequired("github-token")
	bootstrapCmd.MarkFlagRequired("cluster-name")

	rootCmd.AddCommand(bootstrapCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runBootstrap(cmd *cobra.Command, args []string) {
	githubToken, _ := cmd.Flags().GetString("github-token")
	clusterName, _ := cmd.Flags().GetString("cluster-name")
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
	k8sConfig := &kubernetes.Config{
		ConfigPath: kubeConfig,
		Namespace:  kubeNamespace,
		OutputFile: outputFile,
	}
	if err := kubernetes.CreateSecret(k8sConfig, "privatekey", key.PrivateKey); err != nil {
		log.Fatalf("Failed to create Kubernetes secret: %v", err)
	}

	// Create a new repository for Tekton resources
	repo, err := githubClient.CreateRepository("tekton-resources", "Repository for Tekton resources", false)
	if err != nil {
		log.Fatalf("Failed to create GitHub repository: %v", err)
	}

	// Add the public key as a deploy key to the repository
	_, err = githubClient.AddDeployKey(*repo.Owner.Login, *repo.Name, "Tekton Deploy Key", key.PublicKey, true)
	if err != nil {
		log.Fatalf("Failed to add deploy key: %v", err)
	}

	// Set up Tekton event listener
	if err := tekton.SetupEventListener(clusterName, *repo.CloneURL); err != nil {
		log.Fatalf("Failed to set up Tekton event listener: %v", err)
	}

	fmt.Println("Tekton event listener bootstrapped successfully!")
}