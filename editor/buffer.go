package editor

import (
	"bytes"
	"fmt"

	"org.example.goedit/utils"
)

const (
	GAP_LEN       = 1000
	GAP_THRESHOLD = 10
	UNDO_SIZE     = 1000
)

type Cursor struct {
	Row int
	Col int
}

type Mark struct {
	Active bool
	Cursor Cursor
}

type Buffer struct {
	parent       *Editor
	content      []byte
	linePosMem   int
	gapStart     int
	gapEnd       int
	baseRow      int
	markActive   bool
	markPos      int
	killBuffer   []byte
	ReadOnlyMode bool
	Name         string
	undo         *UndoStack
}

func NewEmptyBuffer(parent *Editor) *Buffer {
	content := "\tGOEdit!\nTo open a file use Ctrl-X Ctrl-F"
	buf := make([]byte, GAP_LEN, len(content)+GAP_LEN)
	buf = append(buf, content...)
	return &Buffer{
		parent:       parent,
		Name:         "scratch",
		content:      buf,
		ReadOnlyMode: false,
		gapStart:     0,
		gapEnd:       GAP_LEN,
		undo:         NewUndo(UNDO_SIZE),
	}
}

func NewBuffer(parent *Editor, name string, content []byte, readOnly bool) *Buffer {
	if readOnly {
		return &Buffer{
			parent:       parent,
			Name:         name,
			content:      content,
			ReadOnlyMode: true,
			undo:         NewUndo(0),
		}
	} else {
		buf := make([]byte, GAP_LEN, len(content)+GAP_LEN)
		buf = append(buf, content...)
		return &Buffer{
			parent:       parent,
			Name:         name,
			content:      buf,
			ReadOnlyMode: false,
			gapStart:     0,
			gapEnd:       GAP_LEN,
			undo:         NewUndo(UNDO_SIZE),
		}
	}
}

func (b *Buffer) GetBaseRow() int {
	return b.baseRow
}

func (b *Buffer) GetContent(count int, tabsize int) (string, int, Cursor, Mark) {
	row := 0
	col := 0
	cursor := Cursor{}
	mark := Mark{}
	newLinesPos := make([]int, 0)

	markPos := b.markPos
	if b.gapStart < b.markPos {
		markPos = b.markPos + (b.gapEnd - b.gapStart)
	}

	nonTabs := 0
	var prevByte byte
	for i, currentByte := range b.content {
		if i > b.gapStart && i < b.gapEnd {
			continue
		} else if i == b.gapStart {
			cursor.Row = row
			cursor.Col = col
			if i == markPos {
				mark.Cursor.Row = row
				mark.Cursor.Col = col
			}
			continue
		}
		if i == markPos {
			mark.Cursor.Row = row
			mark.Cursor.Col = col
		}
		if currentByte == '\t' {
			if col == 0 {
				prevByte = currentByte
				col += tabsize
			} else if prevByte == '\t' {
				col += tabsize
				prevByte = currentByte
			} else {
				col += tabsize - nonTabs%tabsize
				nonTabs = 0
			}
		} else if currentByte == '\n' {
			row++
			col = 0
			nonTabs = 0
			newLinesPos = append(newLinesPos, i)
		} else {
			col++
			nonTabs++
		}
	}

	if b.gapStart == len(b.content) {
		cursor.Row = row
		cursor.Col = col
	}

	if markPos == len(b.content) {
		mark.Cursor.Row = row
		mark.Cursor.Col = col
	}

	if cursor.Row >= b.baseRow+count {
		b.baseRow = cursor.Row + 1 - count
	} else if cursor.Row < b.baseRow {
		b.baseRow = cursor.Row
	}

	startSlice := 0
	endSlice := len(b.content)
	totalRows := len(newLinesPos) + 1

	if b.baseRow == 0 {
		if b.baseRow+count < totalRows {
			endSlice = newLinesPos[b.baseRow+count-1] + 1
		}
	} else {
		startSlice = newLinesPos[b.baseRow-1] + 1
		if b.baseRow+count < totalRows {
			endSlice = newLinesPos[b.baseRow+count-1] + 1
		}
	}

	// if cursor(gap) is not in range [startSlice:endSlice] something is wrong
	if b.gapStart < startSlice || b.gapEnd > endSlice {
		panic(fmt.Sprintf(
			"Error startSlice %d, gapStart %d, gapEnd %d, endSlice %d\n",
			startSlice, b.gapStart,
			b.gapEnd, endSlice,
		))
	}

	resBuf := make(
		[]byte, 0,
		len(b.content[startSlice:b.gapStart])+len(b.content[b.gapEnd:endSlice]),
	)

	resBuf = append(resBuf, b.content[startSlice:b.gapStart]...)
	resBuf = append(resBuf, b.content[b.gapEnd:endSlice]...)

	if b.markActive {
		mark.Active = true
	}

	return string(resBuf), totalRows, cursor, mark
}

func (b *Buffer) Insert(str string, withUndo bool) {
	if b.markActive {
		b.deleteToMark()
	}

	if b.gapEnd-b.gapStart < GAP_THRESHOLD {
		b.resizeGap()
	}
	for i, ch := range []byte(str) {
		b.content[b.gapStart+i] = ch
	}
	forceNewEvent := false
	if str == "\n" {
		forceNewEvent = true
	}
	if withUndo {
		b.undo.EmitEvent(INSERT_EVENT, b.gapStart, "", forceNewEvent)
	}
	b.gapStart = b.gapStart + len(str)
	b.updateLinePosMem()
}

func (b *Buffer) deleteToMark() {
	gapLen := b.gapEnd - b.gapStart
	if b.gapStart > b.markPos {
		for i := b.gapStart - 1; i >= b.markPos; i-- {
			b.gapStart -= 1
			b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[b.gapStart]), false)
		}
	} else {
		for i := b.gapEnd; i < b.markPos+gapLen; i++ {
			b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[b.gapEnd]), false)
			b.gapEnd += 1
		}
	}
	b.ToggleMark()
	b.updateLinePosMem()
}

func (b *Buffer) DeleteBefore() {
	if b.markActive {
		b.deleteToMark()
		return
	}
	if b.gapStart > 0 {
		b.undo.EmitEvent(DELETE_EVENT, b.gapStart-1, string(b.content[b.gapStart-1]), false)
		b.gapStart -= 1
		b.updateLinePosMem()
	}
}

func (b *Buffer) DeleteAfter(withUndo bool) {
	if b.markActive {
		b.deleteToMark()
		return
	}
	if b.gapEnd < len(b.content) {
		if withUndo {
			b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[b.gapEnd]), false)
		}
		b.gapEnd += 1
		b.updateLinePosMem()
	}
}

func (b *Buffer) DeleteWordBefore() {
	if b.gapStart == 0 {
		return
	}
	if b.markActive {
		b.deleteToMark()
		return
	}
	if b.content[b.gapStart-1] == '\n' {
		b.gapStart -= 1
		b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[b.gapStart]), false)
		return
	} else if utils.IsWhitespace(b.content[b.gapStart-1]) {
		for i := b.gapStart - 1; i >= -1; i-- {
			if i < 0 || b.content[i] == '\n' {
				break
			} else {
				if utils.IsWhitespace(b.content[i]) {
					b.gapStart -= 1
					b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[b.gapStart]), false)
				} else {
					break
				}
			}
		}
	} else {
		for i := b.gapStart - 1; i >= -1; i-- {
			if i < 0 || b.content[i] == '\n' {
				break
			} else {
				if !utils.IsWhitespace(b.content[i]) {
					b.gapStart -= 1
					b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[b.gapStart]), false)
				} else {
					break
				}
			}
		}

	}
	b.updateLinePosMem()
}

func (b *Buffer) DeleteToEnd() {
	if b.markActive {
		b.ToggleMark()
	}
	b.killBuffer = b.killBuffer[0:0]
	if b.gapEnd < len(b.content) {
		for i := b.gapEnd; i <= len(b.content); i++ {
			if i == len(b.content) || b.content[i] == '\n' {
				break
			} else {
				b.killBuffer = append(b.killBuffer, b.content[i])
				b.gapEnd += 1
				b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[i]), false)
			}
		}
	}
}

func (b *Buffer) Copy() {
	b.killBuffer = b.killBuffer[0:0]
	if !b.markActive {
		return
	}
	if b.gapStart > b.markPos {
		for i := b.markPos; i < b.gapStart; i++ {
			b.killBuffer = append(b.killBuffer, b.content[i])
		}
	} else {
		gapLen := b.gapEnd - b.gapStart
		for i := b.gapEnd; i < b.markPos+gapLen; i++ {
			b.killBuffer = append(b.killBuffer, b.content[i])
		}
	}
	b.ToggleMark()
	b.parent.Minibuffer.SetMessage("Copied region")
}

func (b *Buffer) Cut() {
	b.killBuffer = b.killBuffer[0:0]
	if !b.markActive {
		return
	}

	if b.gapStart > b.markPos {
		for i := b.gapStart - 1; i >= b.markPos; i-- {
			b.killBuffer = append(b.killBuffer, b.content[i])
			b.gapStart -= 1
			b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[i]), false)
		}
		for i, j := 0, len(b.killBuffer)-1; i < j; i, j = i+1, j-1 {
			b.killBuffer[i], b.killBuffer[j] = b.killBuffer[j], b.killBuffer[i]
		}
	} else {
		gapLen := b.gapEnd - b.gapStart
		for i := b.gapEnd; i < b.markPos+gapLen; i++ {
			b.killBuffer = append(b.killBuffer, b.content[i])
			b.gapEnd += 1
			b.undo.EmitEvent(DELETE_EVENT, b.gapStart, string(b.content[i]), false)
		}
	}
	b.ToggleMark()
	b.parent.Minibuffer.SetMessage("Cut region")
}

func (b *Buffer) Yank() {
	if b.markActive {
		b.ToggleMark()
		return
	}
	for _, ch := range b.killBuffer {
		b.Insert(string(ch), true)
	}
}

func (b *Buffer) IsMarkActive() bool {
	return b.markActive
}

func (b *Buffer) ToggleMark() {
	b.markActive = !b.markActive
	b.markPos = b.gapStart
	if b.markActive {
		b.parent.Minibuffer.SetMessage("Mark set")
	} else {
		b.parent.Minibuffer.SetMessage("Quit")
	}
}

func (b *Buffer) resizeGap() {
	newBuf := make([]byte, 0, len(b.content)+GAP_LEN)
	newBuf = append(newBuf, b.content[:b.gapEnd]...)
	newBuf = append(newBuf, bytes.Repeat([]byte(" "), GAP_LEN)...)
	newBuf = append(newBuf, b.content[b.gapEnd:]...)
	b.content = newBuf
	b.gapEnd = b.gapEnd + GAP_LEN
}

func (b *Buffer) shiftGapLeft(count int) {
	for i := 0; i < count; i++ {
		if b.gapStart > 0 {
			b.content[b.gapEnd-1] = b.content[b.gapStart-1]
			b.gapStart -= 1
			b.gapEnd -= 1
		} else {
			break
		}
	}
}

func (b *Buffer) shiftGapRight(count int) {
	for i := 0; i < count; i++ {
		if b.gapEnd < len(b.content) {
			b.content[b.gapStart] = b.content[b.gapEnd]
			b.gapStart += 1
			b.gapEnd += 1
		} else {
			break
		}
	}
}

func (b *Buffer) updateLinePosMem() {
	pos := 0
	for i := b.gapStart - 1; i >= -1; i-- {
		if i == -1 || b.content[i] == '\n' {
			b.linePosMem = pos
			break
		} else {
			pos += 1
		}
	}
}

// cursor will always have the same pos as gapStart
func (b *Buffer) MoveForward() {
	b.shiftGapRight(1)
	b.updateLinePosMem()
}

func (b *Buffer) MoveBack() {
	b.shiftGapLeft(1)
	b.updateLinePosMem()
}

func (b *Buffer) MoveUp() {
	found := false
	col := 0
	prevLineLen := 0
	for i := b.gapStart - 1; i >= -1; i-- {
		if i == -1 || b.content[i] == '\n' {
			if !found {
				found = true
			} else {
				if b.linePosMem >= prevLineLen {
					b.shiftGapLeft(col + 1)
				} else {
					b.shiftGapLeft(col + 1)
					b.shiftGapLeft(prevLineLen - b.linePosMem)
				}
				break
			}
		} else {
			if found {
				prevLineLen += 1
			} else {
				col += 1
			}
		}
	}
}

func (b *Buffer) MoveDown() {
	found := false
	remaining := 0
	nextLineLen := 0

	for i := b.gapEnd; i <= len(b.content); i++ {
		if i == len(b.content) || b.content[i] == '\n' {
			if !found {
				found = true
			} else {
				if b.linePosMem >= nextLineLen {
					b.shiftGapRight(remaining)
					b.shiftGapRight(nextLineLen + 1)
				} else {
					b.shiftGapRight(remaining)
					b.shiftGapRight(b.linePosMem + 1)
				}
				break
			}
		} else {
			if found {
				nextLineLen += 1
			} else {
				remaining += 1
			}
		}
	}
}

func (b *Buffer) MoveEndLine() {
	remaining := 0
	for i := b.gapEnd; i <= len(b.content); i++ {
		if i == len(b.content) || b.content[i] == '\n' {
			b.shiftGapRight(remaining)
			b.updateLinePosMem()
			break
		} else {
			remaining += 1
		}
	}
}

func (b *Buffer) MoveStartLine() {
	untilNewline := 0
	untilFirstLeft := 0
	for i := b.gapStart - 1; i >= 0; i-- {
		if b.content[i] == '\n' {
			break
		} else if !utils.IsWhitespace(b.content[i]) {
			untilNewline += 1
			untilFirstLeft = untilNewline
		} else {
			untilNewline += 1
		}
	}

	untilFirstRight := 0
	for i := b.gapEnd; i < len(b.content); i++ {
		if !utils.IsWhitespace(b.content[i]) || b.content[i] == '\n' {
			break
		} else {
			untilFirstRight += 1
		}
	}

	if untilNewline > 0 && untilFirstLeft == 0 && untilFirstRight == 0 {
		b.shiftGapLeft(untilNewline)
		b.linePosMem = 0
	} else if untilFirstLeft > 0 {
		b.shiftGapLeft(untilFirstLeft)
		b.updateLinePosMem()
	} else if untilFirstLeft == 0 && untilFirstRight > 0 {
		b.shiftGapRight(untilFirstRight)
		b.updateLinePosMem()
	}
}

func (b *Buffer) MoveForwardWord() {
	if b.gapEnd == len(b.content) {
		return
	}
	if !utils.IsWhitespace(b.content[b.gapEnd]) {
		for i := b.gapEnd + 1; i <= len(b.content); i++ {
			if i == len(b.content) || b.content[i] == '\n' {
				b.shiftGapRight(i - b.gapEnd)
				b.updateLinePosMem()
				return
			} else {
				if utils.IsDelimiter(b.content[i]) {
					b.shiftGapRight(i - b.gapEnd)
					b.updateLinePosMem()
					return
				}
			}
		}
	} else {
		for i := b.gapEnd + 1; i <= len(b.content); i++ {
			if i == len(b.content) || b.content[i] == '\n' {
				b.shiftGapRight(i - b.gapEnd)
				b.updateLinePosMem()
				return
			} else {
				if !utils.IsWhitespace(b.content[i]) {
					b.shiftGapRight(i - b.gapEnd)
					b.updateLinePosMem()
					return
				}
			}
		}
	}
}

func (b *Buffer) MoveBackWord() {
	if b.gapStart == 0 {
		return
	}
	if !utils.IsWhitespace(b.content[b.gapStart-1]) {
		for i := b.gapStart - 2; i >= -1; i-- {
			if i < 0 || b.content[i] == '\n' {
				b.shiftGapLeft(b.gapStart - i - 1)
				b.updateLinePosMem()
				return
			} else {
				if utils.IsDelimiter(b.content[i]) {
					b.shiftGapLeft(b.gapStart - i - 1)
					b.updateLinePosMem()
					return
				}
			}
		}
	} else {
		for i := b.gapStart - 2; i >= -1; i-- {
			if i < 0 || b.content[i] == '\n' {
				b.shiftGapLeft(b.gapStart - i - 1)
				b.updateLinePosMem()
				return
			} else {
				if !utils.IsWhitespace(b.content[i]) {
					b.shiftGapLeft(b.gapStart - i - 1)
					b.updateLinePosMem()
					return
				}
			}
		}
	}
}

func (b *Buffer) MoveStartFile() {
	b.shiftGapLeft(b.gapStart)
	b.linePosMem = 0
}

func (b *Buffer) MoveEndFile() {
	b.shiftGapRight(len(b.content) - b.gapEnd)
	b.updateLinePosMem()
}

func (b *Buffer) Undo() {
	if b.markActive {
		b.ToggleMark()
	}
	ev, err := b.undo.PopUndoEvent()
	if err != nil {
		b.parent.Minibuffer.SetMessage(err.Error())
		return
	} else {
		switch ev.Type {
		case INSERT_EVENT:
			if b.gapStart > ev.Pos {
				b.shiftGapLeft(b.gapStart - ev.Pos)
				for i := 0; i < ev.NumChar; i++ {
					b.DeleteAfter(false)
				}
			} else {
				b.shiftGapRight(ev.Pos - b.gapStart)
				for i := 0; i < ev.NumChar; i++ {
					b.DeleteAfter(false)
				}
			}
		case DELETE_EVENT:
			if b.gapStart > ev.Pos {
				b.shiftGapLeft(b.gapStart - ev.Pos)
				b.Insert(ev.StoredText, false)
			} else {
				b.shiftGapRight(ev.Pos - b.gapStart)
				b.Insert(ev.StoredText, false)
			}
		}
	}
}
