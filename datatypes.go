package main

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

// bracePos tracks an opening brace position in the text buffer.
type bracePos struct {
	row int
	col int
}
