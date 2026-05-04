package main

import (
	"log"

	flag "github.com/spf13/pflag"
)

var sysconfdir = "/etc/"

var configDirFlag = flag.StringP("config", "c", "", "Path to config directory")

func main() {
	// Read flags
	readFlags()

	// Read main config
	readMainConfig()

	// Read keybindings config
	readKeybindingsConfig()

	// Read styles directory
	readStyles()

	// Read syntax directory
	ReadSyntaxHighlighters()

	// Initialize commands
	initCommands()

	window, err := CreateWindow()
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	if flag.NArg() > 0 {
		for i, file := range flag.Args() {
			b, err := CreateFileBuffer(file, true)
			if err != nil {
				window.PrintMessage("Could not open file: "+file, TYPER_MESSAGE_ERROR)
				continue
			}

			if i == 0 {
				window.CurrentBuffer = b
				Buffers = Buffers[1:]
			}
		}
	}

	// Create logs buffer
	logsBuffer, err := CreateBuffer("Logs")
	if err != nil {
		log.Fatalf("Could not create logs buffer")
	}
	logsBuffer.filetype = "typer_logs"
	logsBuffer.canEdit = false

	for !window.closed {
		window.Draw()
		window.ProcessEvents()
	}

	window.screen.Fini()
	window.screen = nil
}

func readFlags() {
	flag.Parse()
}
