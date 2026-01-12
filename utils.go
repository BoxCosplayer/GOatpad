package main

import "unicode"

func is_string_alphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func get_symbol_from_line(line []rune, startingIndex int) (int, int) {
	if len(line) == 0 {
		return 0, 0
	}
	if startingIndex < 0 {
		return 0, 0
	}
	if startingIndex >= len(line) {
		return 0, 0
	}

	rightIndex := startingIndex
	leftIndex := startingIndex

	if unicode.IsLetter(line[startingIndex]) || unicode.IsDigit(line[startingIndex]) {
		// Scan left while alphanumeric
		for left := startingIndex - 1; left >= 0; left-- {
			if !unicode.IsLetter(line[left]) && !unicode.IsDigit(line[left]) {
				break
			}
			leftIndex = left
		}
		// Scan right while alphanumeric
		for right := startingIndex + 1; right < len(line); right++ {
			if !unicode.IsLetter(line[right]) && !unicode.IsDigit(line[right]) {
				break
			}
			rightIndex = right + 1
		}
	}

	return leftIndex, rightIndex
}

func does_line_contain_rune(line []rune, searchRune rune) bool {
	for _, r := range line {
		if r == searchRune {
			return true
		}
	}
	return false
}

func find_current_block(counter int) (int, int) {
	if len(textBuffer) == 0 {
		return 0, 0
	}
	if currentRow < 0 {
		return 0, 0
	}
	if currentRow >= len(textBuffer) {
		last := len(textBuffer) - 1
		return last, last
	}

	type bracePos struct {
		row int
		col int
	}
	openStack := make([]bracePos, 0, 8)

	// Track unmatched opening braces up to the current row.
	for row := 0; row <= currentRow; row++ {
		line := textBuffer[row]
		for col, ch := range line {
			switch ch {
			case '{':
				openStack = append(openStack, bracePos{row: row, col: col})
			case '}':
				if len(openStack) > 0 {
					openStack = openStack[:len(openStack)-1]
				}
			}
		}
	}

	targetIndex := len(openStack) - 1 - counter
	if targetIndex < 0 {
		return currentRow, currentRow
	}
	start := openStack[targetIndex]

	depth := 0
	for row := start.row; row < len(textBuffer); row++ {
		line := textBuffer[row]
		colStart := 0
		if row == start.row {
			colStart = start.col + 1
			if colStart > len(line) {
				colStart = len(line)
			}
		}
		for col := colStart; col < len(line); col++ {
			switch line[col] {
			case '{':
				depth++
			case '}':
				if depth == 0 {
					return start.row, row
				}
				depth--
			}
		}
	}

	return start.row, start.row
}
