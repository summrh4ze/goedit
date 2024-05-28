package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/gbin/goncurses"
)

const (
	LARGE_FILE = 50 * 1024 * 1024
	TABSIZE    = 8
	ALT        = 27
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
	Altf string = fmt.Sprintf("%s f", goncurses.KeyString(ALT))
	Altb string = fmt.Sprintf("%s b", goncurses.KeyString(ALT))
)

type Cursor struct {
	Row int
	Col int
	Mem int
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
		Content:         []string{"\tGOEdit!", "To open a file use Ctrl-X Ctrl-F"},
		ReadOnlyMode:    false,
		Cursor:          Cursor{0, 0, 0},
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
		OpenBuffers:       []*Buffer{NewEmptyBuffer()},
		CurrentBuffer:     0,
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
	case Altf:
		buffer := getCurrentBuffer(e)
		var line string
		if len(buffer.Content) > buffer.Cursor.Row && buffer.Cursor.Row >= 0 {
			line = buffer.Content[buffer.Cursor.Row]
		}

		// need to expand the string before checking characters
		if buffer.Cursor.Col < tlen(line, TABSIZE) {
			expLine := texp(line, TABSIZE)
			row := buffer.Cursor.Row
			col := buffer.Cursor.Col
			if expLine[buffer.Cursor.Col] != ' ' {
				for i := buffer.Cursor.Col + 1; i < len(expLine); i++ {
					if isDelimiter(expLine[i]) {
						col = i
						break
					}
				}
				if col > buffer.Cursor.Col {
					buffer.Cursor = Cursor{buffer.Cursor.Row, col, col}
				} else {
					// word goes until the end of the line so stop at the very end
					buffer.Cursor = Cursor{buffer.Cursor.Row, tlen(line, TABSIZE), tlen(line, TABSIZE)}
				}
			} else {
				// search for the first non whitespace character
				for i := buffer.Cursor.Col + 1; i < len(expLine); i++ {
					if !isDelimiter(expLine[i]) {
						col = i
						break
					}
				}
				if col > buffer.Cursor.Col {
					buffer.Cursor = Cursor{buffer.Cursor.Row, col, col}
				} else if buffer.Cursor.Row+1 < len(buffer.Content) {
					// not found on current line search the next line
					row += 1
					col = 0
					nextLine := buffer.Content[buffer.Cursor.Row+1]
					nexpLine := texp(nextLine, TABSIZE)
					for i := 0; i < len(nexpLine); i++ {
						if !isDelimiter(nexpLine[i]) {
							col = i
							break
						}
					}
					buffer.Cursor = Cursor{row, col, col}
				}
			}
		} else if tlen(line, TABSIZE) == 0 || buffer.Cursor.Col == tlen(line, TABSIZE) {
			if buffer.Cursor.Row+1 < len(buffer.Content) {
				nexpLine := texp(buffer.Content[buffer.Cursor.Row+1], TABSIZE)
				col := 0
				for i := 0; i < len(nexpLine); i++ {
					if !isDelimiter(nexpLine[i]) {
						col = i
						break
					}
				}
				buffer.Cursor = Cursor{buffer.Cursor.Row + 1, col, col}
			}
		}
		if buffer.Cursor.Row >= buffer.MinDisplayedRow+e.MaxRows {
			buffer.MinDisplayedRow++
		}
		//case Altb:

	}
}

func isDelimiter(b byte) bool {
	delimiters := []byte(" `~!@#$%^&*()-=+[{]}\\|;:'\",.<>/?")
	for _, c := range delimiters {
		if b == c {
			return true
		}
	}
	return false
}

func tlen(str string, tabsize int) int {
	tlen := 0
	if str == "" {
		return tlen
	}

	nonTabs := 0
	if str[0] != '\t' {
		nonTabs = 1
		tlen += 1
	} else {
		tlen += tabsize
	}
	prevChar := str[0]
	for i := 1; i < len(str); i++ {
		if str[i] == '\t' && prevChar == '\t' {
			tlen += tabsize
			prevChar = str[i]
		} else if str[i] == '\t' && prevChar != '\t' {
			tlen += tabsize - nonTabs%tabsize
			nonTabs = 0
		} else if str[i] != '\t' {
			tlen += 1
			nonTabs += 1
		}
	}
	return tlen
}

func texp(str string, tabsize int) string {
	res := ""
	if str == "" {
		return res
	}

	nonTabs := 0
	if str[0] != '\t' {
		nonTabs = 1
		res += string(str[0])
	} else {
		res += string(bytes.Repeat([]byte(" "), tabsize))
	}
	prevChar := str[0]
	for i := 1; i < len(str); i++ {
		if str[i] == '\t' && prevChar == '\t' {
			res += string(bytes.Repeat([]byte(" "), tabsize))
			prevChar = str[i]
		} else if str[i] == '\t' && prevChar != '\t' {
			res += string(bytes.Repeat([]byte(" "), tabsize-nonTabs%tabsize))
			nonTabs = 0
		} else if str[i] != '\t' {
			res += string(str[i])
			nonTabs += 1
		}
	}
	return res
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
			nextLine := buffer.Content[buffer.Cursor.Row+1]
			if buffer.Cursor.Col > tlen(nextLine, TABSIZE) || buffer.Cursor.Mem > tlen(nextLine, TABSIZE) {
				buffer.Cursor = Cursor{buffer.Cursor.Row + 1, tlen(nextLine, TABSIZE), buffer.Cursor.Mem}
			} else if buffer.Cursor.Mem < buffer.Cursor.Col {
				buffer.Cursor = Cursor{buffer.Cursor.Row + 1, buffer.Cursor.Col, buffer.Cursor.Mem}
			} else {
				buffer.Cursor = Cursor{buffer.Cursor.Row + 1, buffer.Cursor.Mem, buffer.Cursor.Mem}
			}

			if buffer.Cursor.Row >= buffer.MinDisplayedRow+e.MaxRows {
				buffer.MinDisplayedRow++
			}
		}
	case Ctrl('p'):
		buffer := getCurrentBuffer(e)
		if buffer.Cursor.Row > 0 {
			prevLine := buffer.Content[buffer.Cursor.Row-1]
			if buffer.Cursor.Col > tlen(prevLine, TABSIZE) || buffer.Cursor.Mem > tlen(prevLine, TABSIZE) {
				buffer.Cursor = Cursor{buffer.Cursor.Row - 1, tlen(prevLine, TABSIZE), buffer.Cursor.Mem}
			} else if buffer.Cursor.Mem < buffer.Cursor.Col {
				buffer.Cursor = Cursor{buffer.Cursor.Row - 1, buffer.Cursor.Col, buffer.Cursor.Mem}
			} else {
				buffer.Cursor = Cursor{buffer.Cursor.Row - 1, buffer.Cursor.Mem, buffer.Cursor.Mem}
			}

			if buffer.Cursor.Row < buffer.MinDisplayedRow {
				buffer.MinDisplayedRow--
			}
		}
	case Ctrl('f'):
		buffer := getCurrentBuffer(e)
		var line string
		if len(buffer.Content) > buffer.Cursor.Row && buffer.Cursor.Row >= 0 {
			line = buffer.Content[buffer.Cursor.Row]
		}
		if buffer.Cursor.Col < tlen(line, TABSIZE) {
			buffer.Cursor = Cursor{buffer.Cursor.Row, buffer.Cursor.Col + 1, buffer.Cursor.Col + 1}
		} else if tlen(line, TABSIZE) == 0 || buffer.Cursor.Col == tlen(line, TABSIZE) {
			if buffer.Cursor.Row+1 < len(buffer.Content) {
				buffer.Cursor = Cursor{buffer.Cursor.Row + 1, 0, 0}
			}
		}
		if buffer.Cursor.Row >= buffer.MinDisplayedRow+e.MaxRows {
			buffer.MinDisplayedRow++
		}
	case Ctrl('b'):
		buffer := getCurrentBuffer(e)
		var line string
		if len(buffer.Content) > buffer.Cursor.Row-1 && buffer.Cursor.Row-1 >= 0 {
			line = buffer.Content[buffer.Cursor.Row-1]
		}
		if buffer.Cursor.Col > 0 {
			buffer.Cursor = Cursor{buffer.Cursor.Row, buffer.Cursor.Col - 1, buffer.Cursor.Col - 1}
		} else if buffer.Cursor.Col == 0 && buffer.Cursor.Row-1 >= 0 {
			buffer.Cursor = Cursor{buffer.Cursor.Row - 1, tlen(line, TABSIZE), tlen(line, TABSIZE)}
		}
		if buffer.Cursor.Row < buffer.MinDisplayedRow {
			buffer.MinDisplayedRow--
		}
	case Ctrl('a'):
		buffer := getCurrentBuffer(e)
		var line string
		if len(buffer.Content) > buffer.Cursor.Row && buffer.Cursor.Row >= 0 {
			line = buffer.Content[buffer.Cursor.Row]
		}
		expLine := texp(line, TABSIZE)
		firstNonWh := 0
		for i := 0; i < len(expLine); i++ {
			if expLine[i] != ' ' {
				firstNonWh = i
				break
			}
		}
		if buffer.Cursor.Col == 0 || buffer.Cursor.Col > firstNonWh {
			buffer.Cursor = Cursor{buffer.Cursor.Row, firstNonWh, firstNonWh}
		} else {
			buffer.Cursor = Cursor{buffer.Cursor.Row, 0, 0}
		}
	case Ctrl('e'):
		buffer := getCurrentBuffer(e)
		var line string
		if len(buffer.Content) > buffer.Cursor.Row && buffer.Cursor.Row >= 0 {
			line = buffer.Content[buffer.Cursor.Row]
		}
		buffer.Cursor = Cursor{buffer.Cursor.Row, tlen(line, TABSIZE), tlen(line, TABSIZE)}
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
			Content:         []string{""},
			Cursor:          Cursor{0, 0, 0},
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
		Cursor:          Cursor{0, 0, 0},
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
