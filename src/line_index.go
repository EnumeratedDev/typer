package main

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

func drawLineIndex(window *Window) {
	screen := window.screen
	buffer := window.CurrentBuffer

	lineIndexStyle := tcell.StyleDefault.Background(CurrentStyle.LineIndexBg).Foreground(CurrentStyle.LineIndexFg)

	lineIndexSize := getLineIndexSize(window)

	_, bufferY1, _, bufferY2 := window.GetTextAreaDimensions()

	lineIndex := 1 + buffer.Offset.Y
	for y := bufferY1; y <= bufferY2; y++ {
		if lineIndex > len(buffer.Contents) {
			if Config.ExtendLineIndex {
				for x := 0; x < lineIndexSize; x++ {
					screen.SetContent(x, y, ' ', nil, lineIndexStyle)
				}
				continue
			} else {
				break
			}
		}

		for x := 0; x < lineIndexSize; x++ {
			screen.SetContent(x, y, ' ', nil, lineIndexStyle)
		}

		text := strconv.Itoa(lineIndex)

		drawText(screen, lineIndexSize-len(text)-1, y, lineIndexSize, y, lineIndexStyle, text)

		lineIndex++
	}
}

func getLineIndexSize(window *Window) int {
	i := len(window.CurrentBuffer.Contents)
	if i == 0 {
		return 4
	}
	count := 0
	for i != 0 {
		i /= 10
		count++
	}

	if count < 3 {
		count = 4
	} else {
		count += 1
	}

	return count
}
