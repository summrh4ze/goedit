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
		if m.input[m.col] != ' ' {
			for i := m.col + 1; i <= len(m.input); i++ {
				if i == len(m.input) {
					m.col = len(m.input)
					break
				}
				if utils.IsDelimiter(m.input[i]) {
					m.col = i
					break
				}
			}
		} else {
			for i := m.col + 1; i <= len(m.input); i++ {
				if i == len(m.input) {
					m.col = len(m.input)
					break
				}
				if !utils.IsWhitespace(m.input[i]) {
					m.col = i
					break
				}
			}
		}
	}
}

func (m *Minibuffer) MoveBackWord() {
	if m.col > 0 {
		if m.input[m.col-1] != ' ' {
			for i := m.col - 1; i >= -1; i-- {
				if i < 0 {
					m.col = 0
					break
				}
				if utils.IsDelimiter(m.input[i]) {
					m.col = i + 1
					break
				}
			}
		} else {
			for i := m.col - 1; i >= -1; i-- {
				if i < 0 {
					m.col = 0
					break
				}
				if !utils.IsWhitespace(m.input[i]) {
					m.col = i + 1
					break
				}
			}
		}
	}
}
