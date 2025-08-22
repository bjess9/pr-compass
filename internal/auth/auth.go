package auth

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bjess9/pr-pilot/internal/ui"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

const (
	service  = "prpilot"
	tokenKey = "auth_token"
)

// getTokenFilePath returns the path to the token file
func getTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".prpilot_token"), nil
}

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

// Authenticate returns a GitHub token from various sources in order of preference:
// 1. GITHUB_TOKEN environment variable
// 2. GitHub CLI token (gh auth token) 
// 3. Stored token (keyring/file)
// 4. Manual token input from user
func Authenticate() (string, error) {
	// 1. Check environment variable first (highest priority)
	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		if validateToken(envToken) {
			fmt.Println("[✓] Using GitHub token from GITHUB_TOKEN environment variable")
			return envToken, nil
		}
		fmt.Println("[!] GITHUB_TOKEN environment variable contains invalid token")
	}

	// 2. Try GitHub CLI token (gh auth token)
	if ghToken := getGitHubCLIToken(); ghToken != "" && validateToken(ghToken) {
		fmt.Println("[✓] Using GitHub CLI token (gh auth token)")
		return ghToken, nil
	}

	// 3. Try to load stored token
	token, err := loadToken()
	if err == nil && validateToken(token) {
		fmt.Println("[✓] Using stored GitHub token")
		return token, nil
	}

	// 4. Prompt user for manual token input
	fmt.Println("[AUTH] No valid GitHub token found. Please provide one manually.")
	fmt.Println()
	fmt.Println("To create a GitHub Personal Access Token:")
	fmt.Println("1. Go to: https://github.com/settings/tokens")
	fmt.Println("2. Click 'Generate new token (classic)'")
	fmt.Println("3. Give it a name like 'pr-pilot'")
	fmt.Println("4. Select scopes: 'repo' and 'read:org' (minimum required)")
	fmt.Println("5. Click 'Generate token'")
	fmt.Println()
	
	for {
		fmt.Print("Enter your GitHub Personal Access Token: ")
		tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", fmt.Errorf("failed to read token: %w", err)
		}
		fmt.Println() // New line after hidden input
		
		token := strings.TrimSpace(string(tokenBytes))
		if token == "" {
			fmt.Println("[X] Token cannot be empty. Please try again.")
			continue
		}
		
		if !validateToken(token) {
			fmt.Println("[X] Invalid token format. GitHub tokens start with 'ghp_', 'gho_', 'ghu_', or 'ghs_'")
			fmt.Println("   Please try again or press Ctrl+C to exit.")
			continue
		}
		
		// Save the token for future use
		if err := saveToken(token); err != nil {
			fmt.Printf("[!] Token accepted but failed to save: %v\n", err)
			fmt.Println("   You'll need to enter it again next time.")
		} else {
			fmt.Println("[✓] Token saved for future use")
		}
		
		return token, nil
	}
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

// saveToken stores the token securely
func saveToken(token string) error {
	if ui.IsWSL() {
		tokenFilePath, err := getTokenFilePath()
		if err != nil {
			return err
		}
		return os.WriteFile(tokenFilePath, []byte(token), 0600)
	}
	return keyring.Set(service, tokenKey, token)
}

// loadToken retrieves the stored token
func loadToken() (string, error) {
	if ui.IsWSL() {
		tokenFilePath, err := getTokenFilePath()
		if err != nil {
			return "", err
		}
		token, err := os.ReadFile(tokenFilePath)
		if err != nil {
			return "", fmt.Errorf("token file not found: %w", err)
		}
		return strings.TrimSpace(string(token)), nil
	}
	return keyring.Get(service, tokenKey)
}

// DeleteToken removes the stored authentication token
func DeleteToken() error {
	fmt.Println("[DEL] Clearing stored GitHub token...")
	
	if ui.IsWSL() {
		tokenFilePath, err := getTokenFilePath()
		if err != nil {
			return err
		}
		if err := os.Remove(tokenFilePath); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("   No token file found to delete")
				return nil
			}
			return err
		}
		fmt.Println("[✓] Token file deleted successfully")
		return nil
	}
	
	if err := keyring.Delete(service, tokenKey); err != nil {
		fmt.Printf("[!] Failed to delete token from keyring: %v\n", err)
		return err
	}
	fmt.Println("[✓] Token deleted from keyring successfully")
	return nil
}
