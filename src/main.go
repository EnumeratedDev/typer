package main

import (
	"log"
	"os"
)

var sysconfdir = "/etc/"

func main() {
	// Read config
	readConfig()

	// Read key bindings
	readKeybindings()

	// Read styles
	readStyles()

	// Initialize commands
	initCommands()

	window, err := CreateWindow()
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	if len(os.Args) > 1 {
		for i, file := range os.Args[1:] {
			b, err := CreateFileBuffer(file, true)
			if err != nil {
				window.PrintMessage("Could not open file: " + file)
				continue
			}

			if i == 0 {
				window.CurrentBuffer = b
				Buffers = Buffers[1:]
			}
		}
	}

	for !window.closed {
		window.Draw()
		window.ProcessEvents()
	}

	window.screen.Fini()
	window.screen = nil
}
