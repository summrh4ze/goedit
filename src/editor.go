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
	Ctrlx_Ctrlk string = fmt.Sprintf(
		"%s %s", goncurses.KeyString(Ctrl('x')), goncurses.KeyString(Ctrl('k')),
	)
	Ctrln string = goncurses.KeyString(Ctrl('n'))
	Ctrlp string = goncurses.KeyString(Ctrl('p'))
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

type EditorAction struct {
	Name     string
	Callback func()
}

type Editor struct {
	Closed        bool
	Keybindings   map[string]*EditorAction
	OpenBuffers   []*Buffer
	CurrentBuffer int
	MaxRows       int
}

func getCurrentBuffer(editor *Editor) *Buffer {
	if editor.CurrentBuffer >= 0 && editor.CurrentBuffer < len(editor.OpenBuffers) {
		return editor.OpenBuffers[editor.CurrentBuffer]
	} else {
		return &Buffer{
			Content:         make([]string, 0),
			ReadOnlyMode:    false,
			Cursor:          Cursor{0, 0},
			MinDisplayedRow: 0,
		}
	}
}

func initKeybindings(editor *Editor) {
	editor.Keybindings[Ctrlx_Ctrlc] = &EditorAction{
		Name:     "Close Editor",
		Callback: func() { editor.Closed = true },
	}

	editor.Keybindings[Ctrlx_Ctrlk] = &EditorAction{
		Name: "Close Buffer",
		Callback: func() {
			if len(editor.OpenBuffers) < 2 {
				editor.Closed = true
			} else {
				if editor.CurrentBuffer > 0 {
					editor.OpenBuffers = append(
						editor.OpenBuffers[0:editor.CurrentBuffer],
						editor.OpenBuffers[editor.CurrentBuffer+1:]...,
					)
					editor.CurrentBuffer = editor.CurrentBuffer - 1
				} else { // CurrentBuffer = 0
					editor.OpenBuffers = editor.OpenBuffers[1:]
				}
			}
		},
	}

	editor.Keybindings[Ctrln] = &EditorAction{
		Name: "Move Cursor Down",
		Callback: func() {
			buffer := getCurrentBuffer(editor)
			if buffer.Cursor.Row < len(buffer.Content)-1 {
				buffer.Cursor = Cursor{buffer.Cursor.Row + 1, buffer.Cursor.Col}
				if buffer.Cursor.Row >= buffer.MinDisplayedRow+editor.MaxRows {
					buffer.MinDisplayedRow++
				}
			}
		},
	}

	editor.Keybindings[Ctrlp] = &EditorAction{
		Name: "Move Cursor Up",
		Callback: func() {
			buffer := getCurrentBuffer(editor)
			if buffer.Cursor.Row > 0 {
				buffer.Cursor = Cursor{buffer.Cursor.Row - 1, buffer.Cursor.Col}
				if buffer.Cursor.Row < buffer.MinDisplayedRow {
					buffer.MinDisplayedRow--
				}
			}
		},
	}
}

func CreateEditor(window *goncurses.Window) *Editor {
	maxRows, _ := window.MaxYX()
	editor := &Editor{
		Closed:        false,
		OpenBuffers:   make([]*Buffer, 0, 1),
		Keybindings:   make(map[string]*EditorAction),
		CurrentBuffer: -1,
		MaxRows:       maxRows,
	}
	initKeybindings(editor)
	return editor
}

func (e *Editor) handleInput(keybinding string) {
	if editorAction, ok := e.Keybindings[keybinding]; ok {
		editorAction.Callback()
	}
}

func (e *Editor) OpenBuffer(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return err
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
		//fmt.Println(line)
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

func (e *Editor) Display(window *goncurses.Window) {
	if e.CurrentBuffer >= 0 && e.CurrentBuffer < len(e.OpenBuffers) {
		e.OpenBuffers[e.CurrentBuffer].Display(window)
	}
}

func (b *Buffer) Display(window *goncurses.Window) {
	window.Erase()

	maxRows, _ := window.MaxYX()
	for i := b.MinDisplayedRow; i < b.MinDisplayedRow+maxRows; i++ {
		if i < len(b.Content) {
			window.MovePrint(i-b.MinDisplayedRow, 0, b.Content[i])
		}
	}

	// convert cursor to relative to rows boundary
	window.Move(b.Cursor.Row-b.MinDisplayedRow, b.Cursor.Col)

}
