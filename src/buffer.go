package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"typer/runestring"

	"github.com/gdamore/tcell/v2"
)

type Buffer struct {
	Name     string
	Contents []runestring.RuneString

	CursorPos Position
	Offset    Position

	Selection *Selection

	canSave  bool
	canEdit  bool
	filetype string
	filename string
}

type Selection struct {
	selectionStart, selectionEnd Position
}

var Buffers = make([]*Buffer, 0)

func GetBufferByName(name string) *Buffer {
	for _, buffer := range Buffers {
		if buffer.Name == name {
			return buffer
		}
	}
	return nil
}

func GetBufferByFilename(filename string) *Buffer {
	for _, buffer := range Buffers {
		if buffer.filename == filename {
			return buffer
		}
	}
	return nil
}

func drawBuffer(window *Window) {
	buffer := window.CurrentBuffer

	x, y, _, _ := window.GetTextAreaDimensions()

	bufferX, bufferY, _, _ := window.GetTextAreaDimensions()

	parsedSyntaxes, err := HighlightString(string(buffer.GetContentsAsString()), buffer.filetype)
	if err != nil {
		window.PrintMessage(fmt.Sprintf("Could not parse regular expression in '%s' syntax: %s", buffer.filetype, err), TYPER_MESSAGE_ERROR)
	}

	i := -1
	for lineIndex, line := range buffer.Contents {
		for runeIndex, r := range append(line, ' ') {
			i++
			drawPosition := Position{runeIndex, lineIndex}

			if x-buffer.Offset.X >= bufferX && y-buffer.Offset.Y >= bufferY {
				// Default style
				style := tcell.StyleDefault.Background(CurrentStyle.BufferAreaBg).Foreground(CurrentStyle.BufferAreaFg)

				// Check for syntax highlighting
				for _, parsedSyntax := range parsedSyntaxes {
					if i >= parsedSyntax.StartIndex && i < parsedSyntax.EndIndex {
						switch parsedSyntax.Type {
						case "comment":
							style = style.Foreground(CurrentStyle.SyntaxComment)
						case "keyword":
							style = style.Foreground(CurrentStyle.SyntaxKeyword)
						case "identifier":
							style = style.Foreground(CurrentStyle.SyntaxIdentifier)
						case "constant":
							style = style.Foreground(CurrentStyle.SyntaxConstant)
						case "variable":
							style = style.Foreground(CurrentStyle.SyntaxVariable)
						case "string":
							style = style.Foreground(CurrentStyle.SyntaxString)

						// Special types for typer logs
						case "info":
							style = style.Foreground(CurrentStyle.SyntaxInfo)
						case "warning":
							style = style.Foreground(CurrentStyle.SyntaxWarning)
						case "error":
							style = style.Foreground(CurrentStyle.SyntaxError)
						}

						break
					}
				}

				// Change background if under cursor
				if buffer.CursorPos.Equals(runeIndex, lineIndex) {
					style = style.Background(CurrentStyle.BufferAreaSel)
				}

				// Change background if selected
				if buffer.Selection != nil {
					edge1, edge2 := buffer.GetSelectionEdges()

					if ComparePositions(drawPosition, edge1) >= 0 && ComparePositions(drawPosition, edge2) <= 0 {
						style = style.Background(CurrentStyle.BufferAreaSel)

						// Show selection on entire tab space
						if r == '\t' {
							for j := 0; j < int(Config.TabIndentation); j++ {
								window.screen.SetContent(x+j-buffer.Offset.X, y-buffer.Offset.Y, r, nil, style)
							}
						}
					}
				}

				window.screen.SetContent(x-buffer.Offset.X, y-buffer.Offset.Y, r, nil, style)
			}

			// Change position for next character
			if r == '\t' {
				x += int(Config.TabIndentation)
			} else {
				x++
			}
		}
		// Draw new line
		x = bufferX
		y++
	}

}

func (buffer *Buffer) Load() error {
	// Do not load if canSave is false or filename is not set
	if !buffer.canSave || buffer.filename == "" {
		return nil
	}

	// Replace tilde with home directory
	if strings.HasPrefix(buffer.filename, "~/") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		buffer.filename = filepath.Join(homedir, buffer.filename[2:])
	}

	contentBytes, err := os.ReadFile(buffer.filename)
	if err != nil {
		return err
	}
	content := runestring.RuneString(string(contentBytes))

	if len(content) != 0 {
		buffer.Contents = runestring.Split(content, '\n')

		// Add empty line at end of buffer for last newline
		if content[len(content)-1] == '\n' {
			buffer.Contents = append(buffer.Contents, make(runestring.RuneString, 0))
		}

		buffer.CursorPos.Y = len(buffer.Contents) - 1
		buffer.CursorPos.X = len(buffer.Contents[buffer.CursorPos.Y])
	}

	// Set buffer filetype
	for _, syntax := range AvailableSyntaxes {
		if syntax.Filenames == "" {
			continue
		}

		if ok, _ := regexp.MatchString(syntax.Filenames, buffer.filename); ok {
			buffer.filetype = syntax.Filetype
		}
	}

	return nil
}

func (buffer *Buffer) Save() error {
	// Do not save if canSave is false or filename is not set
	if !buffer.canSave || buffer.filename == "" {
		return nil
	}

	// Replace tilde with home directory
	if strings.HasPrefix(buffer.filename, "~/") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		buffer.filename = filepath.Join(homedir, buffer.filename[2:])
	}

	// Add newline at the end of buffer Contents
	line := buffer.Contents[len(buffer.Contents)-1]
	if len(line) != 0 {
		buffer.Contents = append(buffer.Contents, make(runestring.RuneString, 0))
	}

	err := os.WriteFile(buffer.filename, []byte(string(buffer.GetContentsAsString())), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (buffer *Buffer) GetContentsAsString() runestring.RuneString {
	finalText := make(runestring.RuneString, 0)
	for i, line := range buffer.Contents {
		finalText = append(finalText, line...)

		if i != len(buffer.Contents)-1 {
			finalText = append(finalText, '\n')
		}
	}

	return finalText
}

func (buffer *Buffer) PositionToAbsolutePosition(position Position) int {
	i := 0
	for lineIndex, line := range buffer.Contents {
		if len(line) == 0 {
			line = append(line, ' ')
		}
		for runeIndex, _ := range append(line, ' ') {
			if position.Equals(runeIndex, lineIndex) {
				return i
			}
			i++
		}
	}

	return i
}

func (buffer *Buffer) AbsolutePositionToPosition(absolutePosition int) Position {
	i := 0

	for lineIndex, line := range buffer.Contents {
		for runeIndex, _ := range append(line, ' ') {
			if i == absolutePosition {
				return Position{runeIndex, lineIndex}
			}
			i++
		}
	}

	lastLine := buffer.Contents[len(buffer.Contents)-1]
	return Position{len(buffer.Contents) - 1, len(lastLine) - 1}
}

func (buffer *Buffer) GetSelectionEdges() (Position, Position) {
	if buffer.Selection == nil {
		return Position{-1, -1}, Position{-1, -1}
	}

	if ComparePositions(buffer.Selection.selectionStart, buffer.Selection.selectionEnd) == -1 {
		return buffer.Selection.selectionStart, buffer.Selection.selectionEnd
	} else {
		return buffer.Selection.selectionEnd, buffer.Selection.selectionStart
	}
}

func (buffer *Buffer) GetSelectedText() runestring.RuneString {
	if buffer.Selection == nil {
		return make(runestring.RuneString, 0)
	}
	if len(buffer.Contents) == 0 {
		return make(runestring.RuneString, 0)
	}

	edge1, edge2 := buffer.GetSelectionEdges()

	selectedText := make(runestring.RuneString, 0)
	if r := buffer.GetCharAtPosition(edge1); r != 0 {
		selectedText = append(selectedText, r)
	}

	for ComparePositions(edge1, edge2) < 0 {
		edge1.X++
		if edge1.X > len(buffer.Contents[edge1.Y]) {
			if edge1.Y < len(buffer.Contents) {
				edge1.Y++
				edge1.X = 0
			} else {
				edge1.X = len(buffer.Contents[edge1.Y])
				break
			}
		}

		if r := buffer.GetCharAtPosition(edge1); r != 0 {
			selectedText = append(selectedText, buffer.GetCharAtPosition(edge1))
		}
	}

	return selectedText
}

func (buffer *Buffer) CutText(window *Window) (runestring.RuneString, int) {
	if buffer.Selection == nil {
		// Cut current line
		cutText := append(buffer.Contents[buffer.CursorPos.Y], '\n')

		// Remove line from buffer contents
		if len(buffer.Contents) == 1 {
			buffer.Contents[0] = make(runestring.RuneString, 0)
		} else {
			buffer.Contents = slices.Delete(buffer.Contents, buffer.CursorPos.Y, buffer.CursorPos.Y+1)
		}

		buffer.CursorPos.Y -= 1
		if buffer.CursorPos.Y < 0 {
			buffer.CursorPos = Position{0, 0}
		}

		return cutText, 0
	} else {
		// Cut selection
		cutText := buffer.GetSelectedText()

		// Remove selected text
		_, edge2 := buffer.GetSelectionEdges()
		buffer.CursorPos = edge2
		buffer.MoveRight(1)

		buffer.Delete(len(cutText))

		// Remove selection
		buffer.Selection = nil

		return cutText, 1
	}
}

func (buffer *Buffer) CopyText() (runestring.RuneString, int) {
	if buffer.Selection == nil {
		// Cut current line
		copiedText := append(buffer.Contents[buffer.CursorPos.Y], '\n')

		return copiedText, 0
	} else {
		// Copy selection
		return buffer.GetSelectedText(), 1
	}
}

func (buffer *Buffer) PasteText(window *Window, text runestring.RuneString) {
	// Remove selected text
	if buffer.Selection != nil {
		_, edge2 := buffer.GetSelectionEdges()

		buffer.CursorPos = edge2
		buffer.Delete(len(buffer.GetSelectedText()))

		buffer.Selection = nil
	}

	buffer.WriteString(text)
}

func (buffer *Buffer) FindSubstring(substring runestring.RuneString, afterPos Position) Position {
	// Return no match if afterPos is larger than the buffer contents size
	contents := buffer.GetContentsAsString()
	absAfterPos := buffer.PositionToAbsolutePosition(afterPos)

	if absAfterPos >= len(contents) {
		return Position{-1, -1}
	}

	index := runestring.Index(contents[absAfterPos+1:], substring)

	if index != -1 {
		index += absAfterPos + 1
	}
	return buffer.AbsolutePositionToPosition(index)
}

func (buffer *Buffer) FindAndReplaceSubstring(substring, replacement runestring.RuneString, afterPos Position) Position {
	// Return no match if afterPos is larger than the buffer contents size
	contents := buffer.GetContentsAsString()
	absAfterPos := buffer.PositionToAbsolutePosition(afterPos)

	if absAfterPos >= len(contents) {
		return Position{-1, -1}
	}

	index := runestring.Index(contents[absAfterPos+1:], substring)

	if index != -1 {
		index += absAfterPos + 1
	}

	// Replace substring with replacement string
	contents = slices.Insert(contents, index, replacement...)

	buffer.Contents = runestring.Split(contents, '\n')

	return buffer.AbsolutePositionToPosition(index)
}

func (buffer *Buffer) FindAndReplaceAll(substring, replacement runestring.RuneString) int {
	replacements := 0
	position := Position{}
	for position.X != -1 && position.Y != -1 {
		position = buffer.FindAndReplaceSubstring(substring, replacement, position)
		if position.X != -1 && position.Y != -1 {
			replacements++
		}
	}

	return replacements
}

func (buffer *Buffer) MoveUp(i int) bool {
	buffer.CursorPos.Y -= i
	if buffer.CursorPos.Y < 0 {
		buffer.CursorPos.Y = 0
		return false
	}

	if buffer.CursorPos.X >= len(buffer.Contents[buffer.CursorPos.Y]) {
		buffer.CursorPos.X = len(buffer.Contents[buffer.CursorPos.Y])
	}

	return true
}

func (buffer *Buffer) MoveDown(i int) bool {
	buffer.CursorPos.Y += i
	if buffer.CursorPos.Y >= len(buffer.Contents) {
		buffer.CursorPos.Y = len(buffer.Contents) - 1
		return false
	}

	if buffer.CursorPos.X >= len(buffer.Contents[buffer.CursorPos.Y]) {
		buffer.CursorPos.X = len(buffer.Contents[buffer.CursorPos.Y])
	}

	return true
}

func (buffer *Buffer) MoveLeft(i int) bool {
	remainingSteps := i

	for remainingSteps > 0 {
		buffer.CursorPos.X--
		if buffer.CursorPos.X < 0 {
			if buffer.CursorPos.Y > 0 {
				buffer.CursorPos.Y--
				buffer.CursorPos.X = len(buffer.Contents[buffer.CursorPos.Y])
			} else {
				buffer.CursorPos.X = 0
				return false
			}
		}

		remainingSteps--
	}

	return true
}

func (buffer *Buffer) MoveRight(i int) bool {
	remainingSteps := i

	for remainingSteps > 0 {
		buffer.CursorPos.X++
		if buffer.CursorPos.X > len(buffer.Contents[buffer.CursorPos.Y]) {
			if buffer.CursorPos.Y < len(buffer.Contents)-1 {
				buffer.CursorPos.Y++
				buffer.CursorPos.X = 0
			} else {
				buffer.CursorPos.X = len(buffer.Contents[buffer.CursorPos.Y])
				return false
			}
		}

		remainingSteps--
	}

	return true
}

func (buffer *Buffer) WriteRune(r rune) {
	if r == '\n' {
		if buffer.CursorPos.Y == len(buffer.Contents) {
			buffer.Contents = append(buffer.Contents, make(runestring.RuneString, 0))
		} else {
			buffer.Contents = slices.Insert(buffer.Contents, buffer.CursorPos.Y+1, make(runestring.RuneString, 0))
		}

		// Move line content after cursor X to the new line
		line := buffer.Contents[buffer.CursorPos.Y]
		buffer.Contents[buffer.CursorPos.Y+1] = slices.Insert(buffer.Contents[buffer.CursorPos.Y+1], 0, line[buffer.CursorPos.X:]...)
		buffer.Contents[buffer.CursorPos.Y] = line[:buffer.CursorPos.X]

		buffer.MoveDown(1)
		buffer.CursorPos.X = 0
	} else {
		buffer.Contents[buffer.CursorPos.Y] = slices.Insert(buffer.Contents[buffer.CursorPos.Y], buffer.CursorPos.X, r)
		buffer.MoveRight(1)
	}
}

func (buffer *Buffer) WriteString(str runestring.RuneString) {
	for _, r := range str {
		buffer.WriteRune(r)
	}
}

func (buffer *Buffer) Delete(i int) bool {
	remainingSteps := i

	for remainingSteps > 0 {
		buffer.CursorPos.X--
		if buffer.CursorPos.X < 0 {
			if buffer.CursorPos.Y > 0 {
				// Save deleted line text
				deletedLine := buffer.Contents[buffer.CursorPos.Y]

				buffer.CursorPos.Y--

				// Delete line
				buffer.Contents = slices.Delete(buffer.Contents, buffer.CursorPos.Y+1, buffer.CursorPos.Y+2)

				// Append deleted line text to end of current line
				buffer.Contents[buffer.CursorPos.Y] = append(buffer.Contents[buffer.CursorPos.Y], deletedLine...)

				buffer.CursorPos.X = len(buffer.Contents[buffer.CursorPos.Y]) - len(deletedLine)
			} else {
				buffer.CursorPos.X = 0
				return false
			}
		} else {
			buffer.Contents[buffer.CursorPos.Y] = slices.Delete(buffer.Contents[buffer.CursorPos.Y], buffer.CursorPos.X, buffer.CursorPos.X+1)
		}

		remainingSteps--
	}

	return true
}

func (buffer *Buffer) GetCharAtPosition(position Position) rune {
	if position.Y < 0 || position.Y >= len(buffer.Contents) {
		return 0
	}
	line := buffer.Contents[position.Y]

	if position.X == len(line) {
		// Do not return newline for last line if it's empty
		if position.Y == len(buffer.Contents)-1 && len(line) == 0 {
			return 0
		}

		return '\n'
	} else if position.X < 0 || position.X > len(line) {
		return 0
	}

	return rune(line[position.X])
}

func GetOpenFileBuffer(filename string) *Buffer {
	// Replace tilde with home directory
	if filename != "~" && strings.HasPrefix(filename, "~/") {
		homedir, err := os.UserHomeDir()

		if err != nil {
			return nil
		}

		filename = filepath.Join(homedir, filename[2:])
	}

	// Get absolute path of file
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return nil
	}

	for _, buffer := range Buffers {
		if buffer.filename == absFilename {
			return buffer
		}
	}

	return nil
}

func CreateFileBuffer(filename string, openNonExistentFile bool) (*Buffer, error) {
	// Replace tilde with home directory
	if filename != "~" && strings.HasPrefix(filename, "~/") {
		homedir, err := os.UserHomeDir()

		if err != nil {
			return nil, err
		}

		filename = filepath.Join(homedir, filename[2:])
	}

	// Get absolute path of file
	abs, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(abs)
	if !openNonExistentFile {
		if err != nil {
			return nil, err
		}

		if !stat.Mode().IsRegular() {
			return nil, fmt.Errorf("%s is not a regular file", filename)
		}
	}

	if GetBufferByName(filename) != nil {
		return nil, fmt.Errorf("a buffer with the name (%s) is already open", filename)
	}

	if GetBufferByFilename(abs) != nil {
		return nil, fmt.Errorf("%s is already open in another buffer", filename)
	}

	buffer := Buffer{
		Name:      filename,
		Contents:  make([]runestring.RuneString, 1),
		CursorPos: Position{0, 0},
		canSave:   true,
		canEdit:   true,
		filename:  abs,
	}

	// Load file contents if no error was encountered in stat call
	if err == nil {
		err = buffer.Load()

		if err != nil {
			return nil, err
		}
	}

	Buffers = append(Buffers, &buffer)

	return &buffer, nil
}

func CreateBuffer(bufferName string) (*Buffer, error) {
	buffer := Buffer{
		Name:      bufferName,
		Contents:  make([]runestring.RuneString, 1),
		CursorPos: Position{0, 0},
		canSave:   true,
		canEdit:   true,
		filename:  "",
	}

	if GetBufferByName(bufferName) != nil {
		return nil, fmt.Errorf("a buffer with the name (%s) is already open", bufferName)
	}

	Buffers = append(Buffers, &buffer)

	return &buffer, nil
}
