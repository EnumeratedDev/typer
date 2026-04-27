package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type TyperConfig struct {
	SelectedStyle     string `yaml:"selected_style,omitempty"`
	FallbackStyle     string `yaml:"fallback_style,omitempty"`
	ShowTopMenu       bool   `yaml:"show_top_menu,omitempty"`
	ShowLineIndex     bool   `yaml:"show_line_index,omitempty"`
	ExtendLineIndex   bool   `yaml:"extend_line_index,omitempty"`
	BufferInfoMessage string `yaml:"buffer_info_message,omitempty"`
	TabIndentation    int    `yaml:"tab_indentation,omitempty"`
}

var Config TyperConfig

func readConfig() {
	Config = TyperConfig{
		SelectedStyle:     "default",
		FallbackStyle:     "default-fallback",
		ShowTopMenu:       true,
		ShowLineIndex:     true,
		ExtendLineIndex:   false,
		BufferInfoMessage: "File: %f Cursor: (%x, %y, %p) Chars: %c",
		TabIndentation:    4,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not get home directory: %s", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Could not get path to executable: %s", err)
	}

	configPaths := make([]string, 0)
	switch runtime.GOOS {
	case "windows":
		configPaths = append(configPaths, filepath.Join(homeDir, "AppData/Roaming/Typer/config.yml"))
		configPaths = append(configPaths, filepath.Join(filepath.Dir(execPath), "etc/typer/config.yml"))
	case "darwin":
		configPaths = append(configPaths, filepath.Join(homeDir, "Library/Typer/config.yml"))
		configPaths = append(configPaths, "/Library/Typer/config.yml")
		configPaths = append(configPaths, filepath.Join(sysconfdir, "typer/config.yml"))
	default:
		configPaths = append(configPaths, filepath.Join(homeDir, ".config/typer/config.yml"))
		configPaths = append(configPaths, filepath.Join(sysconfdir, "typer/config.yml"))
	}

	for _, configPath := range configPaths {
		// Ensure config exists at path
		if _, err := os.Stat(configPath); err != nil {
			continue
		}

		// Read config file
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Could not read config.yml: %s", err)
		}

		// Unmarshal contents into struct
		err = yaml.Unmarshal(data, &Config)
		if err != nil {
			log.Fatalf("Could not unmarshal config.yml: %s", err)
		}

		break
	}

	// Validate config options
	if Config.TabIndentation < 1 {
		Config.TabIndentation = 1
	}
}
