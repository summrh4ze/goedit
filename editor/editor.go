package editor

import (
	"bufio"
	"fmt"
	"os"
)

const (
	LARGE_FILE = 50 * 1024 * 1024
	TABSIZE    = 8
	ALT        = 27
)

type Editor struct {
	OpenBuffers   []*Buffer
	CurrentBuffer int
}

func CreateEditor() *Editor {
	editor := &Editor{
		OpenBuffers:   []*Buffer{NewEmptyBuffer()},
		CurrentBuffer: 0,
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

func (e *Editor) OpenBuffer(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		// open a fake file. It will be created at first save
		e.OpenBuffers = append(e.OpenBuffers, &Buffer{
			ReadOnlyMode: false,
			Content:      []string{""},
			Cursor:       Cursor{0, 0, 0},
		})
		e.CurrentBuffer = len(e.OpenBuffers) - 1
		return nil
	}
	fileSize := fileInfo.Size()
	readOnlyMode := false
	if fileSize > LARGE_FILE {
		fmt.Println("File too large. Opening file in READ ONLY mode")
		readOnlyMode = true
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return err
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
		fmt.Println(readErr)
		return readErr
	}

	e.OpenBuffers = append(e.OpenBuffers, &Buffer{
		ReadOnlyMode: readOnlyMode,
		Content:      content,
		Cursor:       Cursor{0, 0, 0},
	})
	e.CurrentBuffer = len(e.OpenBuffers) - 1

	return nil
}
