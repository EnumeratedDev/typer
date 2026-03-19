package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Buffer struct {
	Name     string
	Contents []string

	CursorPos Position
	Offset    Position

	Selection *Selection

	canSave  bool
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

	for lineIndex, line := range buffer.Contents {
		for runeIndex, r := range line + " " {
			drawPosition := Position{runeIndex, lineIndex}

			if x-buffer.Offset.X >= bufferX && y-buffer.Offset.Y >= bufferY {
				// Default style
				style := tcell.StyleDefault.Background(CurrentStyle.BufferAreaBg).Foreground(CurrentStyle.BufferAreaFg)

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

	content, err := os.ReadFile(buffer.filename)
	if err != nil {
		return err
	}

	buffer.Contents = strings.Split(string(content), "\n")
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
		buffer.Contents = append(buffer.Contents, "")
	}

	err := os.WriteFile(buffer.filename, []byte(buffer.GetContentsAsString()), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (buffer *Buffer) GetContentsAsString() string {
	finalText := strings.Builder{}
	for i, line := range buffer.Contents {
		for _, rune := range line {
			finalText.WriteRune(rune)
		}

		if i != len(buffer.Contents)-1 {
			finalText.WriteRune('\n')
		}
	}

	return finalText.String()
}

func (buffer *Buffer) PositionToAbsolutePosition(position Position) int {
	i := 0
	for lineIndex, line := range buffer.Contents {
		if len(line) == 0 {
			line += " "
		}
		for runeIndex, _ := range line + " " {
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
		for runeIndex, _ := range line + " " {
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

func (buffer *Buffer) GetSelectedText() string {
	if buffer.Selection == nil {
		return ""
	}
	if len(buffer.Contents) == 0 {
		return ""
	}

	edge1, edge2 := buffer.GetSelectionEdges()

	selectedText := strings.Builder{}
	if r := buffer.GetCharAtPosition(edge1); r != 0 {
		selectedText.WriteRune(r)
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
			selectedText.WriteRune(buffer.GetCharAtPosition(edge1))
		}
	}

	return selectedText.String()
}

func (buffer *Buffer) CutText(window *Window) (string, int) {
	if buffer.Selection == nil {
		// Cut current line
		cutText := buffer.Contents[buffer.CursorPos.Y] + "\n"

		// Remove line from buffer contents
		if len(buffer.Contents) == 1 {
			buffer.Contents[0] = ""
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
		buffer.Delete(len(cutText))

		// Remove selection
		buffer.Selection = nil

		return cutText, 1
	}
}

func (buffer *Buffer) CopyText() (string, int) {
	if buffer.Selection == nil {
		// Cut current line
		copiedText := buffer.Contents[buffer.CursorPos.Y] + "\n"

		return copiedText, 0
	} else {
		// Copy selection
		return buffer.GetSelectedText(), 1
	}
}

func (buffer *Buffer) PasteText(window *Window, text string) {
	contents := buffer.GetContentsAsString()

	// Remove selected text
	if buffer.Selection != nil {
		edge1, edge2 := buffer.GetSelectionEdges()
		absEdge1 := buffer.PositionToAbsolutePosition(edge1)
		absEdge2 := buffer.PositionToAbsolutePosition(edge2)

		if absEdge2 == len(buffer.Contents) {
			absEdge2 = len(buffer.Contents) - 1
		}

		contents = contents[:absEdge1] + contents[absEdge2+1:]
		buffer.Contents = strings.Split(contents, "\n")
		buffer.CursorPos = buffer.AbsolutePositionToPosition(absEdge1)
		buffer.Selection = nil
	}

	buffer.WriteString(text)
}

func (buffer *Buffer) FindSubstring(substring string, afterPos Position) Position {
	// Return no match if afterPos is larger than the buffer contents size
	contents := buffer.GetContentsAsString()
	absAfterPos := buffer.PositionToAbsolutePosition(afterPos)

	if absAfterPos >= len(contents) {
		return Position{-1, -1}
	}

	index := strings.Index(contents[absAfterPos+1:], substring)

	if index != -1 {
		index += absAfterPos + 1
	}
	return buffer.AbsolutePositionToPosition(index)
}

func (buffer *Buffer) FindAndReplaceSubstring(substring, replacement string, afterPos Position) Position {
	// Return no match if afterPos is larger than the buffer contents size
	contents := buffer.GetContentsAsString()
	absAfterPos := buffer.PositionToAbsolutePosition(afterPos)

	if absAfterPos >= len(contents) {
		return Position{-1, -1}
	}

	index := strings.Index(contents[absAfterPos+1:], substring)

	if index != -1 {
		index += absAfterPos + 1
	}

	// Replace substring with replacement string
	contents = contents[:index] + replacement + contents[index+len(substring):]

	buffer.Contents = strings.Split(contents, "\n")

	return buffer.AbsolutePositionToPosition(index)
}

func (buffer *Buffer) FindAndReplaceAll(substring, replacement string) int {
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
			buffer.Contents = append(buffer.Contents, "")
		} else {
			buffer.Contents = slices.Insert(buffer.Contents, buffer.CursorPos.Y+1, "")
		}

		// Move line content after cursor X to the new line
		line := buffer.Contents[buffer.CursorPos.Y]
		buffer.Contents[buffer.CursorPos.Y+1] = line[buffer.CursorPos.X:] + buffer.Contents[buffer.CursorPos.Y+1]
		buffer.Contents[buffer.CursorPos.Y] = line[:buffer.CursorPos.X]

		buffer.MoveDown(1)
		buffer.CursorPos.X = 0
	} else {
		line := buffer.Contents[buffer.CursorPos.Y]
		buffer.Contents[buffer.CursorPos.Y] = line[:buffer.CursorPos.X] + string(r) + line[buffer.CursorPos.X:]
		buffer.MoveRight(1)
	}
}

func (buffer *Buffer) WriteString(str string) {
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
				buffer.Contents[buffer.CursorPos.Y] += deletedLine

				buffer.CursorPos.X = len(buffer.Contents[buffer.CursorPos.Y]) - len(deletedLine)
			} else {
				buffer.CursorPos.X = 0
				return false
			}
		} else {
			line := buffer.Contents[buffer.CursorPos.Y]
			buffer.Contents[buffer.CursorPos.Y] = line[:buffer.CursorPos.X] + line[buffer.CursorPos.X+1:]
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
		Contents:  make([]string, 1),
		CursorPos: Position{0, 0},
		canSave:   true,
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
		Contents:  make([]string, 1),
		CursorPos: Position{0, 0},
		canSave:   true,
		filename:  "",
	}

	if GetBufferByName(bufferName) != nil {
		return nil, fmt.Errorf("a buffer with the name (%s) is already open", bufferName)
	}

	Buffers = append(Buffers, &buffer)

	return &buffer, nil
}
