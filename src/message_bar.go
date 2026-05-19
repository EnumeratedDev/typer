package main

import (
	"slices"
	"time"
	"typer/runestring"

	"github.com/gdamore/tcell/v2"
)

type TyperMessageUrgency uint

const (
	TYPER_MESSAGE_INFO TyperMessageUrgency = iota
	TYPER_MESSAGE_WARNING
	TYPER_MESSAGE_ERROR
)

type TyperMessage struct {
	Timestamp int64
	Urgency   TyperMessageUrgency
	Message   string
}

var lastMessage *TyperMessage

func (window *Window) PrintMessage(message string, urgency TyperMessageUrgency) {
	lastMessage = &TyperMessage{Timestamp: time.Now().UnixMilli(), Message: message, Urgency: urgency}

	logsBuffer := GetBufferByName("Typer Logs")
	if logsBuffer != nil {
		messageToPrint := ""
		switch lastMessage.Urgency {
		case TYPER_MESSAGE_INFO:
			messageToPrint = "[INFO] "
		case TYPER_MESSAGE_WARNING:
			messageToPrint = "[WARNING] "
		case TYPER_MESSAGE_ERROR:
			messageToPrint = "[ERROR] "
		default:
			messageToPrint = "[???] "
		}

		messageToPrint += "[" + time.UnixMilli(lastMessage.Timestamp).Format("15:04:05") + "] "
		messageToPrint += lastMessage.Message + "\n"

		if len(logsBuffer.Contents) >= 1000 {
			logsBuffer.Contents = slices.Delete(logsBuffer.Contents, 0, len(logsBuffer.Contents)-999)
		}

		logsBuffer.CursorPos = Position{
			X: len(logsBuffer.Contents[len(logsBuffer.Contents)-1]),
			Y: len(logsBuffer.Contents) - 1,
		}

		logsBuffer.WriteString(runestring.RuneString(messageToPrint))
	}

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
	if lastMessage != nil && time.Since(time.UnixMilli(lastMessage.Timestamp)).Seconds() < 5 {
		switch lastMessage.Urgency {
		case TYPER_MESSAGE_INFO:
			if Config.ColorMessageBar {
				messageBarStyle = messageBarStyle.Foreground(CurrentStyle.SyntaxInfo)
			}
			messageToPrint = "[INFO] "
		case TYPER_MESSAGE_WARNING:
			if Config.ColorMessageBar {
				messageBarStyle = messageBarStyle.Foreground(CurrentStyle.SyntaxWarning)
			}
			messageToPrint = "[WARNING] "
		case TYPER_MESSAGE_ERROR:
			if Config.ColorMessageBar {
				messageBarStyle = messageBarStyle.Foreground(CurrentStyle.SyntaxError)
			}
			messageToPrint = "[ERROR] "
		default:
			messageToPrint = "[???] "
		}
		messageToPrint += lastMessage.Message
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
