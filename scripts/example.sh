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