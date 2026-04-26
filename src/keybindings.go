package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"gopkg.in/yaml.v3"
)

type TyperKeybindings struct {
	Keybindings []Keybinding `yaml:"keybindings"`
}

type Keybinding struct {
	Keybinding  string   `yaml:"keybinding"`
	CursorModes []string `yaml:"cursor_modes"`
	Command     string   `yaml:"command"`
}

var Keybindings TyperKeybindings

func readKeybindings() {
	Keybindings = TyperKeybindings{
		Keybindings: make([]Keybinding, 0),
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
	if runtime.GOOS == "windows" {
		configPaths = append(configPaths, filepath.Join(homeDir, "AppData/Roaming/Typer/keybindings.yml"))
		configPaths = append(configPaths, filepath.Join(filepath.Dir(execPath), "etc/typer/keybindings.yml"))
	} else {
		configPaths = append(configPaths, filepath.Join(homeDir, ".config/typer/keybindings.yml"))
		configPaths = append(configPaths, filepath.Join(sysconfdir, "typer/keybindings.yml"))
	}

	for _, configPath := range configPaths {
		// Ensure config exists at path
		if _, err := os.Stat(configPath); err != nil {
			continue
		}

		// Read config file
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Could not read keybindings.yml: %s", err)
		}

		// Unmarshal contents into struct
		err = yaml.Unmarshal(data, &Keybindings)
		if err != nil {
			log.Fatalf("Could not unmarshal keybindings.yml: %s", err)
		}

		break
	}
}

func (keybinding *Keybinding) GetCursorModes() []CursorMode {
	ret := make([]CursorMode, 0)

	for _, cursorModeStr := range keybinding.CursorModes {
		for key, value := range CursorModeNames {
			if cursorModeStr == value {
				ret = append(ret, key)
			}
		}
	}

	return ret
}

func (keybinding *Keybinding) IsPressed(ev *tcell.EventKey) bool {
	keys := strings.SplitN(keybinding.Keybinding, "+", 2)

	if len(keys) == 0 {
		return false
	} else if len(keys) == 1 {
		for k, v := range tcell.KeyNames {
			if k != tcell.KeyRune {
				if keybinding.Keybinding == v {
					if ev.Key() == k {
						return true
					}
				}
			} else {
				if keybinding.Keybinding == string(ev.Rune()) {
					return true
				}
			}
		}
	} else {
		modKey := keys[0]
		key := keys[1]

		switch modKey {
		case "Shift":
			if ev.Modifiers() != tcell.ModShift {
				return false
			}
		case "Alt":
			if ev.Modifiers() != tcell.ModAlt {
				return false
			}
		case "Ctrl":
			if ev.Modifiers() != tcell.ModCtrl {
				return false
			}
		case "Meta":
			if ev.Modifiers() != tcell.ModMeta {
				return false
			}
		}

		for k, v := range tcell.KeyNames {
			if k != tcell.KeyRune {
				if key == v {
					if ev.Key() == k {
						return true
					}
				}
			}
		}

		if strings.ToLower(key) == string(ev.Rune()) {
			return true
		}
	}

	return false
}
