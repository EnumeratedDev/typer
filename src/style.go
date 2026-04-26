package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"gopkg.in/yaml.v3"
)

type TyperStyle struct {
	// Metadata
	Name        string
	Description string
	StyleType   string

	// Colors
	BufferAreaBg  tcell.Color `name:"buffer_area_bg"`
	BufferAreaFg  tcell.Color `name:"buffer_area_fg"`
	BufferAreaSel tcell.Color `name:"buffer_area_sel"`
	TopMenuBg     tcell.Color `name:"top_menu_bg"`
	TopMenuFg     tcell.Color `name:"top_menu_fg"`
	DropdownBg    tcell.Color `name:"dropdown_bg"`
	DropdownFg    tcell.Color `name:"dropdown_fg"`
	DropdownSel   tcell.Color `name:"dropdown_sel"`
	LineIndexBg   tcell.Color `name:"line_index_bg"`
	LineIndexFg   tcell.Color `name:"line_index_fg"`
	MessageBarBg  tcell.Color `name:"message_bar_bg"`
	MessageBarFg  tcell.Color `name:"message_bar_fg"`
	InputBarBg    tcell.Color `name:"input_bar_bg"`
	InputBarFg    tcell.Color `name:"input_bar_fg"`
}

type typerStyleYaml struct {
	// Metadata
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	StyleType   string `yaml:"style_type"`

	// Colors
	Colors map[string]string `yaml:"colors"`
}

var FallbackStyle = TyperStyle{
	Name:        "fallback",
	Description: "Fallback style",
	StyleType:   "8-color",

	BufferAreaBg:  tcell.ColorBlack,
	BufferAreaFg:  tcell.ColorWhite,
	BufferAreaSel: tcell.ColorNavy,
	TopMenuBg:     tcell.ColorWhite,
	TopMenuFg:     tcell.ColorBlack,
	DropdownBg:    tcell.ColorWhite,
	DropdownFg:    tcell.ColorBlack,
	DropdownSel:   tcell.ColorNavy,
	LineIndexBg:   tcell.ColorWhite,
	LineIndexFg:   tcell.ColorBlack,
	MessageBarBg:  tcell.ColorWhite,
	MessageBarFg:  tcell.ColorBlack,
	InputBarBg:    tcell.ColorWhite,
	InputBarFg:    tcell.ColorBlack,
}

var AvailableStyles = make(map[string]TyperStyle)
var CurrentStyle = FallbackStyle

func readStyles() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not get home directory: %s", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Could not get path to executable: %s", err)
	}

	stylesPaths := make([]string, 0)
	if runtime.GOOS == "windows" {
		stylesPaths = append(stylesPaths, path.Join(homeDir, "AppData/Roaming/Typer/styles"))
		stylesPaths = append(stylesPaths, path.Join(path.Dir(execPath), "etc/typer/styles"))
	} else {
		stylesPaths = append(stylesPaths, path.Join(homeDir, ".config/typer/styles"))
		stylesPaths = append(stylesPaths, path.Join(sysconfdir, "typer/styles"))
	}

	for _, stylesPath := range stylesPaths {
		// Ensure directory exists at path
		if stat, err := os.Stat(stylesPath); err != nil || !stat.IsDir() {
			fmt.Println(stylesPath)
			continue
		}

		// Get directory entries
		entries, err := os.ReadDir(stylesPath)
		if err != nil {
			log.Fatalf("Could not read user style directory: %s", err)
		}

		// Read entries in directory
		for _, entry := range entries {
			entryPath := path.Join(stylesPath, entry.Name())

			style, err := readStyleYamlFile(entryPath)
			if err != nil {
				log.Fatalf("Could not read style file (%s): %s", entryPath, err)
			}

			if _, ok := AvailableStyles[style.Name]; !ok {
				AvailableStyles[style.Name] = style
			}
		}
	}
}

func readStyleYamlFile(filepath string) (TyperStyle, error) {
	styleYaml := typerStyleYaml{}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return TyperStyle{}, fmt.Errorf("could not read file: %s", err)
	}
	err = yaml.Unmarshal(data, &styleYaml)
	if err != nil {
		return TyperStyle{}, fmt.Errorf("could not unmarshal style: %s", err)
	}

	style := TyperStyle{
		Name:        styleYaml.Name,
		Description: styleYaml.Description,
		StyleType:   styleYaml.StyleType,
	}

	for name, colorStr := range styleYaml.Colors {
		var color tcell.Color

		if n, err := strconv.Atoi(colorStr); err == nil && n >= 0 && n < 256 {
			color = tcell.ColorValid + tcell.Color(n)
		} else if strings.HasPrefix(colorStr, "#") && len(colorStr) == 7 {
			n, err := strconv.ParseInt(colorStr[1:], 16, 32)
			if err != nil {
				return TyperStyle{}, fmt.Errorf("could not parse color (%s): %s", colorStr, err)
			}

			color = tcell.NewHexColor(int32(n))
		} else if c, ok := tcell.ColorNames[colorStr]; ok {
			color = c
		} else {
			return TyperStyle{}, fmt.Errorf("could not parse color (%s): %s", colorStr, err)
		}

		pt := reflect.TypeOf(&style)
		t := pt.Elem()
		pv := reflect.ValueOf(&style)
		v := pv.Elem()

		for i := 0; i < t.NumField(); i++ {
			field := v.Field(i)

			if tag, ok := t.Field(i).Tag.Lookup("name"); ok && tag == name {
				field.Set(reflect.ValueOf(color))
			}
		}
	}

	return style, nil
}

func SetCurrentStyle(screen tcell.Screen, styleName string) bool {
	availableTypes := make([]string, 1)
	availableTypes[0] = "8-color"
	if screen.Colors() >= 16 {
		availableTypes = append(availableTypes, "16-color")
	}
	if screen.Colors() >= 256 {
		availableTypes = append(availableTypes, "256-color")
	}
	if screen.Colors() >= 16777216 {
		availableTypes = append(availableTypes, "true-color")
	}

	if style, ok := AvailableStyles[styleName]; ok && slices.Index(availableTypes, style.StyleType) != -1 {
		CurrentStyle = style

		screen.SetStyle(tcell.StyleDefault.Foreground(CurrentStyle.BufferAreaFg).Background(CurrentStyle.BufferAreaBg))
		screen.Sync()

		return true
	}
	return false
}
