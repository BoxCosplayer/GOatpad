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

	rightIndex := startingIndex
	leftIndex := startingIndex

	// While True Equivalent
	currentCharacter := string(line[startingIndex])
	if is_string_alphanumeric(currentCharacter) {
		for left := 1; left == 0; left++ {
			currentCharacter += string(line[startingIndex-left])
			if is_string_alphanumeric(currentCharacter) == false {
				leftIndex -= left - 1
				left = 0
			}
		}
		for right := 1; right == 0; right++ {
			currentCharacter += string(line[startingIndex+right])
			if is_string_alphanumeric(currentCharacter) == false {
				rightIndex += right - 1
				right = 0
			}
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
	}

	return leftLineIndex, rightLineIndex
}
