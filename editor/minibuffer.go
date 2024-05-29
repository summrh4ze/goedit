package editor

import "org.example.goedit/utils"

type Minibuffer struct {
	message string
	input   string
	col     int
	ready   chan<- bool
	Focused bool
	Dirty   bool
}

func NewMinibuffer(ready chan<- bool) *Minibuffer {
	return &Minibuffer{
		ready: ready,
	}
}

func (m *Minibuffer) ConfirmAction() {
	m.ready <- true
}

func (m *Minibuffer) RejectAction() {
	m.ConsumeInput()
	m.ready <- false
}

func (m *Minibuffer) SetMessage(msg string) {
	m.message = msg
	m.Dirty = true
}

func (m *Minibuffer) GetLine() string {
	m.Dirty = false
	return m.message + m.input
}

func (m *Minibuffer) ConsumeInput() string {
	res := m.input
	m.input = ""
	m.col = 0
	return res
}

func (m *Minibuffer) GetCursor() int {
	return len(m.message) + m.col
}

func (m *Minibuffer) InsertAtCol(str string) {
	if m.col >= 0 && m.col <= len(m.input) {
		m.input = m.input[0:m.col] + str + m.input[m.col:]
		m.col += 1
	}
}

func (m *Minibuffer) DeleteAtCol() {
	if m.col > 0 && m.col <= len(m.input) {
		m.input = m.input[0:m.col-1] + m.input[m.col:]
		m.col -= 1
	}
}

func (m *Minibuffer) MoveForward() {
	if m.col < len(m.input) {
		m.col += 1
	}
}

func (m *Minibuffer) MoveBack() {
	if m.col > 0 {
		m.col -= 1
	}
}

func (m *Minibuffer) MoveEndLine() {
	m.col = len(m.input)
}

func (m *Minibuffer) MoveStartLine() {
	m.col = 0
}

func (m *Minibuffer) MoveForwardWord() {
	if m.col < len(m.input) {
		col := m.col
		if m.input[m.col] != ' ' {
			for i := m.col + 1; i < len(m.input); i++ {
				if utils.IsDelimiter(m.input[i]) {
					col = i
					break
				}
			}
			if col > m.col {
				m.col = col
			} else {
				// word goes until the end of the line so stop at the very end
				m.col = len(m.input)
			}
		} else {
			// search for the first non whitespace character
			for i := m.col + 1; i < len(m.input); i++ {
				if !utils.IsDelimiter(m.input[i]) {
					col = i
					break
				}
			}
			if col > m.col {
				m.col = col
			}
		}
	}
}
