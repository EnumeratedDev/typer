package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

type SyntaxRule struct {
	Type      string `yaml:"type"`
	Regex     string `yaml:"regex"`
	Multiline bool   `yaml:"multiline"`
}

type Syntax struct {
	Filetype  string       `yaml:"filetype"`
	Filenames string       `yaml:"filenames"`
	Rules     []SyntaxRule `yaml:"rules"`
}

type ParsedSyntax struct {
	StartIndex int
	EndIndex   int
	Type       string
}

var AvailableSyntaxes map[string]Syntax = make(map[string]Syntax)

func ReadSyntaxHighlighters() {
	// Get syntax directory path
	syntaxDirPath := GetConfigPath("syntax")

	// Ensure directory exists at path
	if stat, err := os.Stat(syntaxDirPath); syntaxDirPath == "" || err != nil || !stat.IsDir() {
		return
	}

	// Get directory entries
	entries, err := os.ReadDir(syntaxDirPath)
	if err != nil {
		log.Fatalf("Could not read syntax directory: %s", err)
	}

	// Read entries in directory
	for _, entry := range entries {
		entryPath := filepath.Join(syntaxDirPath, entry.Name())

		data, err := os.ReadFile(entryPath)
		if err != nil {
			log.Fatalf("Could not read syntax file (%s): %s", entryPath, err)
		}

		syntax := Syntax{}
		err = yaml.Unmarshal(data, &syntax)
		if err != nil {
			log.Fatalf("Could not read syntax file (%s): %s", entryPath, err)
		}

		if _, ok := AvailableSyntaxes[syntax.Filetype]; !ok {
			AvailableSyntaxes[syntax.Filetype] = syntax
		}
	}
}

func HighlightString(s string, filetype string) (parsedSyntaxes []ParsedSyntax, err error) {
	// Get syntax for filetype
	syntax, ok := AvailableSyntaxes[filetype]
	if !ok {
		return nil, nil
	}

	for _, rule := range syntax.Rules {

		regex := rule.Regex
		if rule.Multiline {
			regex = "(?m)" + rule.Regex
		}
		r, err := regexp.Compile(regex)
		if err != nil {
			return nil, err
		}

		matches := r.FindAllStringIndex(s, -1)
		for _, match := range matches {
			skip := false
			for _, parsedSyntax := range parsedSyntaxes {
				if (match[0] >= parsedSyntax.StartIndex && match[0] < parsedSyntax.EndIndex) || (match[1] >= parsedSyntax.StartIndex && match[1] < parsedSyntax.EndIndex) {
					skip = true
				}
			}
			if skip {
				continue
			}

			parsedSyntax := ParsedSyntax{
				StartIndex: match[0],
				EndIndex:   match[1],
				Type:       rule.Type,
			}

			parsedSyntaxes = append(parsedSyntaxes, parsedSyntax)
		}
	}

	return parsedSyntaxes, nil
}
