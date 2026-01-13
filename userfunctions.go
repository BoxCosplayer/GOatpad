package main

import (
	"bufio"
	"fmt"
	"os"
)

func write_file(filename string, fileExtension string) {
	// Create or open the 'filename.extension'
	file, err := os.Create(filename + fileExtension)
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	// Write each line to the file manually
	// by ensuring to add newlines
	writer := bufio.NewWriter(file)
	for row, line := range textBuffer {
		newLine := "\n"

		if row == len(textBuffer)-1 {
			newLine = ""
		}

		_, err = writer.WriteString(string(line) + newLine)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		writer.Flush()
		modified = false
	}

}

func insert_line() {

	// Create a new line, and wrap the remainder of the current line onto the next one
	newLine := make([]rune, len(textBuffer[currentRow])-currentCol)
	copy(newLine, textBuffer[currentRow][currentCol:])

	textBuffer[currentRow] = textBuffer[currentRow][:currentCol]
	textBuffer = append(textBuffer[:currentRow+1], append([][]rune{newLine}, textBuffer[currentRow+1:]...)...)

	currentRow++
	currentCol = 0

	mark_viewport_dirty()
	mark_line_dirty(currentRow)
}

// ---------- Symbol Copying ----------

func copy_symbol() {
	// Find the first and last character of a symbol,
	// which is detected using non Alphanumeric chars
	currentLine := textBuffer[currentRow]
	left, right := get_symbol_from_line(currentLine, currentCol)
	symbol := currentLine[left:right]

	symbolCopy := make([]rune, len(symbol))
	copy(symbolCopy, symbol)
	copyBuffer.contents = [][]rune{symbolCopy}
	copyBuffer.bufferType = "symbol"
}

func cut_symbol() {
	copy_symbol()
	delete_symbol()
}

func paste_symbol() {
	if len(copyBuffer.contents[0]) != 0 && copyBuffer.bufferType == "symbol" {
		symbolLength := len(copyBuffer.contents[0])

		newLine := make([]rune, len(textBuffer[currentRow])+symbolLength)

		copy(newLine[:currentCol], textBuffer[currentRow][:currentCol])
		copy(newLine[currentCol:currentCol+symbolLength], copyBuffer.contents[0])
		copy(newLine[currentCol+symbolLength:], textBuffer[currentRow][currentCol:])

		textBuffer[currentRow] = newLine
		currentCol += len(copyBuffer.contents[0])
		mark_line_dirty(currentRow)
	}
}

func delete_symbol() {
	currentLine := textBuffer[currentRow]
	left, right := get_symbol_from_line(currentLine, currentCol)

	textBuffer[currentRow] = append(textBuffer[currentRow][:left], textBuffer[currentRow][right:]...)
	mark_line_dirty(currentRow)
}

// TODO: rename symbol

// ---------- Line Copying ----------

func copy_line() {
	copyLine := make([]rune, len(textBuffer[currentRow]))
	copy(copyLine, textBuffer[currentRow])
	copyBuffer.contents[0] = copyLine
	copyBuffer.bufferType = "line"
}

func cut_line() {
	if (currentRow >= len(textBuffer)) == false {
		copy_line()
		delete_line()
	}
}

func paste_line() {
	if len(copyBuffer.contents[0]) != 0 && copyBuffer.bufferType == "line" {
		// move the data from the copy buffer into a newline **below** the current line
		// append to the text buffer
		newLine := make([]rune, len(copyBuffer.contents[0]))
		copy(newLine, copyBuffer.contents[0])
		textBuffer = append(textBuffer[:currentRow+1], append([][]rune{newLine}, textBuffer[currentRow+1:]...)...)

		currentRow++
		currentCol = 0
		mark_viewport_dirty()
		mark_line_dirty(currentRow)
	}
}

func delete_line() {
	textBuffer = append(textBuffer[:currentRow], textBuffer[currentRow+1:]...)
	mark_viewport_dirty()
	mark_line_dirty(currentRow)
}

// ---------- Block Copying ----------

var (
	blockCounter = 0
	prevRow      = 0
	prevCol      = 0
)

func copy_block() {
	// Cycle through blocks so that you can "choose" the scope
	// if cursor is in the middle of a block

	// This only works by using the same keybind whilst
	// cursorPos doesnt change
	if currentRow != prevRow || currentCol != prevCol {
		blockCounter = 0
	}

	prevRow = currentRow
	prevCol = currentCol

	// Find the first and last line of a block,
	// which is where the curly braces are located
	left, right := find_current_block(blockCounter)
	copyBuffer.contents = make([][]rune, right-left+1)

	for i := left; i <= right; i++ {
		copyLine := make([]rune, len(textBuffer[i]))
		copy(copyLine, textBuffer[i])
		copyBuffer.contents[i-left] = copyLine
	}
	copyBuffer.bufferType = "block"
	blockCounter++
}

func cut_block() {
	copy_block()
	delete_block()
}

// for line in copy buffer, paste_line()
func paste_block() {
	if len(copyBuffer.contents[0]) != 0 && copyBuffer.bufferType == "block" {
		// for line in copyBuffer (which is of type [][]rune), paste_line()
		for _, line := range copyBuffer.contents {
			newLine := make([]rune, len(line))
			copy(newLine, line)
			textBuffer = append(textBuffer[:currentRow+1], append([][]rune{newLine}, textBuffer[currentRow+1:]...)...)
			currentRow++
		}

		currentCol = 0
		mark_viewport_dirty()
		mark_line_dirty(currentRow)
	}
}

func delete_block() {
	// Like the above, this has the same function as delete_line()
	// if there are no blocks selected, needs same safeguards
	if len(textBuffer) > 1 && currentRow != len(textBuffer)-1 {
		left, right := find_current_block(0)
		textBuffer = append(textBuffer[:left], textBuffer[right+1:]...)
		mark_viewport_dirty()
		mark_line_dirty(currentRow)
	}
}

// ---------- State Saving ----------

func push_state() {
	// Push the current textBuffer onto a stack
	stateCopy := make([][]rune, len(textBuffer))
	for i := range textBuffer {
		stateCopy[i] = make([]rune, len(textBuffer[i]))
		copy(stateCopy[i], textBuffer[i])
	}
	undoStack.push(stateCopy)
}

func pull_state() {
	// Pull from the top of the stack, replace the textBuffer with it
	if len(undoStack.contents) > 0 {
		textBuffer = undoStack.pop().([][]rune)
		mark_viewport_dirty()
		mark_line_dirty(currentRow)
	}
}
