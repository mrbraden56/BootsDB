/*
Token Types:

	"select": Keyword - Represents the SQL SELECT statement for querying data
	"from": Keyword - Specifies the table to query from
	"where": Keyword - Defines conditions for filtering data
	"insert": Keyword - Adds new data to a table
	"update": Keyword - Modifies existing data in a table
	"delete": Keyword - Removes data from a table
	"create": Keyword - Defines a new table or structure
	"table": Keyword - Specifies a table in a creation or query
	"drop": Keyword - Deletes a table or structure
	"integer": Keyword - Indicates an integer data type
	"text": Keyword - Indicates a text data type
	"real": Keyword - Indicates a floating-point data type
	"primary_key": Keyword - Defines a primary key constraint
	"": Identifier - Represents a variable or table name (non-keyword)
	"=": Operator - Equality comparison
	";": Operator - Statement terminator
	",": Operator - Item separator
	"*": Operator - Wildcard, often for selecting all columns
	"'": Operator - Delimiter for string literals
	"(": Operator - Opens a grouped expression
	")": Operator - Closes a grouped expression
*/
package query_processor

import (
	"io"
	"os"
	"unicode"
)

var TokenMap = map[string]string{
	"select":      "Keyword",
	"from":        "Keyword",
	"where":       "Keyword",
	"insert_into": "Keyword",
	"update":      "Keyword",
	"delete":      "Keyword",
	"create":      "Keyword",
	"table":       "Keyword",
	"drop":        "Keyword",
	"values":      "Keyword",
	"integer":     "Keyword",
	"text":        "Keyword",
	"real":        "Keyword",
	"primary_key": "Keyword",
	"string":      "Literal",
	"number":      "Literal",
	"float":       "Literal",
	"=":           "Operator",
	";":           "Operator",
	",":           "Operator",
	"*":           "Operator",
	"'":           "Operator",
	"(":           "Operator",
	")":           "Operator",
}

type Token struct {
	Token_type string
	Val        string
	is_int     bool
	line       int
}

type Scanner struct {
	Tokens []*Token

	CurrentRune   rune   // The most recently read character
	CurrentLine   int    // Line number of the current character
	CurrentColumn int    // Column number of the current character
	Err           error  // Error state, e.g., io.EOF
	content       string // The entire file content stored in memory
	index         int    // Current position in the content string
	line          int    // Next line position
	column        int    // Next column position
}

func (s *Scanner) AddToken(text string, is_int bool) {
	if is_int {
		token := &Token{
			Token_type: "Literal",
			Val:        text,
			line:       s.CurrentLine,
			is_int:     is_int,
		}
		s.Tokens = append(s.Tokens, token)
		return
	}
	token_type := TokenMap[text]
	if text == "'" {
		s.Next()
		text = s.GetWord()
		s.Next()
		token_type = "Literal"
	}
	if token_type == "" {
		token_type = "Identifier"
	}
	token := &Token{
		Token_type: string(token_type),
		Val:        text,
		line:       s.CurrentLine,
		is_int:     is_int,
	}
	s.Tokens = append(s.Tokens, token)
}

// NewScanner creates a new Scanner by reading the entire file into memory
func NewScanner(filePath string) (*Scanner, error) {
	// Read the entire file into a byte slice and convert it to a string
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &Scanner{
		content: string(content),
		line:    1, // Start at line 1
		column:  1, // Start at column 1
	}, nil
}

// Next reads the next character from the in-memory content and updates the scanner's state
func (s *Scanner) Next() error {
	if s.Err != nil {
		return s.Err
	}

	// Check if we've reached the end of the content
	if s.index >= len(s.content) {
		s.Err = io.EOF
		return io.EOF
	}

	// Get the current character and update the scanner's state
	r := rune(s.content[s.index])
	s.CurrentRune = r
	s.CurrentLine = s.line
	s.CurrentColumn = s.column

	// Update line and column numbers based on the character
	if r == '\n' {
		s.line += 1  // Move to the next line
		s.column = 1 // Reset column to the start
	} else {
		s.column += 1 // Move to the next column
	}

	// Move to the next character
	s.index += 1
	return nil
}

func (s *Scanner) Prev() error {
	if s.Err != nil {
		return s.Err
	}

	// Check if we've reached the beginning of the content
	if s.index <= 0 {
		s.Err = io.EOF // Or a custom error for "start of file"
		return io.EOF
	}

	// Move back one character
	s.index -= 1

	// Get the previous character and update the scanner's state
	r := rune(s.content[s.index])
	s.CurrentRune = r

	// Update line and column numbers based on the character
	if r == '\n' {
		s.line -= 1 // Move to the previous line
		// To set the column correctly, we need to find the last column of the previous line
		if s.index > 0 {
			// Look back to find the previous newline or start of content
			prevLineStart := s.index - 1
			for prevLineStart >= 0 && rune(s.content[prevLineStart]) != '\n' {
				prevLineStart--
			}
			s.column = s.index - prevLineStart // Length from last newline (or start) to current pos
		} else {
			s.column = 1 // At the start of the content
		}
	} else {
		s.column -= 1 // Move to the previous column
	}

	s.CurrentLine = s.line
	s.CurrentColumn = s.column

	// Clear EOF if we were at the end and moved back
	if s.Err == io.EOF && s.index < len(s.content) {
		s.Err = nil
	}

	return nil
}

func (s *Scanner) IsChar() bool {
	return !unicode.IsSpace(s.CurrentRune) && s.CurrentRune != 0 && s.CurrentRune != ',' && s.CurrentRune != '\'' && s.CurrentRune != ')' && s.CurrentRune != ';'
}

func (s *Scanner) GetWord() string {
	text := ""
	for s.IsChar() {
		text += string(s.CurrentRune)
		s.Next()
	}
	s.Prev()

	return text
}
