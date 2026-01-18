package main

import "unicode"

func sync_dirty_rows() {
	if len(dirtyRows) > len(textBuffer) {
		dirtyRows = dirtyRows[:len(textBuffer)]
		return
	}
	if len(dirtyRows) < len(textBuffer) {
		dirtyRows = append(dirtyRows, make([]bool, len(textBuffer)-len(dirtyRows))...)
	}
}

func mark_line_dirty(row int) {
	modified = true
	sync_dirty_rows()
	if row < 0 || row >= len(textBuffer) {
		return
	}
	dirtyRows[row] = true
}

func mark_screen_dirty() {
	mark_line_dirty(currentRow)
}

func mark_viewport_dirty() {
	viewportDirty = true
}

func is_string_alphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func get_symbol_from_line(line []rune, startingIndex int) (int, int) {
	// Symbol is detected using surrounding
	// alphanumeric characters

	// edge cases
	if len(line) == 0 {
		return 0, 0
	}
	if startingIndex < 0 {
		return 0, 0
	}
	if startingIndex >= len(line) {
		return 0, 0
	}

	rightIndex := startingIndex + 1
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

func does_array_contain_rune(line []rune, searchRune rune) bool {
	for _, r := range line {
		if r == searchRune {
			return true
		}
	}
	return false
}

func find_current_block(counter int) (int, int) {
	// Fast-path guardrails for empty buffer or invalid cursor row.
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

	openStack := make([]position, 0, 8)

	// Track unmatched opening braces up to the current row.
	for row := 0; row <= currentRow; row++ {
		line := textBuffer[row]
		for col, ch := range line {
			switch ch {
			case '{':
				openStack = append(openStack, position{row: row, col: col})
			case '}':
				if len(openStack) > 0 {
					openStack = openStack[:len(openStack)-1]
				}
			}
		}
	}

	// Pick the Nth enclosing block (counter=0 is the innermost).
	targetIndex := len(openStack) - 1 - counter
	if targetIndex < 0 {
		return currentRow, currentRow
	}
	start := openStack[targetIndex]

	// Scan forward from the opening brace to find the matching closing brace.
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
