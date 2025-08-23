package auth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bjess9/pr-compass/internal/errors"
)

// getGitHubCLIToken attempts to get the current GitHub CLI token
func getGitHubCLIToken() string {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return ""
	}

	return token
}

// Authenticate returns a GitHub token from available sources in order of preference:
// 1. GITHUB_TOKEN environment variable
// 2. GitHub CLI token (gh auth token)
func Authenticate() (string, error) {
	// 1. Check environment variable first (highest priority)
	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		if validateToken(envToken) {
			fmt.Println("[✓] Using GitHub token from GITHUB_TOKEN environment variable")
			return envToken, nil
		}
		fmt.Println("[!] GITHUB_TOKEN environment variable contains invalid token format")
	}

	// 2. Try GitHub CLI token (gh auth token)
	if ghToken := getGitHubCLIToken(); ghToken != "" && validateToken(ghToken) {
		fmt.Println("[✓] Using GitHub CLI token (gh auth token)")
		return ghToken, nil
	}

	// No valid token found - provide helpful instructions
	return "", errors.NewAuthTokenMissingError()
}

// validateToken performs basic validation on token format
func validateToken(token string) bool {
	token = strings.TrimSpace(token)
	if len(token) < 20 {
		return false
	}

	// GitHub tokens have specific prefixes
	validPrefixes := []string{"ghp_", "gho_", "ghu_", "ghs_", "github_pat_"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(token, prefix) {
			return true
		}
	}

	// Legacy tokens (40 characters, no prefix) - still supported
	if len(token) == 40 {
		return true
	}

	return false
}
