# Rune CLI Tool

Rune is a command-line interface (CLI) tool designed to simplify the process of setting up Version Control System (VCS) integrations and Kubernetes secrets. It currently supports GitHub and GitLab as VCS providers.

## Features

- Authenticate with GitHub or GitLab
- Generate GPG and SSH key pairs
- Create Kubernetes secrets (in-cluster or as YAML files)
- Add deploy keys to GitHub repositories or GitLab projects

## Prerequisites

- Go 1.16 or later
- Access to GitHub or GitLab account
- (Optional) Access to a Kubernetes cluster

## Installation

1. Clone the repository:
   ```
   git clone git@github.com:The-Mines/Rune.git
   ```

2. Build the project:
   ```
   go build -o rune .
   ```

3. (Optional) Move the binary to a directory in your PATH for easy access:
   ```
   sudo mv rune /usr/local/bin/
   ```

## Usage

### Basic Command Structure

```
./rune bootstrap [flags]
```

### Flags

- `--vcs-type`: VCS type (github or gitlab)
- `--vcs-token`: VCS Personal Access Token
- `--vcs-repo`: VCS repository (owner/repo for GitHub, namespace/project-name for GitLab)
- `--gpg-name`: Name for GPG key
- `--gpg-email`: Email for GPG key
- `--key-length`: Key length for both GPG and SSH keys
- `--gpg-expiry-days`: GPG key expiry in days (0 for no expiry)
- `--kube-config`: Path to Kubernetes config file
- `--kube-namespace`: Kubernetes namespace for resources
- `--output-file`: Path to output file for Kubernetes secret YAML

### Example: GitHub Integration

```bash
./rune bootstrap \
  --vcs-type github \
  --vcs-token ghp_YourGitHubPersonalAccessTokenHere \
  --vcs-repo octocat/Hello-World \
  --gpg-name "Rune Bot" \
  --gpg-email "rune-bot@example.com" \
  --key-length 4096 \
  --gpg-expiry-days 365 \
  --kube-namespace default \
  --output-file ./rune-secret.yaml
```

### Example: GitLab Integration

```bash
./rune bootstrap \
  --vcs-type gitlab \
  --vcs-token glpat_YourGitLabPersonalAccessTokenHere \
  --vcs-repo namespace/project-name \
  --gpg-name "Rune Bot" \
  --gpg-email "rune-bot@example.com" \
  --key-length 4096 \
  --gpg-expiry-days 365 \
  --kube-namespace default \
  --output-file ./rune-secret.yaml
```

## Output

The tool will:
1. Authenticate with the specified VCS (GitHub or GitLab)
2. Generate GPG and SSH key pairs
3. Create a Kubernetes secret (either in your cluster or as a YAML file)
4. Add the SSH public key as a deploy key to the specified VCS repository/project

## Security Notes

- Keep your VCS tokens secret and never share them publicly.
- If using this tool in a CI/CD pipeline, use environment variables or secrets management tools to handle sensitive information like tokens.
- The generated Kubernetes secret contains sensitive information. Ensure it's properly secured.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Apache License 2.0](LICENSE)