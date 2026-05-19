package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type TyperConfig struct {
	SelectedStyle     string `yaml:"selected_style,omitempty"`
	FallbackStyle     string `yaml:"fallback_style,omitempty"`
	ShowTopMenu       bool   `yaml:"show_top_menu,omitempty"`
	ShowLineIndex     bool   `yaml:"show_line_index,omitempty"`
	ColorMessageBar   bool   `yaml:"color_message_bar"`
	ExtendLineIndex   bool   `yaml:"extend_line_index,omitempty"`
	BufferInfoMessage string `yaml:"buffer_info_message,omitempty"`
	TabIndentation    int    `yaml:"tab_indentation,omitempty"`
}

var Config TyperConfig

func readMainConfig() {
	Config = TyperConfig{
		SelectedStyle:     "default",
		FallbackStyle:     "default-fallback",
		ShowTopMenu:       true,
		ShowLineIndex:     true,
		ExtendLineIndex:   false,
		BufferInfoMessage: "File: %f Cursor: (%x, %y, %p) Chars: %c",
		TabIndentation:    4,
	}

	// Get main config path
	mainConfigPath := GetConfigPath("config.yml")

	// Ensure config exists at path
	if mainConfigPath == "" {
		log.Fatalf("config.yml not found in any config directory")
	}

	// Read config file
	data, err := os.ReadFile(mainConfigPath)
	if err != nil {
		log.Fatalf("Could not read config.yml: %s", err)
	}

	// Unmarshal contents into struct
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Fatalf("Could not unmarshal config.yml: %s", err)
	}

	// Validate config options
	if Config.TabIndentation < 1 {
		Config.TabIndentation = 1
	}
}
