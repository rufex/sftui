package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rufex/sftui/internal/models"
)

type ConfigManager struct{}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

func (c *ConfigManager) getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(homeDir, ".silverfin", "config.json")
	
	// Check if file exists, if not use fixture
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "fixtures/silverfin/config.json", nil
	}
	
	return configPath, nil
}

func (c *ConfigManager) LoadSilverfinConfig() (string, string, string) {
	firm := "No firm set"
	host := "No host set"
	output := "Ready"

	configPath, err := c.getConfigPath()
	if err != nil {
		return firm, host, "Error getting config path"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return firm, host, "No Silverfin config found"
	}

	var config models.SilverfinConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return firm, host, "Error parsing Silverfin config"
	}

	host = config.Host

	// Get current working directory name to find firm
	cwd, err := os.Getwd()
	if err != nil {
		return firm, host, output
	}

	repoName := filepath.Base(cwd)
	if firmID, exists := config.DefaultFirmIDs[repoName]; exists {
		// Parse the config again to get firm details
		var rawConfig map[string]interface{}
		json.Unmarshal(data, &rawConfig)

		if firmData, exists := rawConfig[firmID].(map[string]interface{}); exists {
			if firmName, exists := firmData["firmName"].(string); exists {
				firm = fmt.Sprintf("%s (%s)", firmName, firmID)
			}
		}
	}

	return firm, host, output
}

func (c *ConfigManager) LoadFirmOptions() ([]models.FirmOption, error) {
	configPath, err := c.getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil, err
	}

	var firmOptions []models.FirmOption

	// Add firms
	for key, value := range rawConfig {
		if firmData, ok := value.(map[string]interface{}); ok {
			if firmName, exists := firmData["firmName"].(string); exists {
				firmOptions = append(firmOptions, models.FirmOption{
					ID:   key,
					Name: firmName,
					Type: "firm",
				})
			}
		}
	}

	// Add partners
	if partnerCreds, exists := rawConfig["partnerCredentials"].(map[string]interface{}); exists {
		for partnerID, partnerData := range partnerCreds {
			if partner, ok := partnerData.(map[string]interface{}); ok {
				if partnerName, exists := partner["name"].(string); exists {
					firmOptions = append(firmOptions, models.FirmOption{
						ID:   partnerID,
						Name: partnerName,
						Type: "partner",
					})
				}
			}
		}
	}

	return firmOptions, nil
}

func (c *ConfigManager) SetDefaultFirm(firmID string) error {
	configPath, err := c.getConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return err
	}

	// Get current working directory name
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	repoName := filepath.Base(cwd)

	// Ensure defaultFirmIDs exists
	if _, exists := rawConfig["defaultFirmIDs"]; !exists {
		rawConfig["defaultFirmIDs"] = make(map[string]interface{})
	}

	// Set the default firm for this repository
	defaultFirmIDs := rawConfig["defaultFirmIDs"].(map[string]interface{})
	defaultFirmIDs[repoName] = firmID

	// Write back to file
	newData, err := json.MarshalIndent(rawConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, newData, 0644)
}

func (c *ConfigManager) SetHost(host string) error {
	configPath, err := c.getConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return err
	}

	// Set the host
	rawConfig["host"] = host

	// Write back to file
	newData, err := json.MarshalIndent(rawConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, newData, 0644)
}
