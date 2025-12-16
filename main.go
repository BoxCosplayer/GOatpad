package main

import (
	"flag"
	"fmt"
)

func main() {
	// Load all cli flags
	var file string
	flag.StringVar(&file, "file", "", "Path to the input file")
	flag.StringVar(&file, "f", "", "Shorthand for -file")

	// feat:
	// folder integration
	// ...

	flag.Parse()

	workingFile := loadFileDetails(file)

	fmt.Println(workingFile.Content)
}
