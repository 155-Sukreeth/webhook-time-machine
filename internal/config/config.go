package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/155-Sukreeth/webhook-time-machine/internal/models"
	"gopkg.in/yaml.v3"
)

// DefaultConfig returns baseline production config settings.
func DefaultConfig() *models.Config {
	return &models.Config{
		Port:            8080,
		UIPort:          8081,
		ForwardURL:      "http://localhost:3000",
		DBPath:          "./wtm.db",
		LogLevel:        "info",
		StripSignatures: true,
	}
}

// LoadConfig loads config adhering strictly to precedence: CLI Flags > ENV > YAML > Defaults.
func LoadConfig(cfgFilePath string) (*models.Config, error) {
	v := viper.New()
	defaults := DefaultConfig()

	v.SetDefault("port", defaults.Port)
	v.SetDefault("ui_port", defaults.UIPort)
	v.SetDefault("forward_url", defaults.ForwardURL)
	v.SetDefault("db_path", defaults.DBPath)
	v.SetDefault("log_level", defaults.LogLevel)
	v.SetDefault("strip_signatures", defaults.StripSignatures)

	v.SetEnvPrefix("WTM")
	v.AutomaticEnv()

	if cfgFilePath != "" {
		v.SetConfigFile(cfgFilePath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
		v.SetConfigName(".wtm")
		v.SetConfigType("yaml")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg models.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed unmarshaling config: %w", err)
	}
	return &cfg, nil
}

// WriteDefaultConfigFile generates default .wtm.yaml file.
func WriteDefaultConfigFile(targetPath string) error {
	cfg := DefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed marshaling default config: %w", err)
	}

	header := []byte("# Local Webhook Time Machine (.wtm.yaml)\n# Precedence: CLI Flags > Environment Variables > YAML Config > Defaults\n\n")
	content := append(header, data...)

	dir := filepath.Dir(targetPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed creating directory: %w", err)
		}
	}

	return os.WriteFile(targetPath, content, 0644)
}
