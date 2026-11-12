package main

import (
	"os"

	"example.com/note-tools/notes"
)

type Notes = notes.Notes
type Proposal = notes.Proposal
type ToolCall = notes.ToolCall

const (
	maxNoteBytes     = notes.MaxNoteBytes
	maxArgumentBytes = notes.MaxArgumentBytes
)

var noteTools = notes.ToolDefinitions

func newNotes(root *os.Root) *Notes {
	return notes.New(root)
}

func encodeResult(value any) (string, error) {
	return notes.EncodeResult(value)
}
