package main

type Cursor struct {
	Row int
	Col int
}

type Buffer struct {
	Lines  []string
	Cursor Cursor
}

type EditorAction struct {
	Name     string
	Callback func()
}

type Editor struct {
	Closed        bool
	Keybindings   map[string]*EditorAction
	OpenBuffers   []Buffer
	CurrentBuffer int
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
}

func CreateEditor() *Editor {
	editor := &Editor{
		Closed:        false,
		OpenBuffers:   make([]Buffer, 0, 1),
		Keybindings:   make(map[string]*EditorAction),
		CurrentBuffer: -1,
	}
	initKeybindings(editor)
	return editor
}

func (e *Editor) handleInput(keybinding string) {
	if editorAction, ok := e.Keybindings[keybinding]; ok {
		editorAction.Callback()
	}
}
