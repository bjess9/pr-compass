package auth

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bjess9/pr-pilot/internal/ui"
	"github.com/zalando/go-keyring"
)

const (
	service        = "prpilot"
	tokenKey       = "auth_token"
	clientID       = "Ov23lijhjxWcGktMYQoA"
	deviceCodeURL  = "https://github.com/login/device/code"
	accessTokenURL = "https://github.com/login/oauth/access_token"
)

func getTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".prpilot_token"), nil
}

func Authenticate() (string, error) {
	token, err := loadToken()
	if err == nil {
		fmt.Println("Token loaded successfully.")
		return token, nil
	}

	fmt.Println("Token not found. Starting OAuth Device Flow.")
	deviceCodeResp, err := getDeviceCode()
	if err != nil {
		return "", err
	}

	promptUserForAuthorization(deviceCodeResp)
	tokenResp, err := pollForAccessToken(deviceCodeResp.DeviceCode, deviceCodeResp.Interval)
	if err != nil {
		return "", err
	}

	if err := saveToken(tokenResp.AccessToken); err != nil {
		return "", err
	}

	fmt.Println("Token saved successfully.")
	return tokenResp.AccessToken, nil
}

func getDeviceCode() (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", "repo")

	resp, err := http.PostForm(deviceCodeURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL-encoded response: %w", err)
	}

	return &DeviceCodeResponse{
		DeviceCode:      values.Get("device_code"),
		UserCode:        values.Get("user_code"),
		VerificationURI: values.Get("verification_uri"),
		ExpiresIn:       parseInt(values.Get("expires_in")),
		Interval:        parseInt(values.Get("interval")),
	}, nil
}

func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func promptUserForAuthorization(dcr *DeviceCodeResponse) {
	fmt.Printf("To authenticate, please visit the following URL:\n\n%s\n\n", dcr.VerificationURI)
	fmt.Printf("Then enter the code: %s\n", dcr.UserCode)
	fmt.Println("\nPress Enter when you have completed this step.")
	fmt.Scanln()
}

func pollForAccessToken(deviceCode string, interval int) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	for {
		time.Sleep(time.Duration(interval) * time.Second)

		resp, err := http.PostForm(accessTokenURL, data)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL-encoded response: %w", err)
		}

		token := values.Get("access_token")
		if token != "" {
			return &TokenResponse{
				AccessToken: token,
				TokenType:   values.Get("token_type"),
				Scope:       values.Get("scope"),
			}, nil
		}

		switch values.Get("error") {
		case "authorization_pending":
			continue
		case "slow_down":
			interval++
		case "expired_token":
			return nil, errors.New("device code expired")
		default:
			return nil, errors.New("failed to authenticate")
		}
	}
}

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

func loadToken() (string, error) {
	if ui.IsWSL() {
		tokenFilePath, err := getTokenFilePath()
		if err != nil {
			return "", err
		}
		token, err := os.ReadFile(tokenFilePath)
		if err != nil {
			return "", fmt.Errorf("token file not found; please authenticate")
		}
		return strings.TrimSpace(string(token)), nil
	}
	return keyring.Get(service, tokenKey)
}

// DeleteToken removes the stored authentication token
func DeleteToken() error {
	if ui.IsWSL() {
		tokenFilePath, err := getTokenFilePath()
		if err != nil {
			return err
		}
		return os.Remove(tokenFilePath)
	}
	return keyring.Delete(service, tokenKey)
}
