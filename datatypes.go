package main

import (
	termbox "github.com/nsf/termbox-go"
)

type stack struct {
	// array of any type
	contents []interface{}
}

// methods:

// push, to add an extra value onto the array
func (s *stack) push(value interface{}) {
	s.contents = append(s.contents, value)
}

// pop, to return and remove the last item in the array
func (s *stack) pop() interface{} {
	if len(s.contents) == 0 {
		return nil
	}
	value := s.contents[len(s.contents)-1]
	s.contents = s.contents[:len(s.contents)-1]
	return value
}

type CopyBuffer struct {
	contents   [][]rune
	bufferType string
}

// position tracks any position in the text buffer.
type position struct {
	row int
	col int
}

type rulerState struct {
	enabled bool
	col     int
}

func new_ruler_state() rulerState {
	if RULER_COL <= 0 {
		return rulerState{enabled: false, col: -1}
	}
	return rulerState{enabled: true, col: RULER_COL - 1}
}

func (r rulerState) highlight(textBufferCol int) bool {
	return r.enabled && textBufferCol == r.col
}

func (r rulerState) draw_for_short_line(cursorRow int, lineLen int, textCols int, gutterWidth int, offsetCol int) {
	if !r.enabled || textCols <= 0 || r.col < lineLen {
		return
	}
	if r.col < offsetCol || r.col >= offsetCol+textCols {
		return
	}
	rulerScreenCol := gutterWidth + (r.col - offsetCol)
	if rulerScreenCol < gutterWidth || rulerScreenCol >= COLS {
		return
	}
	termbox.SetCell(rulerScreenCol, cursorRow, ' ', termbox.ColorDefault, RULER_BG)
}
