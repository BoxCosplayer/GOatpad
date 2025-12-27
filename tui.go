package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

// runeCol -> byte offset
func byteOffsetAtRuneCol(s string, col int) int {
	if col <= 0 {
		return 0
	}
	off := 0
	for i := 0; i < col && off < len(s); i++ {
		_, size := utf8.DecodeRuneInString(s[off:])
		off += size
	}
	if off > len(s) {
		off = len(s)
	}
	return off
}

// byte offset -> rune column
func runeColAtByteOffset(s string, off int) int {
	if off <= 0 {
		return 0
	}
	if off > len(s) {
		off = len(s)
	}
	col := 0
	i := 0
	for i < off {
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		col++
	}
	return col
}

type model struct {
	cursorPos     [][]int
	cursorTrueCol int
	currentFile   FileDetails
	mode          string
}

func initialModel(workingFile FileDetails) model {
	return model{
		cursorPos:     [][]int{{0, 0}},
		cursorTrueCol: 0,
		currentFile:   workingFile,
		mode:          "VIEW",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case "VIEW":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Keypress confirmed
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit

			case "h":
				y := m.cursorPos[0][1]
				lines := strings.Split(m.currentFile.Content, "\n")
				if y < len(lines) {
					line := lines[y]
					if m.cursorTrueCol == 0 && m.cursorPos[0][0] > 0 {
						m.cursorTrueCol = runeColAtByteOffset(line, m.cursorPos[0][0])
					}
					// cursorPos x is a byte offset; step left by the size of the previous rune
					if m.cursorPos[0][0] > 0 {
						if m.cursorPos[0][0] > len(line) {
							m.cursorPos[0][0] = len(line)
						}
						_, size := utf8.DecodeLastRuneInString(line[:m.cursorPos[0][0]])
						m.cursorPos[0][0] -= size
						if m.cursorPos[0][0] < 0 {
							m.cursorPos[0][0] = 0
						}
						if m.cursorTrueCol > 0 {
							m.cursorTrueCol--
						}
					}
				}

			case "j":
				lines := strings.Split(m.currentFile.Content, "\n")
				y := m.cursorPos[0][1]
				nextY := y + 1
				if y < len(lines) && nextY < len(lines) {
					curCol := m.cursorTrueCol
					if curCol == 0 && m.cursorPos[0][0] > 0 {
						curCol = runeColAtByteOffset(lines[y], m.cursorPos[0][0])
					}
					m.cursorPos[0][1] += 1
					m.cursorPos[0][0] = byteOffsetAtRuneCol(lines[nextY], curCol)
					m.cursorTrueCol = curCol
				}

			case "k":
				lines := strings.Split(m.currentFile.Content, "\n")
				y := m.cursorPos[0][1]
				prevY := y - 1
				if y < len(lines) && prevY >= 0 {
					curCol := m.cursorTrueCol
					if curCol == 0 && m.cursorPos[0][0] > 0 {
						curCol = runeColAtByteOffset(lines[y], m.cursorPos[0][0])
					}
					m.cursorPos[0][1] = prevY
					m.cursorPos[0][0] = byteOffsetAtRuneCol(lines[prevY], curCol)
					m.cursorTrueCol = curCol
				}

			case "l":
				lines := strings.Split(m.currentFile.Content, "\n")
				y := m.cursorPos[0][1]
				if y < len(lines) {
					line := lines[y]
					if m.cursorTrueCol == 0 && m.cursorPos[0][0] > 0 {
						m.cursorTrueCol = runeColAtByteOffset(line, m.cursorPos[0][0])
					}
					lineLength := len(line)
					if m.cursorPos[0][0] < lineLength-1 {
						// cursorPos x is a byte offset; step right by the size of the next rune
						_, size := utf8.DecodeRuneInString(line[m.cursorPos[0][0]:])
						m.cursorPos[0][0] += size
						if m.cursorPos[0][0] > lineLength {
							m.cursorPos[0][0] = lineLength
						}
						m.cursorTrueCol++
					}
				}
			}
		}
	case "EDIT":
		fmt.Println("switch successful!")
	}

	return m, nil
}

func (m model) View() string {
	content := strings.ReplaceAll(m.currentFile.Content, "\r", "")

	// render cursor, based on m.cursorPos
	// [x][y] means yth line, xth character in that line
	// Line y, column x
	if len(m.cursorPos) > 0 && content != "" {
		lines := strings.Split(content, "\n")
		cursorMap := make(map[int][]int, len(m.cursorPos))

		for _, pos := range m.cursorPos {
			if len(pos) < 2 {
				continue
			}
			x, y := pos[0], pos[1]
			if y < 0 || y >= len(lines) {
				continue
			}
			if x < 0 {
				x = 0
			}
			if x > len(lines[y]) {
				x = len(lines[y])
			}
			cursorMap[y] = append(cursorMap[y], x)
		}

		for lineIdx, cols := range cursorMap {
			sort.Ints(cols)
			line := lines[lineIdx]
			var b strings.Builder
			last := 0

			for _, col := range cols {
				if col > len(line) {
					col = len(line)
				}
				b.WriteString(line[last:col])
				b.WriteByte('|')
				last = col
			}

			b.WriteString(line[last:])
			lines[lineIdx] = b.String()
		}

		content = strings.Join(lines, "\n")
	}

	metadata := m.currentFile.Metadata

	footer := fmt.Sprintf("\n -- MODE: %s -- | Ln %d Col %d | %s | %v lines | %s | %s | %s | %s | %s |",
		m.mode, m.cursorPos[0][1], m.cursorTrueCol, metadata.name, metadata.length, metadata.size,
		metadata.encodingType, metadata.newlineType, metadata.extensionType, metadata.modifiedDate)

	return fmt.Sprintf("%s %s", content, footer)
}

func loadTUI(workingFile FileDetails) {
	p := tea.NewProgram(initialModel(workingFile))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

}
