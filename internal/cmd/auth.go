package cmd

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/devplaninc/devplan-cli/internal/components/spinner"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/globals"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/user"

	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Devplan service",
		Long:  `Authenticate with the Devplan service to retrieve and store API key for future communications.`,
		Run:   runAuth,
	}
)

func init() {
	rootCmd.AddCommand(authCmd)
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

func runAuth(_ *cobra.Command, _ []string) {
	fmt.Println("Starting authentication with Devplan service...")

	// Request a login link
	requestID, err := requestLoginLink()
	if err != nil {
		fmt.Printf("Failed to request login link: %v\n", err)
		return
	}
	loginURL := fmt.Sprintf("%s/user/apikey/approve/%s", getBaseURL(), requestID)

	// Print the login link to the user
	fmt.Println("Please open the following link in your browser to authenticate:")
	fmt.Print(lipgloss.NewStyle().Underline(true).Margin(1, 0, 1, 0).Render(loginURL))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		apiKey, err := waitForAuthentication(requestID)
		if err != nil {
			fmt.Printf(out.Failf("%v\n", err))
			os.Exit(1)
		}
		err = storeAPIKey(apiKey)
		if err != nil {
			fmt.Printf(out.Failf("%v\n", err))
			os.Exit(1)
		}
		cancel()
	}()
	err = spinner.Run(ctx, "Waiting for authentication to complete", "Authenticated")
	if err != nil {
		fmt.Println(out.Failf(err.Error()))
		os.Exit(1)
	}
	fmt.Println("Authentication successful! API key has been stored.")
}

// requestLoginLink sends a request to the Devplan service to get a unique login link
func requestLoginLink() (string, error) {
	name := keyName()
	apiKeyRequestURL := fmt.Sprintf("%s/api/v1/apikey/request", getBaseURL())
	resp, err := http.Post(apiKeyRequestURL, "application/json", strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, name)))
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
	url := fmt.Sprintf("%s/api/v1/apikey/request/%s", getBaseURL(), requestID)

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
		time.Sleep(5 * time.Second)
	}

	return "", fmt.Errorf("authentication timed out")
}

// storeAPIKey stores the API key in the configuration file using protobuf
func storeAPIKey(apiKey string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".devplan")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	viper.Set(globals.APIkeyConfig, apiKey)
	viper.SetConfigFile(filepath.Join(configDir, "config.json"))
	err = viper.WriteConfig()
	if err != nil && !os.IsNotExist(err) {
		// Only log this error, don't return it
		fmt.Printf("Warning: Failed to update legacy config file: %v\n", err)
	}

	return nil
}
