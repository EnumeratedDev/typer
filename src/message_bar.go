package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type TyperMessage struct {
	timestamp int64
	message   string
}

var messageLog = make([]TyperMessage, 0)

func (window *Window) PrintMessage(message string) {
	messageLog = append(messageLog, TyperMessage{timestamp: time.Now().UnixMilli(), message: message})

	err := window.screen.PostEvent(tcell.NewEventInterrupt(nil))
	if err != nil {
		return
	}

	go func() {
		time.Sleep(5 * time.Second)

		err := window.screen.PostEvent(tcell.NewEventInterrupt(nil))
		if err != nil {
			return
		}
	}()
}

func drawMessageBar(window *Window) {
	screen := window.screen

	messageBarStyle := tcell.StyleDefault.Background(CurrentStyle.MessageBarBg).Foreground(CurrentStyle.MessageBarFg)

	sizeX, sizeY := screen.Size()

	messageToPrint := ""
	if len(messageLog) > 0 && time.Since(time.UnixMilli(messageLog[len(messageLog)-1].timestamp)).Seconds() < 5 {
		messageToPrint = messageLog[len(messageLog)-1].message
	}

	for x := 0; x < sizeX; x++ {
		char := ' '
		if x < len(messageToPrint) {
			char = int32(messageToPrint[x])
		}

		if currentInputRequest == nil {
			screen.SetContent(x, sizeY-1, char, nil, messageBarStyle)
		} else {
			screen.SetContent(x, sizeY-2, char, nil, messageBarStyle)
		}

	}
}
