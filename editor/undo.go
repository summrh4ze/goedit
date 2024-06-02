package editor

import (
	"errors"
	"fmt"
)

const (
	INSERT_EVENT    = 0 // marks insertion of text
	DELETE_EVENT    = 1 // marks deletion of text
	UNCHANGED_EVENT = 2 // marks moment buffer is saved
)

type UndoEvent struct {
	Prev       *UndoEvent // prev event in list
	Next       *UndoEvent // next event in list
	Type       int        // type of undo event
	Pos        int        // anchor position where the event took place
	NumChar    int        // number of characters
	StoredText string     // storage for deleted text
}

type UndoStack struct {
	head        *UndoEvent
	tail        *UndoEvent
	currentSize int
	size        int
}

func (ue *UndoEvent) String() string {
	return fmt.Sprintf("Type: %d, Anchor: %v, NumChar: %d, Text: %s\n", ue.Type, ue.Pos, ue.NumChar, ue.StoredText)
}

func (u *UndoStack) String() string {
	res := "["
	temp := u.head
	if temp == nil {
		return fmt.Sprintf("[]")
	}
	for temp.Next != nil {
		res += fmt.Sprintf("%v, ", temp)
		temp = temp.Next
	}
	res += fmt.Sprintf("%v]", temp)
	return res
}

func NewUndo(size int) *UndoStack {
	return &UndoStack{
		head:        nil,
		tail:        nil,
		currentSize: 0,
		size:        size,
	}
}

func (u *UndoStack) EmitEvent(t int, pos int, text string, forceNew bool) {
	if u.currentSize < u.size {
		if u.currentSize > 0 && u.head != nil && !forceNew {
			// check if the last event was insert and if it was local
			if u.head.Type == INSERT_EVENT && t == INSERT_EVENT && u.head.Pos+u.head.NumChar == pos {
				u.head.NumChar += 1
				return
			} else if u.head.Type == DELETE_EVENT && t == DELETE_EVENT {
				if u.head.Pos == pos {
					u.head.NumChar += 1
					u.head.StoredText = u.head.StoredText + text
					return
				} else if u.head.Pos == pos+1 {
					u.head.NumChar += 1
					u.head.Pos = pos
					u.head.StoredText = text + u.head.StoredText
					return
				}
			}
		}
		newEvent := &UndoEvent{
			Next:       u.head,
			Type:       t,
			Pos:        pos,
			NumChar:    1,
			StoredText: text,
		}
		if u.head != nil {
			u.head.Prev = newEvent
		} else {
			u.tail = newEvent
		}
		u.head = newEvent
		u.currentSize += 1
	} else {
		// size = max, discard tail and insert at head
		u.tail.Prev.Next = nil
		u.tail.Prev = nil
		newEvent := &UndoEvent{
			Next:       u.head,
			Type:       t,
			Pos:        pos,
			NumChar:    1,
			StoredText: text,
		}
		u.head.Prev = newEvent
		u.head = newEvent
		u.currentSize += 1
	}
}

func (u *UndoStack) PopUndoEvent() (*UndoEvent, error) {
	if u.currentSize == 0 {
		return nil, errors.New("Undo stack is empty")
	} else {
		res := u.head
		if u.head.Next != nil {
			u.head.Next.Prev = nil
			u.head = u.head.Next
		} else {
			u.head = nil
			u.tail = nil
		}

		res.Next = nil
		u.currentSize -= 1
		return res, nil
	}
}
