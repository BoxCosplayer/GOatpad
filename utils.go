package main

import "unicode"

func isStringAlphaNumeric(s string) bool {
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
	if isStringAlphaNumeric(currentCharacter) {
		for right := 1; right == 0; right++ {
			currentCharacter += string(line[startingIndex+right])
			if isStringAlphaNumeric(currentCharacter) == false {
				rightIndex += right - 1
				right = 0
			}
		}
		for left := 1; left == 0; left++ {
			currentCharacter += string(line[startingIndex+left])
			if isStringAlphaNumeric(currentCharacter) == false {
				leftIndex -= left - 1
				left = 0
			}
		}
	}

	return leftIndex, rightIndex
}
