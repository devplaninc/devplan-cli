package devplan

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/devplaninc/devplan-cli/internal/components/spinner"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/user"
	"time"
)

func VerifyAuth() (string, error) {
	key := prefs.GetAPIKey()
	if key != "" {
		return key, nil
	}
	return RequestAuth()
}

func RequestAuth() (string, error) {
	fmt.Println("Starting authentication with Devplan service...")

	// Request a login link
	requestID, err := requestLoginLink()
	if err != nil {
		return "", fmt.Errorf("failed to request login link: %w", err)
	}
	cl := NewClient(Config{})
	loginURL := fmt.Sprintf("%s/user/apikey/approve/%s", cl.BaseURL, requestID)

	// Print the login link to the user
	fmt.Println("Please open the following link in your browser to authenticate:")
	fmt.Print(lipgloss.NewStyle().Underline(true).Margin(1, 0, 1, 0).Render(loginURL))
	type keyResult struct {
		key string
		err error
	}
	ctx, cancel := context.WithCancel(context.Background())
	resChan := make(chan keyResult, 1)
	go func() {
		defer cancel()
		apiKey, err := waitForAuthentication(requestID)
		if err != nil {
			resChan <- keyResult{err: err}
			return
		}
		prefs.SetAPIKey(apiKey)
		resChan <- keyResult{key: apiKey}
	}()

	err = spinner.Run(ctx, "Waiting for authentication to complete", "Authenticated")
	if err != nil {
		fmt.Println(out.Fail(err))
		return "", err
	}
	res := <-resChan
	if err := res.err; err != nil {
		fmt.Println(out.Fail(err))
		return "", err
	}

	fmt.Println("Authentication successful! API key has been stored.")
	return res.key, nil
}

func keyName() string {
	userName := "user"
	curUser, err := user.Current()
	if err == nil && curUser != nil {
		userName = curUser.Username
	}
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "local"
	}
	curDate := time.Now().Format("20060102")
	return fmt.Sprintf("cli-%s-%s-%s-%s", userName, hostname, curDate, rndSuf(4))
}

func rndSuf(sufLen int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := make([]byte, 4)

	for i := 0; i < sufLen; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			ret[i] = letters[0] // fallback in case of error
			continue
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}

// requestLoginLink sends a request to the Devplan service to get a unique login link
func requestLoginLink() (string, error) {
	cl := NewClient(Config{})
	name := keyName()
	apiKeyRequestURL := fmt.Sprintf("%s/api/v1/apikey/request", cl.BaseURL)
	data := map[string]string{"name": name}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to prepare data to send login link request: %w", err)
	}
	resp, err := http.Post(apiKeyRequestURL, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to send login link request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login link request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	_ = resp.Body.Close()

	var createKeyResponse struct {
		RequestID string `json:"requestID"`
	}

	err = json.Unmarshal(body, &createKeyResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal create key response: %w", err)
	}

	return createKeyResponse.RequestID, nil
}

// waitForAuthentication polls the Devplan service to check if the login link has been clicked
// and returns the API key when authentication is complete
func waitForAuthentication(requestID string) (string, error) {
	cl := NewClient(Config{})
	url := fmt.Sprintf("%s/api/v1/apikey/request/%s", cl.BaseURL, requestID)

	maxRetries := 60
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return "", fmt.Errorf("failed to check auth status: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}

		var keyResponse struct {
			PendingMessage string `json:"pendingMessage"`
			APIKey         string `json:"apiKey"`
		}

		err = json.Unmarshal(body, &keyResponse)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal auth response [%v]: %w", string(body), err)
		}

		if keyResponse.APIKey != "" {
			return keyResponse.APIKey, nil
		}

		// Wait before retrying
		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("authentication timed out")
}
