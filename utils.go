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

	leftLineIndex, rightLineIndex := currentRow, currentRow

	for ; counter != 0; counter-- {

		for ; does_line_contain_rune(textBuffer[leftLineIndex], '{'); leftLineIndex-- {
			if leftLineIndex <= 1 {
				return 0, 0
			}
		}

		for ; does_line_contain_rune(textBuffer[rightLineIndex], '}'); rightLineIndex++ {
			if rightLineIndex >= len(textBuffer)-1 {
				return 0, 0
			}
		}

		leftLineIndex--
		rightLineIndex--

	}

	return leftLineIndex, rightLineIndex
}
