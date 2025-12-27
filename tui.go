package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cursorPos   [][]int
	currentFile FileDetails
	mode        string
}

func initialModel(workingFile FileDetails) model {
	return model{
		cursorPos:   [][]int{{0, 0}},
		currentFile: workingFile,
		mode:        "VIEW",
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
				if m.cursorPos[0][0] >= 1 {
					m.cursorPos[0][0]--
				}

			case "j":
				if m.cursorPos[0][1] <= m.currentFile.Metadata.length-2 {
					m.cursorPos[0][1]++
				}

			case "k":
				if m.cursorPos[0][1] >= 1 {
					m.cursorPos[0][1]--
				}

			case "l":
				lines := strings.Split(m.currentFile.Content, "\n")
				lineLength := len(lines[m.cursorPos[0][1]])
				if m.cursorPos[0][0] < lineLength-1 {
					m.cursorPos[0][0]++
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

	// feat add in cursor position
	footer := fmt.Sprintf("\n -- MODE: %s -- | %s | %v lines | %s | %s | %s | %s | %s |",
		m.mode, metadata.name, metadata.length, metadata.size, metadata.encodingType, metadata.newlineType, metadata.extensionType, metadata.modifiedDate)

	return fmt.Sprintf("%s %s", content, footer)
}

func loadTUI(workingFile FileDetails) {
	p := tea.NewProgram(initialModel(workingFile))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

}
