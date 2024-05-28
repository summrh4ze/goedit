package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gbin/goncurses"
)

const (
	LARGE_FILE = 50 * 1024 * 1024
)

var (
	Ctrlx_Ctrlc string = fmt.Sprintf(
		"%s %s", goncurses.KeyString(Ctrl('x')), goncurses.KeyString(Ctrl('c')),
	)
	Ctrlx_k string = fmt.Sprintf(
		"%s k", goncurses.KeyString(Ctrl('x')),
	)
	Ctrlx_Ctrlf string = fmt.Sprintf(
		"%s %s", goncurses.KeyString(Ctrl('x')), goncurses.KeyString(Ctrl('f')),
	)
)

type Cursor struct {
	Row int
	Col int
}

type Buffer struct {
	Content         []string
	Cursor          Cursor
	ReadOnlyMode    bool
	MinDisplayedRow int
}

type MinibufferActionStep struct {
	Label string
	Input string
}

type MinibufferContext struct {
	Type        string
	Steps       []MinibufferActionStep
	TotalSteps  int
	CurrentStep int
}

type Editor struct {
	Closed            bool
	MinibufferContext MinibufferContext
	OpenBuffers       []*Buffer
	CurrentBuffer     int
	MaxRows           int
}

func NewEmptyBuffer() *Buffer {
	return &Buffer{
		Content:         make([]string, 0),
		ReadOnlyMode:    false,
		Cursor:          Cursor{0, 0},
		MinDisplayedRow: 0,
	}
}

func getCurrentBuffer(editor *Editor) *Buffer {
	if editor.CurrentBuffer >= 0 && editor.CurrentBuffer < len(editor.OpenBuffers) {
		return editor.OpenBuffers[editor.CurrentBuffer]
	} else {
		return NewEmptyBuffer()
	}
}

func CreateEditor(maxRows int) *Editor {
	editor := &Editor{
		Closed:            false,
		OpenBuffers:       make([]*Buffer, 0, 1),
		CurrentBuffer:     -1,
		MaxRows:           maxRows,
		MinibufferContext: MinibufferContext{},
	}
	return editor
}

func (e *Editor) hasMinibufferContext() bool {
	return e.MinibufferContext.TotalSteps > 0 && e.MinibufferContext.CurrentStep < e.MinibufferContext.TotalSteps
}

func (e *Editor) completedMinibufferContext() bool {
	return e.MinibufferContext.TotalSteps > 0 && e.MinibufferContext.CurrentStep == e.MinibufferContext.TotalSteps
}

func (e *Editor) handleKeybindInput(keybinding string) {
	switch keybinding {
	case Ctrlx_Ctrlc:
		e.Closed = true
	case Ctrlx_k:
		if len(e.OpenBuffers) < 2 {
			e.Closed = true
		} else {
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
	case Ctrlx_Ctrlf:
		e.MinibufferContext = MinibufferContext{
			Type:        "openBuffer",
			Steps:       []MinibufferActionStep{{Label: "Open file: "}},
			TotalSteps:  1,
			CurrentStep: 0,
		}
	}
}

func (e *Editor) handleNormalInput(key goncurses.Key) {
	switch key {
	case goncurses.KEY_ENTER, 10:
		if e.hasMinibufferContext() {
			e.MinibufferContext.CurrentStep += 1
		}
	case goncurses.KEY_BACKSPACE, 127, '\b':
		if e.hasMinibufferContext() {
			input := e.MinibufferContext.Steps[e.MinibufferContext.CurrentStep].Input
			if len(input) > 0 {
				e.MinibufferContext.Steps[e.MinibufferContext.CurrentStep].Input = input[:len(input)-1]
			}
		}
	case Ctrl('n'):
		buffer := getCurrentBuffer(e)
		if buffer.Cursor.Row < len(buffer.Content)-1 {
			buffer.Cursor = Cursor{buffer.Cursor.Row + 1, buffer.Cursor.Col}
			if buffer.Cursor.Row >= buffer.MinDisplayedRow+e.MaxRows {
				buffer.MinDisplayedRow++
			}
		}
	case Ctrl('p'):
		buffer := getCurrentBuffer(e)
		if buffer.Cursor.Row > 0 {
			buffer.Cursor = Cursor{buffer.Cursor.Row - 1, buffer.Cursor.Col}
			if buffer.Cursor.Row < buffer.MinDisplayedRow {
				buffer.MinDisplayedRow--
			}
		}
	default:
		if e.hasMinibufferContext() {
			input := e.MinibufferContext.Steps[e.MinibufferContext.CurrentStep].Input
			e.MinibufferContext.Steps[e.MinibufferContext.CurrentStep].Input = input + goncurses.KeyString(key)
		}
	}
}

func (e *Editor) OpenBuffer(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		// open a fake file. It will be created at first save
		e.OpenBuffers = append(e.OpenBuffers, &Buffer{
			ReadOnlyMode:    false,
			Content:         make([]string, 0),
			Cursor:          Cursor{0, 0},
			MinDisplayedRow: 0,
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
		ReadOnlyMode:    readOnlyMode,
		Content:         content,
		Cursor:          Cursor{0, 0},
		MinDisplayedRow: 0,
	})
	e.CurrentBuffer = len(e.OpenBuffers) - 1

	return nil
}

func (e *Editor) UpdateMinibuffer() {
	switch e.MinibufferContext.Type {
	case "openBuffer":
		if e.completedMinibufferContext() {
			e.OpenBuffer(e.MinibufferContext.Steps[e.MinibufferContext.CurrentStep-1].Input)
			e.MinibufferContext = MinibufferContext{Type: "done", TotalSteps: 1, CurrentStep: 1}
		}
	case "done":
		e.MinibufferContext = MinibufferContext{}
	}
}

func (e *Editor) minibufferDisplay(miniWin *goncurses.Window) {
	if e.hasMinibufferContext() {
		currentStep := e.MinibufferContext.Steps[e.MinibufferContext.CurrentStep]
		miniWin.Erase()
		miniWin.MovePrintf(0, 0, "%s%s", currentStep.Label, currentStep.Input)
		miniWin.Refresh()
	} else if e.completedMinibufferContext() && e.MinibufferContext.Type == "done" {
		miniWin.Erase()
		miniWin.MovePrint(0, 0, "DONE")
		miniWin.Refresh()
	}
}

func (e *Editor) Display(mainWin, miniWin *goncurses.Window) {
	if e.CurrentBuffer >= 0 && e.CurrentBuffer < len(e.OpenBuffers) {
		if !e.hasMinibufferContext() {
			e.OpenBuffers[e.CurrentBuffer].Display(mainWin)
		}
	}
	e.minibufferDisplay(miniWin)
}

func (b *Buffer) Display(bufferWin *goncurses.Window) {
	bufferWin.Erase()

	maxRows, _ := bufferWin.MaxYX()
	for i := b.MinDisplayedRow; i < b.MinDisplayedRow+maxRows; i++ {
		if i < len(b.Content) {
			bufferWin.MovePrint(i-b.MinDisplayedRow, 0, b.Content[i])
		}
	}

	// convert cursor to relative to rows boundary
	bufferWin.Move(b.Cursor.Row-b.MinDisplayedRow, b.Cursor.Col)
}
