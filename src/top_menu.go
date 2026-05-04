package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type TopMenuButton struct {
	Name   string
	Action func(w *Window)
}

var TopMenuButtons = make([]TopMenuButton, 0)

func initTopMenu() {
	// Buttons
	fileButton := TopMenuButton{
		Name: "File",
		Action: func(window *Window) {
			ClearDropdowns()

			y := 0
			if window.ShowTopMenu {
				y++
			}

			d := CreateDropdownMenu([]string{"New", "Save", "Open", "Close", "Quit"}, 0, y, 0, func(i int) {
				switch i {
				case 0:
					RunCommand(window, "new-buffer")
				case 1:
					RunCommand(window, "save")
				case 2:
					RunCommand(window, "open")
				case 3:
					RunCommand(window, "close-buffer")
				case 4:
					RunCommand(window, "quit")
				}
				ClearDropdowns()
			})
			ActiveDropdown = d
			window.CursorMode = CursorModeDropdown
		},
	}
	EditButton := TopMenuButton{
		Name: "Edit",
		Action: func(window *Window) {
			ClearDropdowns()

			y := 0
			if window.ShowTopMenu {
				y++
			}

			d := CreateDropdownMenu([]string{"Cut", "Copy", "Paste"}, 0, y, 0, func(i int) {
				switch i {
				case 0:
					RunCommand(window, "cut")
				case 1:
					RunCommand(window, "copy")
				case 2:
					RunCommand(window, "paste")
				}
				ClearDropdowns()
				window.CursorMode = CursorModeBuffer
			})
			ActiveDropdown = d
			window.CursorMode = CursorModeDropdown
		},
	}
	Buffers := TopMenuButton{
		Name: "Buffers",
		Action: func(window *Window) {
			ClearDropdowns()

			y := 0
			if window.ShowTopMenu {
				y++
			}

			buffersSlice := make([]string, 0)
			for i, buffer := range Buffers {
				if window.CurrentBuffer == buffer {
					buffersSlice = append(buffersSlice, fmt.Sprintf("[%d] * %s", i+1, buffer.Name))
				} else {
					buffersSlice = append(buffersSlice, fmt.Sprintf("[%d] %s", i+1, buffer.Name))
				}
			}

			d := CreateDropdownMenu(buffersSlice, 0, y, 0, func(i int) {
				window.CurrentBuffer = Buffers[i]
				window.PrintMessage(fmt.Sprintf("Set current buffer to '%s'", window.CurrentBuffer.Name), TYPER_MESSAGE_INFO)
				ClearDropdowns()
				window.CursorMode = CursorModeBuffer
			})
			ActiveDropdown = d
			window.CursorMode = CursorModeDropdown
		},
	}

	// Append buttons
	TopMenuButtons = append(TopMenuButtons, fileButton, EditButton, Buffers)
}

func drawTopMenu(window *Window) {
	screen := window.screen

	topMenuStyle := tcell.StyleDefault.Background(CurrentStyle.TopMenuBg).Foreground(CurrentStyle.TopMenuFg)

	sizeX, _ := screen.Size()

	for x := 0; x < sizeX; x++ {
		screen.SetContent(x, 0, ' ', nil, topMenuStyle)
	}

	currentX := 1
	for _, button := range TopMenuButtons {
		drawText(screen, currentX, 0, currentX+len(button.Name), 0, topMenuStyle, button.Name)
		currentX += len(button.Name) + 1
	}

	// Draw buffer info
	bufferInfoMsg := getBufferInfoMsg(window)
	if sizeX-len(bufferInfoMsg)-1 > currentX+2 {
		drawText(screen, sizeX-len(bufferInfoMsg)-1, 0, sizeX-1, 0, topMenuStyle, bufferInfoMsg)
	}
}

func getBufferInfoMsg(window *Window) string {
	pathToFile := "Not set"
	filename := "Not set"
	filetype := "Not set"
	if window.CurrentBuffer.filename != "" {
		pathToFile = window.CurrentBuffer.filename
	}
	if filepath.Base(window.CurrentBuffer.filename) != "." {
		filename = filepath.Base(window.CurrentBuffer.filename)
	}

	if window.CurrentBuffer.filetype != "" {
		filetype = window.CurrentBuffer.filetype
	}

	contents := window.CurrentBuffer.GetContentsAsString()
	chars := len(contents)
	words := len(strings.Fields(string(contents)))

	ret := Config.BufferInfoMessage

	ret = strings.ReplaceAll(ret, "\n", " ")
	ret = strings.ReplaceAll(ret, "%F", pathToFile)
	ret = strings.ReplaceAll(ret, "%f", filename)
	ret = strings.ReplaceAll(ret, "%t", filetype)
	ret = strings.ReplaceAll(ret, "%x", strconv.Itoa(window.CurrentBuffer.CursorPos.X+1))
	ret = strings.ReplaceAll(ret, "%y", strconv.Itoa(window.CurrentBuffer.CursorPos.Y+1))
	ret = strings.ReplaceAll(ret, "%p", strconv.Itoa(window.CurrentBuffer.PositionToAbsolutePosition(window.CurrentBuffer.CursorPos)+1))
	ret = strings.ReplaceAll(ret, "%c", strconv.Itoa(chars))
	ret = strings.ReplaceAll(ret, "%w", strconv.Itoa(words))

	return ret
}
