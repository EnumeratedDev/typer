package main

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"typer/runestring"
)

type Command struct {
	cmd          string
	run          func(window *Window, args ...string)
	autocomplete func(window *Window, args ...string) []string
}

var commands = make(map[string]*Command)

func initCommands() {
	// Setup commands
	selectAll := Command{
		cmd: "select-all",
		run: func(window *Window, args ...string) {
			// Select entire buffer content
			lastLine := window.CurrentBuffer.Contents[len(window.CurrentBuffer.Contents)-1]
			window.CurrentBuffer.Selection = &Selection{
				selectionStart: Position{0, 0},
				selectionEnd:   Position{len(window.CurrentBuffer.Contents) - 1, len(lastLine) - 1},
			}

			window.PrintMessage("Selected all text", TYPER_MESSAGE_INFO)
		},
	}

	cutCmd := Command{
		cmd: "cut",
		run: func(window *Window, args ...string) {
			// Cut text from buffer
			copiedText, copyingMethod := window.CurrentBuffer.CutText(window)

			// Put cut text to clipboard
			window.Clipboard = copiedText

			// Send appropriate message and remove text depending on copying method
			if copyingMethod == 0 {
				window.PrintMessage("Copied line to clipboard", TYPER_MESSAGE_INFO)
			} else {
				window.PrintMessage("Copied selection to clipboard", TYPER_MESSAGE_INFO)
			}
		},
	}

	copyCmd := Command{
		cmd: "copy",
		run: func(window *Window, args ...string) {
			// Copy text from buffer
			copiedText, copyingMethod := window.CurrentBuffer.CopyText()

			// Put copied text to clipboard
			window.Clipboard = copiedText

			// Send appropriate message depending on copying method
			if copyingMethod == 0 {
				window.PrintMessage("Copied line to clipboard", TYPER_MESSAGE_INFO)
			} else {
				window.PrintMessage("Copied selection to clipboard", TYPER_MESSAGE_INFO)
			}
		},
	}

	pasteCmd := Command{
		cmd: "paste",
		run: func(window *Window, args ...string) {
			if !window.CurrentBuffer.canEdit {
				window.PrintMessage(fmt.Sprintf("Buffer '%s' is read-only", window.CurrentBuffer.Name), TYPER_MESSAGE_WARNING)
				return
			}

			if len(window.Clipboard) != 0 {
				window.CurrentBuffer.PasteText(window, window.Clipboard)
				window.PrintMessage("Pasted text to buffer", TYPER_MESSAGE_INFO)
			}
		},
	}

	saveCmd := Command{
		cmd: "save",
		run: func(window *Window, args ...string) {
			if !window.CurrentBuffer.canSave {
				window.PrintMessage("Cannot save buffer", TYPER_MESSAGE_ERROR)
				return
			}

			inputChannel := RequestInput(window, "Save file [y\\N]:", "")
			go func() {
				input := <-inputChannel

				if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
					return
				}

				inputChannel = RequestInput(window, "Save buffer to:", window.CurrentBuffer.filename)

				input = <-inputChannel

				if strings.TrimSpace(input) == "" {
					window.PrintMessage("No save location was given", TYPER_MESSAGE_ERROR)
					return
				}

				window.CurrentBuffer.filename = strings.TrimSpace(input)
				err := window.CurrentBuffer.Save()
				if err != nil {

					window.PrintMessage(fmt.Sprintf("Could not save file: %s", err), TYPER_MESSAGE_ERROR)
					window.CurrentBuffer.filename = ""
					return
				}

				window.PrintMessage("File saved", TYPER_MESSAGE_INFO)
			}()
		},
		autocomplete: func(window *Window, args ...string) []string {
			return nil
		},
	}

	openCmd := Command{
		cmd: "open",
		run: func(window *Window, args ...string) {
			inputChannel := RequestInput(window, "File to open:", "")
			go func() {
				input := <-inputChannel

				if input == "" {
					return
				}

				if openBuffer := GetOpenFileBuffer(input); openBuffer != nil {
					window.PrintMessage(fmt.Sprintf("File already open! Switching to buffer: %s", openBuffer.Name), TYPER_MESSAGE_INFO)
					window.CurrentBuffer = openBuffer
				} else {
					newBuffer, err := CreateFileBuffer(input)
					if err != nil {
						window.PrintMessage(fmt.Sprintf("Could not open file %s: %s", input, err), TYPER_MESSAGE_ERROR)
						return
					}

					window.PrintMessage(fmt.Sprintf("Opening file at: %s", newBuffer.filename), TYPER_MESSAGE_INFO)
					window.CurrentBuffer = newBuffer
				}
			}()
		},
	}

	reloadCmd := Command{
		cmd: "reload",
		run: func(window *Window, args ...string) {
			err := window.CurrentBuffer.Load()
			if err != nil {
				log.Fatalf("Could not reload buffer: %s", err)
			}

			window.PrintMessage("Buffer reloaded", TYPER_MESSAGE_INFO)
		},
	}

	findCmd := Command{
		cmd: "find",
		run: func(window *Window, args ...string) {
			if len(args) >= 1 {
				input := runestring.RuneString(args[0])

				if len(input) == 0 {
					return
				}

				pos := window.CurrentBuffer.FindSubstring(input, window.CurrentBuffer.CursorPos)
				if pos.X >= 0 && pos.Y >= 0 {
					window.CurrentBuffer.CursorPos = pos
					window.PrintMessage("Match found", TYPER_MESSAGE_INFO)
				} else {
					window.PrintMessage(fmt.Sprintf("'%s' not found in buffer", string(input)), TYPER_MESSAGE_WARNING)
				}

				return
			}

			inputChannel := RequestInput(window, "Substring to search for:", "")
			go func() {
				input := runestring.RuneString(<-inputChannel)

				if len(input) == 0 {
					return
				}

				pos := window.CurrentBuffer.FindSubstring(input, window.CurrentBuffer.CursorPos)
				if pos.X >= 0 && pos.Y >= 0 {
					window.CurrentBuffer.CursorPos = pos
					window.PrintMessage("Match found", TYPER_MESSAGE_INFO)
				} else {
					window.PrintMessage(fmt.Sprintf("'%s' not found in buffer", string(input)), TYPER_MESSAGE_WARNING)
				}
			}()
		},
	}

	replaceCmd := Command{
		cmd: "replace",
		run: func(window *Window, args ...string) {
			if len(args) >= 2 {
				findStr := runestring.RuneString(args[0])
				replaceStr := runestring.RuneString(args[1])

				if len(findStr) == 0 {
					return
				}

				pos := window.CurrentBuffer.FindAndReplaceSubstring(findStr, replaceStr, window.CurrentBuffer.CursorPos)
				if pos.X >= 0 && pos.Y >= 0 {
					window.CurrentBuffer.CursorPos = pos
					window.PrintMessage("Match replaced successfully", TYPER_MESSAGE_INFO)
				} else {
					window.PrintMessage(fmt.Sprintf("'%s' not found in buffer", string(findStr)), TYPER_MESSAGE_WARNING)
				}

				return
			}

			go func() {
				inputChannel := RequestInput(window, "Substring to search for:", "")
				findStr := runestring.RuneString(<-inputChannel)
				if len(findStr) == 0 {
					return
				}

				inputChannel = RequestInput(window, "String to replace with:", "")
				replaceStr := runestring.RuneString(<-inputChannel)

				pos := window.CurrentBuffer.FindAndReplaceSubstring(findStr, replaceStr, window.CurrentBuffer.CursorPos)
				if pos.X >= 0 && pos.Y >= 0 {
					window.CurrentBuffer.CursorPos = pos
					window.PrintMessage("Match replaced successfully", TYPER_MESSAGE_INFO)
				} else {
					window.PrintMessage(fmt.Sprintf("'%s' not found in buffer", string(findStr)), TYPER_MESSAGE_WARNING)
				}
			}()
		},
	}

	replaceAllCmd := Command{
		cmd: "replace-all",
		run: func(window *Window, args ...string) {
			if len(args) >= 2 {
				findStr := runestring.RuneString(args[0])
				replaceStr := runestring.RuneString(args[1])

				if len(findStr) == 0 {
					return
				}

				replacements := window.CurrentBuffer.FindAndReplaceAll(findStr, replaceStr)
				if replacements > 0 {
					window.PrintMessage(fmt.Sprintf("Replaced all %d matches successfully", replacements), TYPER_MESSAGE_INFO)
				} else {
					window.PrintMessage(fmt.Sprintf("'%s' not found in buffer", string(findStr)), TYPER_MESSAGE_WARNING)
				}

				return
			}

			go func() {
				inputChannel := RequestInput(window, "Substring to search for:", "")
				findStr := runestring.RuneString(<-inputChannel)
				if len(findStr) == 0 {
					return
				}

				inputChannel = RequestInput(window, "String to replace with:", "")
				replaceStr := runestring.RuneString(<-inputChannel)

				replacements := window.CurrentBuffer.FindAndReplaceAll(findStr, replaceStr)
				if replacements > 0 {
					window.PrintMessage(fmt.Sprintf("Replaced all %d matches successfully", replacements), TYPER_MESSAGE_INFO)
				} else {
					window.PrintMessage(fmt.Sprintf("'%s' not found in buffer", string(findStr)), TYPER_MESSAGE_WARNING)
				}
			}()
		},
	}

	prevBufferCmd := Command{
		cmd: "prev-buffer",
		run: func(window *Window, args ...string) {
			if window.CursorMode != CursorModeBuffer {
				return
			}

			index := slices.Index(Buffers, window.CurrentBuffer)

			index--
			if index < 0 {
				index = 0
			}

			window.CurrentBuffer = Buffers[index]
			window.PrintMessage(fmt.Sprintf("Set current buffer to '%s'", window.CurrentBuffer.Name), TYPER_MESSAGE_INFO)
		},
	}

	nextBufferCmd := Command{
		cmd: "next-buffer",
		run: func(window *Window, args ...string) {
			if window.CursorMode != CursorModeBuffer {
				return
			}

			index := slices.Index(Buffers, window.CurrentBuffer)

			index++
			if index >= len(Buffers) {
				index = len(Buffers) - 1
			}

			window.CurrentBuffer = Buffers[index]
			window.PrintMessage(fmt.Sprintf("Set current buffer to '%s'", window.CurrentBuffer.Name), TYPER_MESSAGE_INFO)
		},
	}

	newBufferCmd := Command{
		cmd: "new-buffer",
		run: func(window *Window, args ...string) {
			for i := 1; true; i++ {
				buffer, err := CreateBuffer("New Buffer " + strconv.Itoa(i))
				if err == nil {
					window.CurrentBuffer = buffer
					break
				}
			}

			window.CursorMode = CursorModeBuffer
			window.PrintMessage(fmt.Sprintf("New buffer created with the name '%s'", window.CurrentBuffer.Name), TYPER_MESSAGE_INFO)
		},
	}

	closeBufferCmd := Command{
		cmd: "close-buffer",
		run: func(window *Window, args ...string) {
			bufferIndex := slices.Index(Buffers, window.CurrentBuffer)
			Buffers = DeleteFromSlice(Buffers, bufferIndex)
			if len(Buffers) == 0 {
				window.Close()
				return
			}
			if bufferIndex >= len(Buffers) {
				window.CurrentBuffer = Buffers[bufferIndex-1]
			} else {
				window.CurrentBuffer = Buffers[bufferIndex]
			}
			window.CursorMode = CursorModeBuffer
			window.PrintMessage("Buffer closed", TYPER_MESSAGE_INFO)
		},
	}

	toggleTopBar := Command{
		cmd: "toggle-top-bar",
		run: func(window *Window, args ...string) {
			window.ShowTopMenu = !window.ShowTopMenu
		},
	}

	toggleLineIndex := Command{
		cmd: "toggle-line-index",
		run: func(window *Window, args ...string) {
			window.ShowLineIndex = !window.ShowLineIndex
		},
	}

	setStyleCmd := Command{
		cmd: "set-style",
		run: func(window *Window, args ...string) {
			if len(args) >= 1 {
				input := args[0]

				if input == "" {
					return
				}

				if _, ok := AvailableStyles[input]; !ok {
					window.PrintMessage(fmt.Sprintf("Could not set style to '%s'", input), TYPER_MESSAGE_ERROR)
					return
				}

				if ok := SetCurrentStyle(window.screen, input); ok {
					window.PrintMessage(fmt.Sprintf("Setting style to '%s'", input), TYPER_MESSAGE_INFO)
				} else {
					window.PrintMessage(fmt.Sprintf("Could not set style to '%s'", input), TYPER_MESSAGE_ERROR)
				}

				return
			}

			inputChannel := RequestInput(window, "Style to switch to:", "")
			go func() {
				input := <-inputChannel

				if input == "" {
					return
				}

				if _, ok := AvailableStyles[input]; !ok {
					window.PrintMessage(fmt.Sprintf("Could not set style to '%s'", input), TYPER_MESSAGE_ERROR)
					return
				}

				if ok := SetCurrentStyle(window.screen, input); ok {
					window.PrintMessage(fmt.Sprintf("Setting style to '%s'", input), TYPER_MESSAGE_INFO)
				} else {
					window.PrintMessage(fmt.Sprintf("Could not set style to '%s'", input), TYPER_MESSAGE_ERROR)
				}
			}()
		},
	}

	setFiletypeCmd := Command{
		cmd: "set-filetype",
		run: func(window *Window, args ...string) {
			if len(args) >= 1 {
				input := args[0]

				if input == "" {
					return
				}

				if strings.ToLower(input) == "none" {
					window.CurrentBuffer.filetype = ""
					window.PrintMessage("Setting filetype to 'none'", TYPER_MESSAGE_INFO)
					return
				} else if _, ok := AvailableSyntaxes[input]; !ok {
					window.PrintMessage(fmt.Sprintf("Could not set filetype to '%s'", input), TYPER_MESSAGE_ERROR)
					return
				}

				window.CurrentBuffer.filetype = input
				window.PrintMessage(fmt.Sprintf("Setting filetype to '%s'", input), TYPER_MESSAGE_INFO)

				return
			}

			inputChannel := RequestInput(window, "Filetype to switch to:", "")
			go func() {
				input := <-inputChannel

				if input == "" {
					return
				}

				if _, ok := AvailableSyntaxes[input]; !ok {
					window.PrintMessage(fmt.Sprintf("Could not set filetype to '%s'", input), TYPER_MESSAGE_ERROR)
					return
				}

				window.CurrentBuffer.filetype = input
				window.PrintMessage(fmt.Sprintf("Setting filetype to '%s'", input), TYPER_MESSAGE_INFO)

			}()
		},
	}

	menuFileCmd := Command{
		cmd: "menu-file",
		run: func(window *Window, args ...string) {
			for _, button := range TopMenuButtons {
				if button.Name == "File" {
					button.Action(window, &button)
					break
				}
			}
		},
	}

	menuEditCmd := Command{
		cmd: "menu-edit",
		run: func(window *Window, args ...string) {
			for _, button := range TopMenuButtons {
				if button.Name == "Edit" {
					button.Action(window, &button)
					break
				}
			}
		},
	}

	menuBuffersCmd := Command{
		cmd: "menu-buffers",
		run: func(window *Window, args ...string) {
			for _, button := range TopMenuButtons {
				if button.Name == "Buffers" {
					button.Action(window, &button)
					break
				}
			}
		},
	}

	quitCmd := Command{
		cmd: "quit",
		run: func(window *Window, args ...string) {
			window.Close()
			window.CursorMode = CursorModeBuffer
		},
	}

	executeCmd := Command{
		cmd: "execute",
		run: func(window *Window, args ...string) {
			inputChannel := RequestInput(window, "Run:", "")

			go func() {
				input := strings.TrimSpace(<-inputChannel)

				if input == "" {
					return
				}

				var arguments []string

				builder := &strings.Builder{}
				quoted := false
				for _, r := range input {
					if r == '"' {
						quoted = !quoted
					} else if !quoted && r == ' ' {
						arguments = append(arguments, builder.String())
						builder.Reset()
					} else {
						builder.WriteRune(r)
					}
				}
				if builder.Len() > 0 {
					arguments = append(arguments, builder.String())
				}

				window.CursorMode = CursorModeBuffer

				if len(arguments) == 1 {
					RunCommand(window, arguments[0])
				} else {
					RunCommand(window, arguments[0], arguments[1:]...)
				}
			}()
		},
	}

	// Register commands
	commands["select-all"] = &selectAll
	commands["cut"] = &cutCmd
	commands["copy"] = &copyCmd
	commands["paste"] = &pasteCmd
	commands["save"] = &saveCmd
	commands["open"] = &openCmd
	commands["reload"] = &reloadCmd
	commands["find"] = &findCmd
	commands["replace"] = &replaceCmd
	commands["replace-all"] = &replaceAllCmd
	commands["prev-buffer"] = &prevBufferCmd
	commands["next-buffer"] = &nextBufferCmd
	commands["new-buffer"] = &newBufferCmd
	commands["close-buffer"] = &closeBufferCmd
	commands["toggle-top-bar"] = &toggleTopBar
	commands["toggle-line-index"] = &toggleLineIndex
	commands["set-style"] = &setStyleCmd
	commands["set-filetype"] = &setFiletypeCmd
	commands["menu-file"] = &menuFileCmd
	commands["menu-edit"] = &menuEditCmd
	commands["menu-buffers"] = &menuBuffersCmd
	commands["quit"] = &quitCmd
	commands["execute"] = &executeCmd
}

func RunCommand(window *Window, cmd string, args ...string) bool {
	if command, ok := commands[cmd]; ok {
		command.run(window, args...)
		return true
	} else {
		window.PrintMessage(fmt.Sprintf("Could not find command '%s'", cmd), TYPER_MESSAGE_ERROR)
		return false
	}
}
