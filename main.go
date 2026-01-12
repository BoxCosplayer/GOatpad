package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"

    "github.com/mattn/go-runewidth"
    termbox "github.com/nsf/termbox-go"
)

const (
    // View and Edit
    MAX_MODES = 2
)

var (
    mode int

    ROWS int
    COLS int

    currentCol int
    currentRow int

    offsetCol int
    offsetRow int

    textBuffer = [][]rune{{}}
    undoBuffer = [][]rune{}
    copyBuffer = [][]rune{}

    file          string
    filename      string
    fileExtension string
    modified      bool
)

func read_file(filename string) {
    file, err := os.Open(filename)

    // File doesn't exist
    if err != nil {
        // sourceFile := filename
        textBuffer = append(textBuffer, []rune{})
        return
    }

    // `defer` delays the close until the end of function
    // Close will always occur, after error handling to avoid null pointer references
    defer file.Close()

    // Read the file w/ a scanner
    scanner := bufio.NewScanner(file)
    lineNumber := 0

    // For line, append to the textBuffer
    for scanner.Scan() {
        line := scanner.Text()
        for _, ch := range line {
            if ch == '\t' {
                for i := 0; i < TAB_WIDTH; i++ {
                    textBuffer[lineNumber] = append(textBuffer[lineNumber], ' ')
                }
            } else {
                textBuffer[lineNumber] = append(textBuffer[lineNumber], ch)
            }
        }
        textBuffer = append(textBuffer, []rune{})
        lineNumber++
    }

    // If new or empty file, pad textBuffer with empty text to not crash
    if lineNumber == 0 {
        textBuffer = append(textBuffer, []rune{})
    }
}

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

func scroll_text_buffer() {
    if currentRow < offsetRow {
        offsetRow = currentRow
    }
    if currentCol < offsetCol {
        offsetCol = currentCol
    }
    if currentRow >= offsetRow+ROWS {
        offsetRow = currentRow - ROWS + 1
    }
    if currentCol >= offsetCol+COLS {
        offsetCol = currentCol - COLS + 1
    }
}

func display_text_buffer() {
    var (
        cursorRow int
        cursorCol int
    )

    // For every Character...
    for cursorRow = 0; cursorRow < ROWS; cursorRow++ {
        textBufferRow := cursorRow + offsetRow

        // `writingCol` determines where to write to in the terminal
        // `textBuffercol` determines which character to read (& pull from the buffer)
        // `cursorCol` determines which character we are going to write
        writingCol := 0
        for cursorCol = 0; cursorCol < COLS; cursorCol++ {
            textBufferCol := cursorCol + offsetCol

            // Bound checking
            if textBufferRow >= 0 && textBufferRow < len(textBuffer) &&
                textBufferCol < len(textBuffer[textBufferRow]) {
                // ...Print character to terminal
                if textBuffer[textBufferRow][textBufferCol] != '\t' {
                    termbox.SetChar(writingCol, cursorRow, textBuffer[textBufferRow][textBufferCol])
                    writingCol++
                } else {
                    termbox.SetCell(writingCol, cursorRow, ' ', termbox.ColorDefault, termbox.ColorDefault)
                    writingCol++
                }
            } else if cursorRow+offsetRow > len(textBuffer)-1 {
                // Indicate EoF
                termbox.SetCell(0, cursorRow, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
            }
        }
        termbox.SetChar(cursorCol, cursorRow, rune('\n'))
    }
}

func display_status_bar() {
    var (
        modeStatus   string // current mode
        fileStatus   string // filename, total number of lines, modification status
        cursorStatus string // location of cursor (line, column)
        copyStatus   string // whether the copy buffer is active
        undoStatus   string // whether the undo buffer is active
    )

    if mode == 1 {
        modeStatus = " [EDIT] "
    } else {
        modeStatus = " [VIEW] "
    }

    if modified {
        fileStatus += " modified"
    } else {
        fileStatus += " saved"
    }

    if len(copyBuffer) > 0 {
        copyStatus = " [COPY]"
    }
    if len(undoBuffer) > 0 {
        undoStatus = " [UNDO]"
    }

    cursorStatus = " Line " + strconv.Itoa(currentRow+1) + " Col " + strconv.Itoa(currentCol+1) + " "

    fileStatus = fileExtension + " - " + strconv.Itoa(len(textBuffer)) + " lines" + fileStatus

    // Logic to clamp filename to the leftover space
    emptySpace := COLS - (len(modeStatus) + len(copyStatus) + len(undoStatus) + len(cursorStatus))
    filenameLength := len(filename)
    filenameSpace := emptySpace - len(fileStatus) - TAB_WIDTH - 2

    if filenameLength > filenameSpace {
        filenameLength = filenameSpace
        fileStatus = filename[:filenameLength] + ".." + fileStatus
    } else {
        fileStatus = filename[:filenameLength] + fileStatus
    }

    // Determine amount of space to create between left side and right side of status bar
    emptySpace = COLS - (len(modeStatus) + len(fileStatus) + len(copyStatus) + len(undoStatus) + len(cursorStatus)) - 4
    spaces := strings.Repeat(" ", emptySpace)

    message := modeStatus + fileStatus + copyStatus + undoStatus + spaces + cursorStatus
    print_message(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, message)

}

func print_message(col int, row int, fg termbox.Attribute, bg termbox.Attribute, message string) {
    // macro to loop a "SetCell" for any message
    for _, ch := range message {
        termbox.SetCell(col, row, ch, fg, bg)
        col += runewidth.RuneWidth(ch)
    }
}

func get_key() termbox.Event {
    // Function to detect and grab keypresses,
    // handled by process_key in keybinds.go
    var keyEvent termbox.Event

    switch event := termbox.PollEvent(); event.Type {
    case termbox.EventKey:
        keyEvent = event
    case termbox.EventError:
        panic(event.Err)
    }
    return keyEvent
}

func run_editor() {
    err := termbox.Init()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    // Check for args
    if len(os.Args) > 1 {
        file = os.Args[1]
        lastDotIndex := strings.LastIndex(file, ".")
        if lastDotIndex != -1 {
            fileExtension = file[lastDotIndex:]
            filename = file[:lastDotIndex]
        } else {
            fileExtension = ""
            filename = file
        }
        read_file(file)
    } else {
        filename = "out.txt"
        textBuffer = append(textBuffer, []rune{})
    }

    for {

        // -1 row for the status bar
        COLS, ROWS = termbox.Size()
        ROWS--

        // status bar errors is there is too little space
        if COLS < 80 {
            COLS = 80
        }

        // Empty the terminal, and show the template text
        termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
        scroll_text_buffer()
        display_text_buffer()
        display_status_bar()

        // Draw Cursor, and syncronise terminal
        termbox.SetCursor(currentCol-offsetCol, currentRow-offsetRow)
        termbox.Flush()

        // Wait for an event

        if currentCol > len(textBuffer[currentRow]) {
            currentCol = len(textBuffer[currentRow])
        }

        process_key()
    }
}

func main() {
    run_editor()
}
