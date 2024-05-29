package editor

import (
	"bufio"
	"os"
)

const (
	LARGE_FILE = 50 * 1024 * 1024
	TABSIZE    = 8
	ALT        = 27
)

type Editor struct {
	OpenBuffers     []*Buffer
	CurrentBuffer   int
	Minibuffer      *Minibuffer
	MinibufferReady <-chan bool
}

func CreateEditor() *Editor {
	ready := make(chan bool, 1)
	editor := &Editor{
		OpenBuffers:     []*Buffer{NewEmptyBuffer()},
		CurrentBuffer:   0,
		Minibuffer:      NewMinibuffer(ready),
		MinibufferReady: ready,
	}
	return editor
}

func (e *Editor) GetCurrentBuffer() *Buffer {
	if e.CurrentBuffer >= 0 && e.CurrentBuffer < len(e.OpenBuffers) {
		return e.OpenBuffers[e.CurrentBuffer]
	}
	return nil
}

func (e *Editor) CloseCurrentBuffer() {
	if len(e.OpenBuffers) > 0 {
		if e.CurrentBuffer > 0 {
			e.OpenBuffers = append(
				e.OpenBuffers[0:e.CurrentBuffer],
				e.OpenBuffers[e.CurrentBuffer+1:]...,
			)
			e.CurrentBuffer = e.CurrentBuffer - 1
		} else { // CurrentBuffer = 0
			e.OpenBuffers = e.OpenBuffers[1:]
		}
	}
}

func (e *Editor) OpenBuffer() {
	if e.Minibuffer.Focused {
		return
	}

	e.Minibuffer.Focused = true
	e.Minibuffer.SetMessage("Find file: ")
	ready := <-e.MinibufferReady
	defer func() {
		e.Minibuffer.Focused = false
	}()

	if !ready {
		e.Minibuffer.SetMessage("Quit")
		return
	}

	path := e.Minibuffer.ConsumeInput()

	if path == "" {
		e.Minibuffer.SetMessage("Empty path")
		return
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		// open a fake file. It will be created at first save
		e.OpenBuffers = append(e.OpenBuffers, &Buffer{
			ReadOnlyMode: false,
			Content:      []string{""},
			Cursor:       Cursor{0, 0, 0},
		})
		e.CurrentBuffer = len(e.OpenBuffers) - 1
		e.Minibuffer.SetMessage("Done")
		return
	}
	fileSize := fileInfo.Size()
	readOnlyMode := false
	if fileSize > LARGE_FILE {
		e.Minibuffer.SetMessage("File too large. Opening file in READ ONLY mode")
		readOnlyMode = true
	}

	file, err := os.Open(path)
	if err != nil {
		e.Minibuffer.SetMessage("Error opening file")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	content := make([]string, 0)

	for scanner.Scan() {
		line := scanner.Text()
		content = append(content, line)
	}

	readErr := scanner.Err()
	if readErr != nil {
		e.Minibuffer.SetMessage("Error reading file")
		return
	}

	e.OpenBuffers = append(e.OpenBuffers, &Buffer{
		ReadOnlyMode: readOnlyMode,
		Content:      content,
		Cursor:       Cursor{0, 0, 0},
	})
	e.CurrentBuffer = len(e.OpenBuffers) - 1
	e.Minibuffer.SetMessage("Done")
}