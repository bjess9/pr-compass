package auth

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strconv"
	"time"
)

const (
	clientID          = "Ov23lijhjxWcGktMYQoA"
	deviceCodeURL     = "https://github.com/login/device/code"
	accessTokenURL    = "https://github.com/login/oauth/access_token"
)

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

func Authenticate() (string, error) {
	token, err := loadToken()
	if err == nil {
		fmt.Println("Token loaded successfully.")
		return token, nil
	}

	fmt.Println("Token not found. Starting OAuth Device Flow.") // Debugging print
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL-encoded response: %w", err)
	}

	deviceCodeResp := &DeviceCodeResponse{
		DeviceCode:      values.Get("device_code"),
		UserCode:        values.Get("user_code"),
		VerificationURI: values.Get("verification_uri"),
		ExpiresIn:       parseInt(values.Get("expires_in")),
		Interval:        parseInt(values.Get("interval")),
	}

	if deviceCodeResp.DeviceCode == "" || deviceCodeResp.UserCode == "" || deviceCodeResp.VerificationURI == "" {
		return nil, errors.New("failed to retrieve device code information")
	}

	return deviceCodeResp, nil
}

func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
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
	fmt.Scanln() // Wait for the user to press Enter
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

			body, err := ioutil.ReadAll(resp.Body)
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
	usr, err := user.Current()
	if err != nil {
		return err
	}
	tokenFile := usr.HomeDir + "/.prpilot_token"

	err = os.WriteFile(tokenFile, []byte(token), 0600)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	return nil
}

func loadToken() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenFile := usr.HomeDir + "/.prpilot_token"

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("no access token found; please authenticate")
	}
	return string(data), nil
}
