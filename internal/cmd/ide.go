package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/globals"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// IDE type constants
const (
	IDETypeCursor   = "cursor"
	IDETypeIntelliJ = "intellij"
	IDETypeJunie    = "junie"
)

var (
	ideCmd = &cobra.Command{
		Use:   "ide",
		Short: "Manage IDE configurations",
		Long:  `Download rules, configure, and start IDEs with Devplan configurations.`,
		Run:   runIDE,
	}

	cursorCmd = &cobra.Command{
		Use:   "cursor",
		Short: "Configure and start Cursor IDE",
		Long:  `Download Cursor IDE rules from Devplan service, initialize and start Cursor with those rules. Automatically configures MCP servers for AI-powered features.`,
		Run:   runCursor,
	}

	intellijCmd = &cobra.Command{
		Use:   "intellij",
		Short: "Configure and start IntelliJ IDE",
		Long:  `Download IntelliJ IDE rules from Devplan service, initialize and start IntelliJ with those rules.`,
		Run:   runIntelliJ,
	}

	junieCmd = &cobra.Command{
		Use:   "junie",
		Short: "Configure and start a JetBrains IDE with Junie plugin",
		Long:  `Download Junie plugin rules from Devplan service, configure the plugin, and start a compatible JetBrains IDE with those rules. Junie is a plugin for JetBrains IDEs, not a standalone application. Automatically configures MCP servers for AI-powered features.`,
		Run:   runJunie,
	}

	resetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset the default IDE preference",
		Long:  `Reset the default IDE preference, prompting for selection on next run.`,
		Run:   runReset,
	}

	projectPath string
	ideType     string
)

//func init() {
//	rootCmd.AddCommand(ideCmd)
//	ideCmd.AddCommand(cursorCmd)
//	ideCmd.AddCommand(intellijCmd)
//	ideCmd.AddCommand(junieCmd)
//	ideCmd.AddCommand(resetCmd)
//
//	// Common flags for all IDE commands
//	ideCmd.PersistentFlags().StringVarP(&projectPath, "project", "p", "", "Path to the project directory")
//	ideCmd.PersistentFlags().StringVarP(&ideType, "ide", "i", "", "IDE type to use (cursor, intellij, junie)")
//}

// getDefaultIDE gets the default IDE from the configuration
func getDefaultIDE() string {
	return viper.GetString("default_ide")
}

// setDefaultIDE sets the default IDE in the configuration
func setDefaultIDE(ide string) error {
	viper.Set("default_ide", ide)
	return viper.WriteConfig()
}

// runIDE runs the appropriate IDE based on the default IDE or user selection
func runIDE(cmd *cobra.Command, args []string) {
	// Check if IDE type is specified via flag
	if ideType != "" {
		switch ideType {
		case IDETypeCursor:
			runCursor(cmd, args)
		case IDETypeIntelliJ:
			runIntelliJ(cmd, args)
		case IDETypeJunie:
			runJunie(cmd, args)
		default:
			fmt.Printf("Unknown IDE type: %s\n", ideType)
			fmt.Println("Available IDE types: cursor, intellij, junie")
		}
		return
	}

	// Check if default IDE is set
	defaultIDE := getDefaultIDE()
	if defaultIDE != "" {
		fmt.Printf("Using default IDE: %s\n", defaultIDE)
		switch defaultIDE {
		case IDETypeCursor:
			runCursor(cmd, args)
		case IDETypeIntelliJ:
			runIntelliJ(cmd, args)
		case IDETypeJunie:
			runJunie(cmd, args)
		default:
			fmt.Printf("Unknown default IDE type: %s\n", defaultIDE)
			fmt.Println("Prompting for IDE selection...")
			promptForIDESelection(cmd, args)
		}
		return
	}

	// No default IDE set, prompt for selection
	promptForIDESelection(cmd, args)
}

// runReset resets the default IDE preference
func runReset(cmd *cobra.Command, args []string) {
	err := setDefaultIDE("")
	if err != nil {
		fmt.Printf("Failed to reset default IDE preference: %v\n", err)
		return
	}
	fmt.Println("Default IDE preference has been reset.")
	fmt.Println("You will be prompted to select an IDE on the next run.")
}

// IDESelectionModel represents the TUI model for IDE selection
type IDESelectionModel struct {
	choices  []string
	cursor   int
	selected string
	quitting bool
}

// InitIDESelectionModel initializes a new IDE selection model
func InitIDESelectionModel() IDESelectionModel {
	return IDESelectionModel{
		choices:  []string{IDETypeCursor, IDETypeIntelliJ, IDETypeJunie},
		cursor:   0,
		selected: "",
		quitting: false,
	}
}

// Init initializes the IDE selection model
func (m IDESelectionModel) Init() tea.Cmd {
	return nil
}

// Update handles updates to the IDE selection model
func (m IDESelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.choices[m.cursor]
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the IDE selection model
func (m IDESelectionModel) View() string {
	s := "Select your preferred IDE:\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		// Format the choice based on the IDE type
		var description string
		switch choice {
		case IDETypeCursor:
			description = "Cursor IDE with Devplan rules and MCP server configuration"
		case IDETypeIntelliJ:
			description = "IntelliJ IDE with Devplan rules"
		case IDETypeJunie:
			description = "JetBrains IDE with Junie plugin, Devplan rules, and MCP server configuration"
		}

		s += fmt.Sprintf("%s %s - %s\n", cursor, choice, description)
	}

	s += "\nPress up/down to move, enter to select, q to quit\n"
	return s
}

// promptForIDESelection prompts the user to select an IDE and runs it
func promptForIDESelection(cmd *cobra.Command, args []string) {
	fmt.Println("Please select your preferred IDE:")

	model := InitIDESelectionModel()
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running IDE selection: %v\n", err)
		return
	}

	m, ok := finalModel.(IDESelectionModel)
	if !ok {
		fmt.Println("Error: could not cast final model to IDESelectionModel")
		return
	}

	if m.quitting {
		fmt.Println("IDE selection cancelled.")
		return
	}

	if m.selected == "" {
		fmt.Println("No IDE selected.")
		return
	}

	// Save the selected IDE as the default
	err = setDefaultIDE(m.selected)
	if err != nil {
		fmt.Printf("Warning: Failed to save IDE preference: %v\n", err)
	} else {
		fmt.Printf("IDE preference saved: %s\n", m.selected)
	}

	// Run the selected IDE
	switch m.selected {
	case IDETypeCursor:
		runCursor(cmd, args)
	case IDETypeIntelliJ:
		runIntelliJ(cmd, args)
	case IDETypeJunie:
		runJunie(cmd, args)
	}
}

// runCursor downloads Cursor IDE rules and starts Cursor
func runCursor(cmd *cobra.Command, args []string) {
	fmt.Println("Configuring and starting Cursor IDE...")

	// Get API key from config
	apiKey, err := globals.GetAPIKey()
	if err != nil {
		fmt.Printf("Failed to get API key: %v\n", err)
		return
	}
	if apiKey == "" {
		fmt.Println("API key not found. Please authenticate first with 'devplan auth'")
		return
	}

	// Get rules location from Devplan service
	rulesLocation, err := getRulesLocation(apiKey, "cursor")
	if err != nil {
		fmt.Printf("Failed to get rules location: %v\n", err)
		return
	}

	// Download rules
	rulesPath, err := downloadRules(rulesLocation, "cursor")
	if err != nil {
		fmt.Printf("Failed to download rules: %v\n", err)
		return
	}

	// Get MCP server configurations
	mcpServers, err := getMCPServerConfig(apiKey, "cursor")
	if err != nil {
		fmt.Printf("Warning: Failed to get MCP server configurations: %v\n", err)
		// Continue without MCP server configuration
		mcpServers = nil
	}

	// Initialize Cursor with rules and MCP server configurations
	err = initializeCursor(rulesPath, projectPath, mcpServers)
	if err != nil {
		fmt.Printf("Failed to initialize Cursor: %v\n", err)
		return
	}

	fmt.Println("Cursor IDE configured and started successfully!")
}

// runIntelliJ downloads IntelliJ IDE rules and starts IntelliJ
func runIntelliJ(cmd *cobra.Command, args []string) {
	fmt.Println("Configuring and starting IntelliJ IDE...")

	// Get API key from config
	apiKey, err := globals.GetAPIKey()
	if err != nil {
		fmt.Printf("Failed to get API key: %v\n", err)
		return
	}
	if apiKey == "" {
		fmt.Println("API key not found. Please authenticate first with 'devplan auth'")
		return
	}

	// Get rules location from Devplan service
	rulesLocation, err := getRulesLocation(apiKey, "intellij")
	if err != nil {
		fmt.Printf("Failed to get rules location: %v\n", err)
		return
	}

	// Download rules
	rulesPath, err := downloadRules(rulesLocation, "intellij")
	if err != nil {
		fmt.Printf("Failed to download rules: %v\n", err)
		return
	}

	// Initialize IntelliJ with rules
	err = initializeIntelliJ(rulesPath, projectPath)
	if err != nil {
		fmt.Printf("Failed to initialize IntelliJ: %v\n", err)
		return
	}

	fmt.Println("IntelliJ IDE configured and started successfully!")
}

// runJunie downloads Junie plugin rules, configures the plugin, and starts a compatible JetBrains IDE
func runJunie(cmd *cobra.Command, args []string) {
	fmt.Println("Configuring Junie plugin and starting a compatible JetBrains IDE...")

	// Get API key from config
	apiKey, err := globals.GetAPIKey()
	if err != nil {
		fmt.Printf("Failed to get API key: %v\n", err)
		return
	}
	if apiKey == "" {
		fmt.Println("API key not found. Please authenticate first with 'devplan auth'")
		return
	}

	// Get rules location from Devplan service
	rulesLocation, err := getRulesLocation(apiKey, "junie")
	if err != nil {
		fmt.Printf("Failed to get rules location: %v\n", err)
		return
	}

	// Download rules
	rulesPath, err := downloadRules(rulesLocation, "junie")
	if err != nil {
		fmt.Printf("Failed to download rules: %v\n", err)
		return
	}

	// Get MCP server configurations
	mcpServers, err := getMCPServerConfig(apiKey, "junie")
	if err != nil {
		fmt.Printf("Warning: Failed to get MCP server configurations: %v\n", err)
		// Continue without MCP server configuration
		mcpServers = nil
	}

	// Initialize Junie with rules and MCP server configurations
	err = initializeJunie(rulesPath, projectPath, mcpServers)
	if err != nil {
		fmt.Printf("Failed to configure Junie plugin or start JetBrains IDE: %v\n", err)
		return
	}

	fmt.Println("Junie plugin configured and JetBrains IDE started successfully!")
}

// getRulesLocation gets the rules location from the Devplan service
func getRulesLocation(apiKey, ideType string) (string, error) {
	// Use the base URL from the domain flag
	url := fmt.Sprintf("%s/rules/%s", getBaseURL(), ideType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get rules location with status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		RulesLocation string `json:"rules_location"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.RulesLocation, nil
}

// MCPServerConfig represents the configuration for an MCP server
type MCPServerConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"api_key"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// getMCPServerConfig gets the MCP server configuration from the Devplan service
func getMCPServerConfig(apiKey, ideType string) ([]MCPServerConfig, error) {
	// Use the base URL from the domain flag
	url := fmt.Sprintf("%s/mcp-servers/%s", getBaseURL(), ideType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get MCP server config with status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		MCPServers []MCPServerConfig `json:"mcp_servers"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.MCPServers, nil
}

// downloadRules downloads rules from the specified location
func downloadRules(rulesLocation, ideType string) (string, error) {
	// Create rules directory if it doesn't exist
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	rulesDir := filepath.Join(home, ".devplan", "rules", ideType)
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		err = os.MkdirAll(rulesDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create rules directory: %w", err)
		}
	}

	// Download rules file
	resp, err := http.Get(rulesLocation)
	if err != nil {
		return "", fmt.Errorf("failed to download rules: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download rules with status code: %d", resp.StatusCode)
	}

	rulesPath := filepath.Join(rulesDir, "rules.json")
	out, err := os.Create(rulesPath)
	if err != nil {
		return "", fmt.Errorf("failed to create rules file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write rules file: %w", err)
	}

	return rulesPath, nil
}

// initializeCursor initializes and starts Cursor IDE with the specified rules and MCP server configurations
func initializeCursor(rulesPath, projectPath string, mcpServers []MCPServerConfig) error {
	// Determine Cursor executable path based on OS
	cursorPath := getCursorPath()
	if cursorPath == "" {
		return fmt.Errorf("cursor executable not found")
	}

	// Copy rules to Cursor configuration directory
	cursorConfigDir := getCursorConfigDir()
	if cursorConfigDir == "" {
		return fmt.Errorf("cursor configuration directory not found")
	}

	// Create Cursor config directory if it doesn't exist
	if _, err := os.Stat(cursorConfigDir); os.IsNotExist(err) {
		err = os.MkdirAll(cursorConfigDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create cursor config directory: %w", err)
		}
	}

	// Copy rules file to Cursor config directory
	cursorRulesPath := filepath.Join(cursorConfigDir, "devplan_rules.json")
	input, err := os.ReadFile(rulesPath)
	if err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	err = os.WriteFile(cursorRulesPath, input, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cursor rules file: %w", err)
	}

	// Configure MCP servers if provided
	if len(mcpServers) > 0 {
		err = configureCursorMCPServers(cursorConfigDir, mcpServers)
		if err != nil {
			return fmt.Errorf("failed to configure MCP servers for Cursor: %w", err)
		}
		fmt.Println("MCP servers configured for Cursor IDE")
	}

	// Start Cursor with the project path
	var cmd *exec.Cmd
	if projectPath != "" {
		cmd = exec.Command(cursorPath, projectPath)
	} else {
		cmd = exec.Command(cursorPath)
	}

	// Run Cursor in the background
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start cursor: %w", err)
	}

	return nil
}

// configureCursorMCPServers configures Cursor IDE to use the specified MCP servers
func configureCursorMCPServers(configDir string, mcpServers []MCPServerConfig) error {
	// Cursor typically stores its configuration in a settings.json file
	settingsPath := filepath.Join(configDir, "settings.json")

	// Create or update the settings file
	var settings map[string]interface{}

	// Read existing settings if the file exists
	if _, err := os.Stat(settingsPath); err == nil {
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			return fmt.Errorf("failed to read Cursor settings file: %w", err)
		}

		err = json.Unmarshal(data, &settings)
		if err != nil {
			// If the file exists but is not valid JSON, start with an empty settings object
			settings = make(map[string]interface{})
		}
	} else {
		// If the file doesn't exist, start with an empty settings object
		settings = make(map[string]interface{})
	}

	// Configure MCP servers in settings
	mcpSettings := make([]map[string]interface{}, 0, len(mcpServers))
	for _, server := range mcpServers {
		mcpSetting := map[string]interface{}{
			"name":     server.Name,
			"url":      server.URL,
			"apiKey":   server.APIKey,
			"provider": server.Provider,
			"enabled":  true,
		}
		mcpSettings = append(mcpSettings, mcpSetting)
	}

	// Add MCP server settings to the configuration
	settings["ai.mcpServers"] = mcpSettings

	// Write the updated settings back to the file
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Cursor settings: %w", err)
	}

	err = os.WriteFile(settingsPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write Cursor settings file: %w", err)
	}

	return nil
}

// initializeIntelliJ initializes and starts IntelliJ with the specified rules
func initializeIntelliJ(rulesPath, projectPath string) error {
	// Determine IntelliJ executable path based on OS
	intellijPath := getIntelliJPath()
	if intellijPath == "" {
		return fmt.Errorf("intellij executable not found")
	}

	// Copy rules to IntelliJ configuration directory
	intellijConfigDir := getIntelliJConfigDir()
	if intellijConfigDir == "" {
		return fmt.Errorf("intellij configuration directory not found")
	}

	// Create IntelliJ config directory if it doesn't exist
	if _, err := os.Stat(intellijConfigDir); os.IsNotExist(err) {
		err = os.MkdirAll(intellijConfigDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create intellij config directory: %w", err)
		}
	}

	// Copy rules file to IntelliJ config directory
	intellijRulesPath := filepath.Join(intellijConfigDir, "devplan_rules.json")
	input, err := os.ReadFile(rulesPath)
	if err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	err = os.WriteFile(intellijRulesPath, input, 0644)
	if err != nil {
		return fmt.Errorf("failed to write intellij rules file: %w", err)
	}

	// Start IntelliJ with the project path
	var cmd *exec.Cmd
	if projectPath != "" {
		cmd = exec.Command(intellijPath, projectPath)
	} else {
		cmd = exec.Command(intellijPath)
	}

	// Run IntelliJ in the background
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start intellij: %w", err)
	}

	return nil
}

// initializeJunie initializes and starts a JetBrains IDE with the Junie plugin and specified rules
func initializeJunie(rulesPath, projectPath string, mcpServers []MCPServerConfig) error {
	// Determine JetBrains IDE path based on OS
	jetbrainsIDEPath := getJuniePath()
	if jetbrainsIDEPath == "" {
		return fmt.Errorf("no compatible JetBrains IDE found for Junie plugin")
	}

	// Get the plugins configuration directory
	pluginsConfigDir := getJunieConfigDir()
	if pluginsConfigDir == "" {
		return fmt.Errorf("JetBrains plugins configuration directory not found")
	}

	// Create Junie plugin directory if it doesn't exist
	juniePluginDir := filepath.Join(pluginsConfigDir, "junie")
	if _, err := os.Stat(juniePluginDir); os.IsNotExist(err) {
		err = os.MkdirAll(juniePluginDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create Junie plugin directory: %w", err)
		}
	}

	// Copy rules file to Junie plugin directory
	junieRulesPath := filepath.Join(juniePluginDir, "devplan_rules.json")
	input, err := os.ReadFile(rulesPath)
	if err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	err = os.WriteFile(junieRulesPath, input, 0644)
	if err != nil {
		return fmt.Errorf("failed to write Junie rules file: %w", err)
	}

	// Configure MCP servers if provided
	if len(mcpServers) > 0 {
		err = configureJunieMCPServers(juniePluginDir, mcpServers)
		if err != nil {
			return fmt.Errorf("failed to configure MCP servers for Junie plugin: %w", err)
		}
		fmt.Println("MCP servers configured for Junie plugin")
	}

	fmt.Printf("Junie plugin configured with rules at: %s\n", junieRulesPath)
	fmt.Printf("Starting JetBrains IDE: %s\n", jetbrainsIDEPath)

	// Start the JetBrains IDE with the project path
	// Add parameters to enable the Junie plugin if needed
	var cmd *exec.Cmd
	if projectPath != "" {
		cmd = exec.Command(jetbrainsIDEPath, projectPath)
	} else {
		cmd = exec.Command(jetbrainsIDEPath)
	}

	// Run the JetBrains IDE in the background
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start JetBrains IDE with Junie plugin: %w", err)
	}

	return nil
}

// configureJunieMCPServers configures the Junie plugin to use the specified MCP servers
func configureJunieMCPServers(pluginDir string, mcpServers []MCPServerConfig) error {
	// Junie plugin typically stores its configuration in a config.json file
	configPath := filepath.Join(pluginDir, "config.json")

	// Create or update the config file
	var config map[string]interface{}

	// Read existing config if the file exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read Junie config file: %w", err)
		}

		err = json.Unmarshal(data, &config)
		if err != nil {
			// If the file exists but is not valid JSON, start with an empty config object
			config = make(map[string]interface{})
		}
	} else {
		// If the file doesn't exist, start with an empty config object
		config = make(map[string]interface{})
	}

	// Configure MCP servers in config
	mcpSettings := make([]map[string]interface{}, 0, len(mcpServers))
	for _, server := range mcpServers {
		mcpSetting := map[string]interface{}{
			"name":     server.Name,
			"url":      server.URL,
			"apiKey":   server.APIKey,
			"provider": server.Provider,
			"enabled":  true,
		}
		mcpSettings = append(mcpSettings, mcpSetting)
	}

	// Add MCP server settings to the configuration
	config["mcpServers"] = mcpSettings

	// Write the updated config back to the file
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Junie config: %w", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write Junie config file: %w", err)
	}

	return nil
}

// getCursorPath returns the path to the Cursor executable based on the OS
func getCursorPath() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS
		paths := []string{
			"/Applications/Cursor.app/Contents/MacOS/Cursor",
			"/Applications/Cursor.app/Contents/MacOS/CursorLauncher",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "linux":
		// Linux
		paths := []string{
			"/usr/bin/cursor",
			"/usr/local/bin/cursor",
			"/opt/cursor/cursor",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "windows":
		// Windows
		paths := []string{
			"C:\\Program Files\\Cursor\\Cursor.exe",
			"C:\\Program Files (x86)\\Cursor\\Cursor.exe",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}

// getIntelliJPath returns the path to the IntelliJ executable based on the OS
func getIntelliJPath() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS
		paths := []string{
			"/Applications/IntelliJ IDEA.app/Contents/MacOS/idea",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "linux":
		// Linux
		paths := []string{
			"/usr/bin/intellij-idea-ultimate",
			"/usr/local/bin/intellij-idea-ultimate",
			"/opt/intellij-idea/bin/idea.sh",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "windows":
		// Windows
		paths := []string{
			"C:\\Program Files\\JetBrains\\IntelliJ IDEA\\bin\\idea64.exe",
			"C:\\Program Files (x86)\\JetBrains\\IntelliJ IDEA\\bin\\idea.exe",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}

// getJuniePath returns the path to a compatible JetBrains IDE for Junie plugin
func getJuniePath() string {
	// Junie is a plugin for JetBrains IDEs, not a standalone application
	// We need to find an installed JetBrains IDE to use with the Junie plugin
	switch runtime.GOOS {
	case "darwin":
		// macOS - Check for various JetBrains IDEs
		paths := []string{
			"/Applications/IntelliJ IDEA.app/Contents/MacOS/idea",
			"/Applications/PyCharm.app/Contents/MacOS/pycharm",
			"/Applications/WebStorm.app/Contents/MacOS/webstorm",
			"/Applications/CLion.app/Contents/MacOS/clion",
			"/Applications/GoLand.app/Contents/MacOS/goland",
			"/Applications/PhpStorm.app/Contents/MacOS/phpstorm",
			"/Applications/Rider.app/Contents/MacOS/rider",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "linux":
		// Linux - Check for various JetBrains IDEs
		paths := []string{
			// IntelliJ IDEA
			"/usr/bin/intellij-idea-ultimate",
			"/usr/local/bin/intellij-idea-ultimate",
			"/opt/intellij-idea/bin/idea.sh",
			// PyCharm
			"/usr/bin/pycharm",
			"/usr/local/bin/pycharm",
			"/opt/pycharm/bin/pycharm.sh",
			// WebStorm
			"/usr/bin/webstorm",
			"/usr/local/bin/webstorm",
			"/opt/webstorm/bin/webstorm.sh",
			// Other JetBrains IDEs
			"/usr/bin/clion",
			"/usr/bin/goland",
			"/usr/bin/phpstorm",
			"/usr/bin/rider",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "windows":
		// Windows - Check for various JetBrains IDEs
		paths := []string{
			// IntelliJ IDEA
			"C:\\Program Files\\JetBrains\\IntelliJ IDEA\\bin\\idea64.exe",
			"C:\\Program Files (x86)\\JetBrains\\IntelliJ IDEA\\bin\\idea.exe",
			// PyCharm
			"C:\\Program Files\\JetBrains\\PyCharm\\bin\\pycharm64.exe",
			"C:\\Program Files (x86)\\JetBrains\\PyCharm\\bin\\pycharm.exe",
			// WebStorm
			"C:\\Program Files\\JetBrains\\WebStorm\\bin\\webstorm64.exe",
			"C:\\Program Files (x86)\\JetBrains\\WebStorm\\bin\\webstorm.exe",
			// Other JetBrains IDEs
			"C:\\Program Files\\JetBrains\\CLion\\bin\\clion64.exe",
			"C:\\Program Files\\JetBrains\\GoLand\\bin\\goland64.exe",
			"C:\\Program Files\\JetBrains\\PhpStorm\\bin\\phpstorm64.exe",
			"C:\\Program Files\\JetBrains\\Rider\\bin\\rider64.exe",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}

// getCursorConfigDir returns the path to the Cursor configuration directory based on the OS
func getCursorConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS
		return filepath.Join(home, "Library", "Application Support", "Cursor", "User")
	case "linux":
		// Linux
		return filepath.Join(home, ".config", "cursor")
	case "windows":
		// Windows
		return filepath.Join(home, "AppData", "Roaming", "Cursor", "User")
	}
	return ""
}

// getIntelliJConfigDir returns the path to the IntelliJ configuration directory based on the OS
func getIntelliJConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS
		return filepath.Join(home, "Library", "Application Support", "JetBrains", "IntelliJIdea2023.2")
	case "linux":
		// Linux
		return filepath.Join(home, ".config", "JetBrains", "IntelliJIdea2023.2")
	case "windows":
		// Windows
		return filepath.Join(home, "AppData", "Roaming", "JetBrains", "IntelliJIdea2023.2")
	}
	return ""
}

// getJunieConfigDir returns the path to the JetBrains plugins configuration directory based on the OS
func getJunieConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Junie is a plugin for JetBrains IDEs, so we need to find the plugins directory
	// for the JetBrains IDE that's being used
	switch runtime.GOOS {
	case "darwin":
		// macOS - JetBrains plugins are typically stored in:
		// ~/Library/Application Support/JetBrains/<IDE-VERSION>/plugins
		// We'll use a common location for all JetBrains IDEs
		return filepath.Join(home, "Library", "Application Support", "JetBrains", "plugins")
	case "linux":
		// Linux - JetBrains plugins are typically stored in:
		// ~/.config/JetBrains/<IDE-VERSION>/plugins
		// We'll use a common location for all JetBrains IDEs
		return filepath.Join(home, ".config", "JetBrains", "plugins")
	case "windows":
		// Windows - JetBrains plugins are typically stored in:
		// %APPDATA%\JetBrains\<IDE-VERSION>\plugins
		// We'll use a common location for all JetBrains IDEs
		return filepath.Join(home, "AppData", "Roaming", "JetBrains", "plugins")
	}
	return ""
}
