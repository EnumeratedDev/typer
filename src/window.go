package main

import (
	"log"
	"slices"
	"strconv"
	"strings"
	"time"
	"typer/runestring"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type CursorMode uint8

const (
	CursorModeDisabled CursorMode = iota
	CursorModeBuffer
	CursorModeDropdown
	CursorModeInputBar
)

var CursorModeNames = map[CursorMode]string{
	CursorModeDisabled: "disabled",
	CursorModeBuffer:   "buffer",
	CursorModeDropdown: "dropdown",
	CursorModeInputBar: "input_bar",
}

type Window struct {
	ShowTopMenu   bool
	ShowLineIndex bool
	CursorMode    CursorMode

	Clipboard runestring.RuneString

	CurrentBuffer *Buffer

	screen tcell.Screen

	closed bool
}

var mouseHeld = false
var lastClick int64 = 0

func CreateWindow() (*Window, error) {
	window := Window{
		ShowTopMenu:   Config.ShowTopMenu,
		ShowLineIndex: Config.ShowLineIndex,
		CursorMode:    CursorModeBuffer,

		CurrentBuffer: nil,

		screen: nil,
	}

	// Create empty buffer if nil
	for i := 1; window.CurrentBuffer == nil; i++ {
		buffer, err := CreateBuffer("New Buffer " + strconv.Itoa(i))
		if err == nil {
			window.CurrentBuffer = buffer
		}
	}

	// Create tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to initialize tcell: %s", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to initialize screen: %s", err)
	}

	// Enable mouse
	screen.EnableMouse()

	// Set window screen field
	window.screen = screen

	// Try to set screen style to selected one
	if ok := SetCurrentStyle(screen, Config.SelectedStyle); !ok {
		// Try to set screen style to selected fallback one
		if ok := SetCurrentStyle(screen, Config.FallbackStyle); !ok {
			// Use hard-coded fallback style
			screen.SetStyle(tcell.StyleDefault.Foreground(CurrentStyle.BufferAreaFg).Background(CurrentStyle.BufferAreaBg))
			PrintMessage(&window, "Could not set style either to selected one nor to fallback one!")
		}
	}

	// Initialize top menu
	initTopMenu()

	return &window, nil
}

func (window *Window) Draw() {
	// Clear screen
	window.screen.Clear()

	// Sync buffer offset
	window.SyncBufferOffset()

	// Draw top menu
	if window.ShowTopMenu {
		drawTopMenu(window)
	}

	// Draw line index
	if window.ShowLineIndex {
		drawLineIndex(window)
	}

	// Draw current buffer
	if window.CurrentBuffer != nil {
		drawBuffer(window)
	}

	// Draw input bar
	if currentInputRequest != nil {
		drawInputBar(window)
	}

	// Draw message bar
	drawMessageBar(window)

	// Draw dropdowns
	drawDropdowns(window)

	// Draw cursor
	if window.CursorMode == CursorModeInputBar {
		_, sizeY := window.screen.Size()
		window.screen.ShowCursor(len(currentInputRequest.Text)+len(currentInputRequest.input)+1, sizeY-1)
	} else {
		window.screen.HideCursor()
	}

	// Update screen
	window.screen.Show()
}

func (window *Window) ProcessEvents() {
	// Poll event
	ev := window.screen.PollEvent()

	// Process event
	switch ev := ev.(type) {
	case *tcell.EventResize:
		window.screen.Sync()
		window.SyncBufferOffset()
	case *tcell.EventMouse:
		window.handleMouseInput(ev)
	case *tcell.EventKey:
		window.handleKeyInput(ev)
	}
}

func (window *Window) handleKeyInput(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyRight { // Navigation Keys
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos

			if ev.Modifiers()&tcell.ModCtrl != 0 {
				// Move cursor to start of word
				// Set variable to one character right of current position
				window.CurrentBuffer.MoveRight(1)

				// Skip all spaces
				for unicode.IsSpace(window.CurrentBuffer.GetCharAtPosition(window.CurrentBuffer.CursorPos)) {
					if !window.CurrentBuffer.MoveRight(1) {
						break
					}
				}

				// Find end of word
				for !unicode.IsSpace(window.CurrentBuffer.GetCharAtPosition(window.CurrentBuffer.CursorPos)) {
					if !window.CurrentBuffer.MoveRight(1) {
						break
					}
				}
			} else {
				// Move cursor one character forwards
				window.CurrentBuffer.MoveRight(1)
			}

			// Add to selection
			if ev.Modifiers()&tcell.ModShift != 0 {
				if window.CurrentBuffer.Selection == nil {
					// Cancel cursor movement when creating selection without holding ctrl
					if ev.Modifiers()&tcell.ModCtrl == 0 {
						window.CurrentBuffer.CursorPos = pos
					}

					window.CurrentBuffer.Selection = &Selection{
						selectionStart: pos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
					}
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos
				}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
			}
		}
	} else if ev.Key() == tcell.KeyLeft {
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos

			if ev.Modifiers()&tcell.ModCtrl != 0 {
				// Move cursor to start of word
				// Set variable to one character left of current position
				window.CurrentBuffer.MoveLeft(1)

				// Skip all spaces
				for unicode.IsSpace(window.CurrentBuffer.GetCharAtPosition(window.CurrentBuffer.CursorPos)) {
					if !window.CurrentBuffer.MoveLeft(1) {
						break
					}
				}

				// Find end of word
				for {
					char := window.CurrentBuffer.GetCharAtPosition(window.CurrentBuffer.CursorPos)
					if char == 0 || unicode.IsSpace(char) {
						break
					}
					if !window.CurrentBuffer.MoveLeft(1) {
						break
					}
				}

				// Move one character to the right if not selecting
				if window.CurrentBuffer.CursorPos.X != 0 {
					window.CurrentBuffer.MoveRight(1)
				}
			} else {
				// Move cursor one character backwards
				window.CurrentBuffer.MoveLeft(1)
			}

			// Add to selection
			if ev.Modifiers()&tcell.ModShift != 0 {
				if window.CurrentBuffer.Selection == nil {
					// Cancel cursor movement when creating selection without holding ctrl
					if ev.Modifiers()&tcell.ModCtrl == 0 {
						window.CurrentBuffer.CursorPos = pos
					}

					window.CurrentBuffer.Selection = &Selection{
						selectionStart: pos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
					}
					return
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos
				}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
				return
			}
		}
	} else if ev.Key() == tcell.KeyUp {
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos

			if ev.Modifiers()&tcell.ModCtrl != 0 {
				// Move cursor to top of buffer
				window.CurrentBuffer.CursorPos.X = 0
				window.CurrentBuffer.CursorPos.Y = 0
			} else {
				// Move cursor one line up
				window.CurrentBuffer.MoveUp(1)
			}

			// Add to selection
			if ev.Modifiers()&tcell.ModShift != 0 {
				// Add to selection
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: pos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
					}
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos
				}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
				return
			}
		} else if window.CursorMode == CursorModeDropdown {
			dropdown := ActiveDropdown
			dropdown.Selected--
			if dropdown.Selected < 0 {
				dropdown.Selected = 0
			}
		} else if window.CursorMode == CursorModeInputBar {
			if len(inputHistory) == 0 {
				return
			}

			current := slices.Index(inputHistory, currentInputRequest.input)
			if current < 0 {
				current = len(inputHistory) - 1
			} else if current != 0 {
				current--
			}

			currentInputRequest.input = inputHistory[current]
			currentInputRequest.cursorPos = len(inputHistory[current])
		}
	} else if ev.Key() == tcell.KeyDown {
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos

			if ev.Modifiers()&tcell.ModCtrl != 0 {
				// Move cursor to bottom of buffer
				window.CurrentBuffer.CursorPos.Y = len(window.CurrentBuffer.Contents) - 1
				window.CurrentBuffer.CursorPos.X = len(window.CurrentBuffer.Contents[window.CurrentBuffer.CursorPos.Y])
			} else {
				// Move cursor one line down
				window.CurrentBuffer.MoveDown(1)
			}

			// Add to selection
			if ev.Modifiers()&tcell.ModShift != 0 {
				// Add to selection
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: pos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
					}
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos
				}
				// Prevent selecting dummy character at the end of the buffer
				//if window.CurrentBuffer.Selection.selectionEnd >= len(window.CurrentBuffer.Contents) {
				//	window.CurrentBuffer.Selection.selectionEnd = len(window.CurrentBuffer.Contents) - 1
				//}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
				return
			}
		} else if window.CursorMode == CursorModeDropdown {
			dropdown := ActiveDropdown
			dropdown.Selected++
			if dropdown.Selected >= len(dropdown.Options) {
				dropdown.Selected = len(dropdown.Options) - 1
			}
		} else if window.CursorMode == CursorModeInputBar {
			if len(inputHistory) == 0 {
				return
			}

			current := slices.Index(inputHistory, currentInputRequest.input)
			if current < 0 {
				return
			} else if current == len(inputHistory)-1 {
				currentInputRequest.input = ""
				return
			} else {
				current++
			}

			currentInputRequest.input = inputHistory[current]
			currentInputRequest.cursorPos = len(inputHistory[current])
		}
	} else if ev.Key() == tcell.KeyEscape {
		if window.CursorMode == CursorModeInputBar {
			currentInputRequest.inputChannel <- ""
			currentInputRequest = nil
			window.CursorMode = CursorModeBuffer
		} else {
			ClearDropdowns()
			window.CursorMode = CursorModeBuffer
		}
	}

	// Check key bindings
	for _, keybinding := range Keybindings.Keybindings {
		if keybinding.IsPressed(ev) && slices.Index(keybinding.GetCursorModes(), window.CursorMode) != -1 {
			RunCommand(window, keybinding.Command)
			return
		}
	}

	// Typing
	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		if window.CursorMode == CursorModeBuffer {
			if window.CurrentBuffer.Selection != nil {
				window.CurrentBuffer.CutText(window)
			} else {
				window.CurrentBuffer.Delete(1)
			}
		} else if window.CursorMode == CursorModeInputBar {
			str := currentInputRequest.input
			index := currentInputRequest.cursorPos

			if index != 0 {
				str = str[:index-1] + str[index:]
				currentInputRequest.cursorPos--
				currentInputRequest.input = str
			}
		}
	} else if ev.Key() == tcell.KeyTab {
		if window.CursorMode == CursorModeBuffer {
			// Remove selected text
			if window.CurrentBuffer.Selection != nil {
				window.CurrentBuffer.CutText(window)
			}

			window.CurrentBuffer.WriteRune('\t')
		}
	} else if ev.Key() == tcell.KeyEnter {
		if window.CursorMode == CursorModeBuffer {
			// Remove selected text
			if window.CurrentBuffer.Selection != nil {
				window.CurrentBuffer.CutText(window)
			}

			window.CurrentBuffer.WriteRune('\n')
		} else if window.CursorMode == CursorModeInputBar {
			if currentInputRequest.input == "" && slices.Index(inputHistory, currentInputRequest.input) == -1 {
				inputHistory = append(inputHistory, currentInputRequest.input)
			}
			currentInputRequest.inputChannel <- currentInputRequest.input
			currentInputRequest = nil
			window.CursorMode = CursorModeBuffer
		} else if window.CursorMode == CursorModeDropdown {
			d := ActiveDropdown
			d.Action(d.Selected)
		}
	} else if ev.Key() == tcell.KeyRune {
		if window.CursorMode == CursorModeBuffer {
			// Remove selected text
			if window.CurrentBuffer.Selection != nil {
				window.CurrentBuffer.CutText(window)
			}

			window.CurrentBuffer.WriteRune(ev.Rune())
		} else if window.CursorMode == CursorModeInputBar {
			str := currentInputRequest.input
			index := currentInputRequest.cursorPos

			if index == len(str) {
				str += string(ev.Rune())
			} else {
				str = str[:index] + string(ev.Rune()) + str[index:]
			}

			currentInputRequest.cursorPos++
			currentInputRequest.input = str
		}
	}
}

func (window *Window) handleMouseInput(ev *tcell.EventMouse) {
	mouseX, mouseY := ev.Position()

	// Left click was pressed
	if ev.Buttons() == tcell.Button1 {
		// Get last click time
		lastClickTime := time.UnixMilli(lastClick)
		// Ensure click was in buffer area
		x1, y1, x2, y2 := window.GetTextAreaDimensions()
		if mouseX >= x1 && mouseY >= y1 && mouseX <= x2 && mouseY <= y2 {
			currentPos := window.CurrentBuffer.CursorPos
			mouseBufferPos := Position{mouseX + window.CurrentBuffer.Offset.X - x1, mouseY + window.CurrentBuffer.Offset.Y - y1}

			// Keep mouse Y in bounds
			if mouseBufferPos.Y >= len(window.CurrentBuffer.Contents) {
				mouseBufferPos.Y = len(window.CurrentBuffer.Contents) - 1
			}

			// Offset mouse X for each tab character in line
			posInLine := make([]int, 0)
			for i, r := range append(window.CurrentBuffer.Contents[mouseBufferPos.Y], ' ') {
				if r == '\t' {
					for j := 0; j < Config.TabIndentation; j++ {
						posInLine = append(posInLine, i)
					}
				} else {
					posInLine = append(posInLine, i)
				}
			}
			if len(posInLine) == 0 {
				mouseBufferPos.X = 0
			} else if mouseBufferPos.X >= len(posInLine) {
				mouseBufferPos.X = posInLine[len(posInLine)-1]
			} else {
				mouseBufferPos.X = posInLine[mouseBufferPos.X]
			}

			// Keep mouse X in bounds
			if mouseBufferPos.X > len(window.CurrentBuffer.Contents[mouseBufferPos.Y]) {
				mouseBufferPos.X = len(window.CurrentBuffer.Contents[mouseBufferPos.Y])
			}

			if mouseHeld {
				// Add to selection
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: window.CurrentBuffer.CursorPos,
						selectionEnd:   mouseBufferPos,
					}

					// Set last click time
					lastClick = time.Now().UnixMilli()

					return
				} else {
					window.CurrentBuffer.Selection.selectionEnd = mouseBufferPos
				}
			} else if currentPos == mouseBufferPos && time.Since(lastClickTime).Milliseconds() < 300 {
				selectedText := window.CurrentBuffer.GetSelectedText()
				if window.CurrentBuffer.Selection == nil || strings.HasSuffix(string(selectedText), "\n") {
					// Select word
					cursorPos := window.CurrentBuffer.CursorPos
					startOfWord := window.CurrentBuffer.CursorPos.X
					endOfWord := window.CurrentBuffer.CursorPos.X

					// Find end of word
					for i := cursorPos.X + 1; i < len(window.CurrentBuffer.Contents[cursorPos.Y]); i++ {
						currentRune := rune(window.CurrentBuffer.Contents[cursorPos.Y][i])
						if unicode.IsLetter(currentRune) || unicode.IsDigit(currentRune) || currentRune == '_' {
							endOfWord++
						} else {
							break
						}
					}

					// Find start of word
					for i := cursorPos.X - 1; i >= 0; i-- {
						currentRune := rune(window.CurrentBuffer.Contents[cursorPos.Y][i])
						if unicode.IsLetter(currentRune) || unicode.IsDigit(currentRune) || currentRune == '_' {
							startOfWord--
						} else {
							break
						}
					}

					// Add to selection
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: Position{startOfWord, cursorPos.Y},
						selectionEnd:   Position{endOfWord, cursorPos.Y},
					}
				} else {
					// Select line
					cursorPos := window.CurrentBuffer.CursorPos

					// Add to selection
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: Position{0, cursorPos.Y},
						selectionEnd:   Position{len(window.CurrentBuffer.Contents[cursorPos.Y]), cursorPos.Y},
					}
				}

				// Set last click time
				lastClick = time.Now().UnixMilli()

				return
			} else {
				// Clear selection
				if window.CurrentBuffer.Selection != nil {
					window.CurrentBuffer.Selection = nil
				}
			}
			// Move cursor
			window.CurrentBuffer.CursorPos = mouseBufferPos

			// Set last click time
			lastClick = time.Now().UnixMilli()
		}
		mouseHeld = true
	} else if ev.Buttons() == tcell.ButtonNone {
		if mouseHeld {
			mouseHeld = false
		}
	}
}

func (window *Window) Close() {
	window.closed = true
	err := window.screen.PostEvent(tcell.NewEventInterrupt(nil))
	if err != nil {
		return
	}
}

func (window *Window) GetTextAreaDimensions() (int, int, int, int) {
	x1, y1 := 0, 0
	x2, y2 := window.screen.Size()

	if window.ShowTopMenu {
		y1++
	}

	if window.ShowLineIndex {
		x1 += getLineIndexSize(window)
	}

	return x1, y1, x2 - 1, y2 - 2
}

func (window *Window) SyncBufferOffset() {
	cursorPos := window.CurrentBuffer.CursorPos
	bufferX1, bufferY1, bufferX2, bufferY2 := window.GetTextAreaDimensions()

	if cursorPos.Y < window.CurrentBuffer.Offset.Y {
		window.CurrentBuffer.Offset.Y = cursorPos.Y
	} else if cursorPos.Y > window.CurrentBuffer.Offset.Y+(bufferY2-bufferY1) {
		window.CurrentBuffer.Offset.Y = cursorPos.Y - (bufferY2 - bufferY1)
	}

	if cursorPos.X < window.CurrentBuffer.Offset.X {
		window.CurrentBuffer.Offset.X = cursorPos.X
	} else if cursorPos.X > window.CurrentBuffer.Offset.X+(bufferX2-bufferX1) {
		window.CurrentBuffer.Offset.X = cursorPos.X - (bufferX2 - bufferX1)
	}
}
