package editor

import "org.example.goedit/utils"

type Cursor struct {
	Row int
	Col int
	Mem int
}

type Buffer struct {
	Name         string
	Content      []string
	Cursor       Cursor
	ReadOnlyMode bool
	BaseRow      int
}

func NewEmptyBuffer() *Buffer {
	return &Buffer{
		Name:         "scratch",
		Content:      []string{"\tGOEdit!", "To open a file use Ctrl-X Ctrl-F"},
		ReadOnlyMode: false,
		Cursor:       Cursor{0, 0, 0},
	}
}

func (b *Buffer) GetLines(count int) []string {
	if b.BaseRow < len(b.Content) && b.BaseRow+count <= len(b.Content) {
		return b.Content[b.BaseRow : b.BaseRow+count]
	} else if b.BaseRow < len(b.Content) && b.BaseRow+count > len(b.Content) {
		return b.Content[b.BaseRow:len(b.Content)]
	}
	return []string{}
}

func (b *Buffer) Insert(str string) {

}

func (b *Buffer) MoveForward() {
	var line string
	if len(b.Content) > b.Cursor.Row && b.Cursor.Row >= 0 {
		line = b.Content[b.Cursor.Row]
	}
	if b.Cursor.Col < utils.Tlen(line, TABSIZE) {
		b.Cursor = Cursor{b.Cursor.Row, b.Cursor.Col + 1, b.Cursor.Col + 1}
	} else if utils.Tlen(line, TABSIZE) == 0 || b.Cursor.Col == utils.Tlen(line, TABSIZE) {
		if b.Cursor.Row+1 < len(b.Content) {
			b.Cursor = Cursor{b.Cursor.Row + 1, 0, 0}
		}
	}
}

func (b *Buffer) MoveBack() {
	var line string
	if len(b.Content) > b.Cursor.Row-1 && b.Cursor.Row-1 >= 0 {
		line = b.Content[b.Cursor.Row-1]
	}
	if b.Cursor.Col > 0 {
		b.Cursor = Cursor{b.Cursor.Row, b.Cursor.Col - 1, b.Cursor.Col - 1}
	} else if b.Cursor.Col == 0 && b.Cursor.Row-1 >= 0 {
		b.Cursor = Cursor{b.Cursor.Row - 1, utils.Tlen(line, TABSIZE), utils.Tlen(line, TABSIZE)}
	}
}

func (b *Buffer) MoveUp() {
	if b.Cursor.Row > 0 {
		prevLine := b.Content[b.Cursor.Row-1]
		if b.Cursor.Col > utils.Tlen(prevLine, TABSIZE) || b.Cursor.Mem > utils.Tlen(prevLine, TABSIZE) {
			b.Cursor = Cursor{b.Cursor.Row - 1, utils.Tlen(prevLine, TABSIZE), b.Cursor.Mem}
		} else if b.Cursor.Mem < b.Cursor.Col {
			b.Cursor = Cursor{b.Cursor.Row - 1, b.Cursor.Col, b.Cursor.Mem}
		} else {
			b.Cursor = Cursor{b.Cursor.Row - 1, b.Cursor.Mem, b.Cursor.Mem}
		}
	}
}

func (b *Buffer) MoveDown() {
	if b.Cursor.Row < len(b.Content)-1 {
		nextLine := b.Content[b.Cursor.Row+1]
		if b.Cursor.Col > utils.Tlen(nextLine, TABSIZE) || b.Cursor.Mem > utils.Tlen(nextLine, TABSIZE) {
			b.Cursor = Cursor{b.Cursor.Row + 1, utils.Tlen(nextLine, TABSIZE), b.Cursor.Mem}
		} else if b.Cursor.Mem < b.Cursor.Col {
			b.Cursor = Cursor{b.Cursor.Row + 1, b.Cursor.Col, b.Cursor.Mem}
		} else {
			b.Cursor = Cursor{b.Cursor.Row + 1, b.Cursor.Mem, b.Cursor.Mem}
		}
	}
}

func (b *Buffer) MoveEndLine() {
	var line string
	if len(b.Content) > b.Cursor.Row && b.Cursor.Row >= 0 {
		line = b.Content[b.Cursor.Row]
	}
	b.Cursor = Cursor{b.Cursor.Row, utils.Tlen(line, TABSIZE), utils.Tlen(line, TABSIZE)}
}

func (b *Buffer) MoveStartLine() {
	var line string
	if len(b.Content) > b.Cursor.Row && b.Cursor.Row >= 0 {
		line = b.Content[b.Cursor.Row]
	}
	expLine := utils.Texp(line, TABSIZE)
	firstNonWh := 0
	for i := 0; i < len(expLine); i++ {
		if expLine[i] != ' ' {
			firstNonWh = i
			break
		}
	}
	if b.Cursor.Col == 0 || b.Cursor.Col > firstNonWh {
		b.Cursor = Cursor{b.Cursor.Row, firstNonWh, firstNonWh}
	} else {
		b.Cursor = Cursor{b.Cursor.Row, 0, 0}
	}
}

func (b *Buffer) MoveForwardWord() {
	var line string
	if len(b.Content) > b.Cursor.Row && b.Cursor.Row >= 0 {
		line = b.Content[b.Cursor.Row]
	}

	// need to expand the string before checking characters
	if b.Cursor.Col < utils.Tlen(line, TABSIZE) {
		expLine := utils.Texp(line, TABSIZE)
		row := b.Cursor.Row
		col := b.Cursor.Col
		if expLine[b.Cursor.Col] != ' ' {
			for i := b.Cursor.Col + 1; i < len(expLine); i++ {
				if utils.IsDelimiter(expLine[i]) {
					col = i
					break
				}
			}
			if col > b.Cursor.Col {
				b.Cursor = Cursor{b.Cursor.Row, col, col}
			} else {
				// word goes until the end of the line so stop at the very end
				b.Cursor = Cursor{b.Cursor.Row, utils.Tlen(line, TABSIZE), utils.Tlen(line, TABSIZE)}
			}
		} else {
			// search for the first non whitespace character
			for i := b.Cursor.Col + 1; i < len(expLine); i++ {
				if !utils.IsDelimiter(expLine[i]) {
					col = i
					break
				}
			}
			if col > b.Cursor.Col {
				b.Cursor = Cursor{b.Cursor.Row, col, col}
			} else if b.Cursor.Row+1 < len(b.Content) {
				// not found on current line search the next line
				row += 1
				col = 0
				nextLine := b.Content[b.Cursor.Row+1]
				nexpLine := utils.Texp(nextLine, TABSIZE)
				for i := 0; i < len(nexpLine); i++ {
					if !utils.IsDelimiter(nexpLine[i]) {
						col = i
						break
					}
				}
				b.Cursor = Cursor{row, col, col}
			}
		}
	} else if utils.Tlen(line, TABSIZE) == 0 || b.Cursor.Col == utils.Tlen(line, TABSIZE) {
		if b.Cursor.Row+1 < len(b.Content) {
			nexpLine := utils.Texp(b.Content[b.Cursor.Row+1], TABSIZE)
			col := 0
			for i := 0; i < len(nexpLine); i++ {
				if !utils.IsDelimiter(nexpLine[i]) {
					col = i
					break
				}
			}
			b.Cursor = Cursor{b.Cursor.Row + 1, col, col}
		}
	}
}
