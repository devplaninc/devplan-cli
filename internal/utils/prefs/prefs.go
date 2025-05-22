package prefs

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// Domain is used to specify which domain to use (app, beta, local)
var Domain string

const (
	LastCompanyIDKey    = "last_company_id"
	LastProjectIDKey    = "last_project_id"
	LastGitProtocolKey  = "last_git_protocol"
	LastAssistantConfig = "last_assistant"

	APIkeyConfig = "apikey"
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get home directory: %w", err))
	}
	configDir := filepath.Join(home, ".devplan")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			panic(fmt.Errorf("failed to create config directory: %w", err))
		}
	}
	viper.SetConfigFile(filepath.Join(configDir, "config.json"))
}

// GitProtocol represents the git protocol to use (https or ssh)
type GitProtocol string

const (
	HTTPS GitProtocol = "https"
	SSH   GitProtocol = "ssh"
)

// GetLastCompanyID returns the last selected company ID from config
func GetLastCompanyID() int32 {
	return int32(viper.GetInt(LastCompanyIDKey))
}

// SetLastCompanyID saves the last selected company ID to config
func SetLastCompanyID(id int32) {
	viper.Set(LastCompanyIDKey, id)
	_ = viper.WriteConfig()
}

// GetLastProjectID returns the last selected project ID from config
func GetLastProjectID() string {
	return viper.GetString(LastProjectIDKey)
}

// SetLastProjectID saves the last selected project ID to config
func SetLastProjectID(id string) {
	viper.Set(LastProjectIDKey, id)
	_ = viper.WriteConfig()
}

// GetLastGitProtocol returns the last used git protocol from config
func GetLastGitProtocol() GitProtocol {
	protocol := viper.GetString(LastGitProtocolKey)
	if protocol == string(SSH) {
		return SSH
	}
	return HTTPS
}

// SetLastGitProtocol saves the last used git protocol to config
func SetLastGitProtocol(protocol GitProtocol) {
	viper.Set(LastGitProtocolKey, string(protocol))
	_ = viper.WriteConfig()
}

// GetLastAssistant returns the last selected assistant from config
func GetLastAssistant() (string, bool) {
	v := viper.GetString(LastAssistantConfig)
	return v, v != ""
}

// SetLastAssistant saves the last selected assistant to config
func SetLastAssistant(asst string) {
	viper.Set(LastAssistantConfig, asst)
	_ = viper.WriteConfig()
}

func SetAPIKey(apiKey string) {
	viper.Set(APIkeyConfig, apiKey)
	err := viper.WriteConfig()
	if err != nil {
		out.Pfailf("Failed to write config: %v\n", err)
	}
}
