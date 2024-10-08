package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gbin/goncurses"
	"org.example.goedit/editor"
	"org.example.goedit/utils"
)

var graphical = regexp.MustCompile(`^[[:graph:][:space:]]*$`)

type Tui struct {
	bufferWindow     *goncurses.Window
	statuslineWindow *goncurses.Window
	minibufferWindow *goncurses.Window
	oldBufferName    string
}

func RunApp(e *editor.Editor) error {
	ui, err := initTUI()
	if err != nil {
		return err
	}
	defer goncurses.End()

	// first render
	ui.displayEditor(e)

	ui.bufferWindow.Timeout(20)
OUT:
	for {
		buffer := e.GetCurrentBuffer()
		key := ui.bufferWindow.GetChar()
		switch key {
		case 0:
		case Ctrl('x'):
			// wait 2 seconds for the next key otherwise drops the ctrl-x ctrl-<?> action
			ui.bufferWindow.Timeout(2000)
			secondKey := ui.bufferWindow.GetChar()
			if secondKey != 0 {
				switch secondKey {
				case Ctrl('c'):
					break OUT
				case Ctrl('f'):
					go e.OpenBuffer()
				case 'k':
					e.CloseCurrentBuffer()
					buffer = e.GetCurrentBuffer()
					if buffer == nil {
						break OUT
					}
				}
			}
			ui.bufferWindow.Timeout(20)
		case Ctrl('f'), goncurses.KEY_RIGHT:
			if e.Minibuffer.Focused {
				e.Minibuffer.MoveForward()
			} else {
				buffer.MoveForward()
			}
		case Ctrl('b'), goncurses.KEY_LEFT:
			if e.Minibuffer.Focused {
				e.Minibuffer.MoveBack()
			} else {
				buffer.MoveBack()
			}
		case Ctrl('g'):
			if e.Minibuffer.Focused {
				e.Minibuffer.RejectAction()
			} else {
				if buffer.IsMarkActive() {
					buffer.ToggleMark()
				}
			}
		case Ctrl('n'), goncurses.KEY_DOWN:
			buffer.MoveDown()
		case Ctrl('p'), goncurses.KEY_UP:
			buffer.MoveUp()
		case Ctrl('a'):
			if e.Minibuffer.Focused {
				e.Minibuffer.MoveStartLine()
			} else {
				buffer.MoveStartLine()
			}
		case Ctrl('e'):
			if e.Minibuffer.Focused {
				e.Minibuffer.MoveEndLine()
			} else {
				buffer.MoveEndLine()
			}
		case Ctrl('k'):
			buffer.DeleteToEnd()
		case Ctrl('y'):
			buffer.Yank()
		case Ctrl('w'):
			buffer.Cut()
		case Ctrl('d'):
			buffer.DeleteAfter(true)
		case Ctrl('u'):
			buffer.Undo()
		case 27: // Alt-<?>
			secondKey := ui.bufferWindow.GetChar()
			switch secondKey {
			case 'f':
				if e.Minibuffer.Focused {
					e.Minibuffer.MoveForwardWord()
				} else {
					buffer.MoveForwardWord()
				}
			case 'b':
				if e.Minibuffer.Focused {
					e.Minibuffer.MoveBackWord()
				} else {
					buffer.MoveBackWord()
				}
			case '>':
				buffer.MoveEndFile()
			case '<':
				buffer.MoveStartFile()
			case goncurses.KEY_BACKSPACE, 127: // Alt-Backspace
				buffer.DeleteWordBefore()
			case ' ':
				buffer.ToggleMark()
			case 'w':
				buffer.Copy()
			}
		case goncurses.KEY_ENTER, 10:
			if e.Minibuffer.Focused {
				e.Minibuffer.ConfirmAction()
			} else {
				buffer.Insert("\n", true)
			}
		case goncurses.KEY_BACKSPACE, 127, '\b':
			if e.Minibuffer.Focused {
				e.Minibuffer.DeleteAtCol()
			} else {
				buffer.DeleteBefore()
			}
		case goncurses.KEY_TAB:
			buffer.Insert("\t", true)
		default:
			if graphical.MatchString(goncurses.KeyString(key)) {
				if e.Minibuffer.Focused {
					e.Minibuffer.InsertAtCol(goncurses.KeyString(key))
				} else {
					buffer.Insert(goncurses.KeyString(key), true)
				}
			}
		}

		ui.displayEditor(e)
	}
	return nil
}

func initTUI() (*Tui, error) {
	bufferWindow, err := goncurses.Init()
	if err != nil {
		return nil, err
	}

	goncurses.CBreak(true)
	goncurses.Echo(false)
	goncurses.Raw(true)
	goncurses.StartColor()
	goncurses.InitColor(199, 196, 188, 184)
	goncurses.InitColor(200, 113, 125, 129)
	goncurses.InitColor(201, 984, 945, 780)
	//goncurses.InitColor(202, 659, 600, 518)
	goncurses.InitColor(202, 400, 361, 329)
	goncurses.InitPair(1, 201, 199)
	goncurses.InitPair(2, 201, 200)
	goncurses.InitPair(3, 202, 200)

	bufferWindow.ScrollOk(true)
	bufferWindow.Keypad(true)

	maxRows, maxCols := bufferWindow.MaxYX()
	bufferWindow.Resize(maxRows-2, maxCols)

	statuslineWindow, err := goncurses.NewWindow(1, maxCols, maxRows-2, 0)
	if err != nil {
		return nil, err
	}

	minibufferWindow, err := goncurses.NewWindow(1, maxCols, maxRows-1, 0)
	if err != nil {
		return nil, err
	}
	minibufferWindow.Keypad(true)

	bufferWindow.ColorOn(2)
	bufferWindow.SetBackground(goncurses.ColorPair(2))
	bufferWindow.Refresh()

	minibufferWindow.ColorOn(2)
	minibufferWindow.SetBackground(goncurses.ColorPair(2))
	minibufferWindow.Refresh()

	statuslineWindow.ColorOn(1)
	statuslineWindow.SetBackground(goncurses.ColorPair(1))
	statuslineWindow.Refresh()

	return &Tui{
		bufferWindow:     bufferWindow,
		statuslineWindow: statuslineWindow,
		minibufferWindow: minibufferWindow,
	}, nil
}

func (ui *Tui) displayEditor(e *editor.Editor) {
	buffer := e.GetCurrentBuffer()
	if buffer != nil && !e.Minibuffer.Focused {
		ui.displayBuffer(buffer)
		ui.displayStatusLine(buffer)
	}
	ui.displayMinibuffer(e.Minibuffer)
}

func (ui *Tui) displayBuffer(b *editor.Buffer) {
	ui.bufferWindow.Erase()

	maxRows, maxCols := ui.bufferWindow.MaxYX()

	data, totalRows, cursor, mark := b.GetContent(maxRows, editor.TABSIZE)
	lines := strings.Split(data, "\n")

	digits := len(fmt.Sprint(totalRows))

	for i, line := range lines {
		if b.GetBaseRow()+i == cursor.Row {
			ui.bufferWindow.ColorOn(2)
		} else {
			ui.bufferWindow.ColorOn(3)
		}
		ui.bufferWindow.MovePrintf(i, 0, "%*d ", digits, b.GetBaseRow()+i)
		ui.bufferWindow.ColorOn(2)

		for j, ch := range utils.Texp(line, editor.TABSIZE) {
			if mark.Active {
				//panic(fmt.Sprintf("%v\n", mark))
				if mark.Cursor.Row < cursor.Row {
					if b.GetBaseRow()+i > mark.Cursor.Row && b.GetBaseRow()+i < cursor.Row {
						ui.bufferWindow.AttrOn(goncurses.A_REVERSE)
					} else if b.GetBaseRow()+i == mark.Cursor.Row && j >= mark.Cursor.Col {
						ui.bufferWindow.AttrOn(goncurses.A_REVERSE)
					} else if b.GetBaseRow()+i == cursor.Row && j < cursor.Col {
						ui.bufferWindow.AttrOn(goncurses.A_REVERSE)
					}
				} else if mark.Cursor.Row > cursor.Row {
					if b.GetBaseRow()+i > cursor.Row && b.GetBaseRow()+i < mark.Cursor.Row {
						ui.bufferWindow.AttrOn(goncurses.A_REVERSE)
					} else if b.GetBaseRow()+i == mark.Cursor.Row && j <= mark.Cursor.Col {
						ui.bufferWindow.AttrOn(goncurses.A_REVERSE)
					} else if b.GetBaseRow()+i == cursor.Row && j >= cursor.Col {
						ui.bufferWindow.AttrOn(goncurses.A_REVERSE)
					}
				} else if b.GetBaseRow()+i == mark.Cursor.Row {
					if (mark.Cursor.Col <= j && cursor.Col > j) || (cursor.Col <= j && mark.Cursor.Col > j) {
						ui.bufferWindow.AttrOn(goncurses.A_REVERSE)
					}
				}
			}
			if digits+1+cursor.Col >= maxCols {
				surplus := ((digits + 1 + cursor.Col) - maxCols) + 1
				if j < surplus {
					continue
				} else {
					ui.bufferWindow.MovePrint(i, digits+1+(j-surplus), string(ch))
				}
			} else {
				ui.bufferWindow.MovePrint(i, digits+1+j, string(ch))
			}
			ui.bufferWindow.AttrOff(goncurses.A_REVERSE)
		}

	}

	// convert cursor to relative to rows boundary
	if digits+1+cursor.Col >= maxCols {
		ui.bufferWindow.Move(cursor.Row-b.GetBaseRow(), maxCols-1)
	} else {
		ui.bufferWindow.Move(cursor.Row-b.GetBaseRow(), cursor.Col+digits+1)
	}
}

func (ui *Tui) displayStatusLine(b *editor.Buffer) {
	if ui.oldBufferName != b.Name {
		ui.oldBufferName = b.Name
		ui.statuslineWindow.Erase()
		ui.statuslineWindow.Print(b.Name)
		ui.statuslineWindow.Refresh()
	}
}

func (ui *Tui) displayMinibuffer(m *editor.Minibuffer) {
	if m.Focused || m.Dirty {
		line := m.GetLine()
		ui.minibufferWindow.Erase()
		ui.minibufferWindow.MovePrint(0, 0, line)
		ui.minibufferWindow.Move(0, m.GetCursor())
		ui.minibufferWindow.Refresh()
	}
}

func Ctrl(ch goncurses.Key) goncurses.Key {
	return ch & 0x1f
}
