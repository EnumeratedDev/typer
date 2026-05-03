package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/gdamore/tcell/v2"
)

func GetConfigPath(relativeConfigPath string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not get home directory: %s", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Could not get path to executable: %s", err)
	}

	paths := make([]string, 0)
	if *configDirFlag != "" {
		paths = append(paths, path.Join(*configDirFlag, relativeConfigPath))
	}
	switch runtime.GOOS {
	case "windows":
		paths = append(paths, filepath.Join(homeDir, "AppData/Roaming/Typer", relativeConfigPath))
		paths = append(paths, "C:/ProgramData/Typer", relativeConfigPath)
	case "darwin":
		paths = append(paths, filepath.Join(homeDir, "Library/Typer", relativeConfigPath))
		paths = append(paths, filepath.Join(homeDir, "Library/typer", relativeConfigPath))
		paths = append(paths, filepath.Join(sysconfdir, "Typer", relativeConfigPath))
		paths = append(paths, filepath.Join(sysconfdir, "typer", relativeConfigPath))
	default:
		paths = append(paths, filepath.Join(homeDir, ".config/typer", relativeConfigPath))
		paths = append(paths, filepath.Join(sysconfdir, "typer", relativeConfigPath))
	}
	paths = append(paths, filepath.Join(filepath.Dir(execPath), "config", relativeConfigPath))

	for _, p := range paths {
		// Return true if path exists
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			s.SetContent(col, row, ' ', nil, style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, tcell.RuneHLine, nil, style)
		s.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, tcell.RuneVLine, nil, style)
		s.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		s.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		s.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		s.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		s.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}

	drawText(s, x1+1, y1+1, x2-1, y2-1, style, " ")
}

func DeleteFromSlice[T any](slice []T, i int) []T {
	if i >= len(slice) {
		return slice
	} else if i < 0 {
		return slice
	} else if i == len(slice)-1 {
		return slice[:len(slice)-1]
	} else {
		return append(slice[:i], slice[i+1:]...)
	}
}
