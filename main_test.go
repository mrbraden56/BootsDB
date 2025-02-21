package main

import (
    "bytes"
    "os"
    "os/exec"
    "strings"
    "testing"
    "fmt"
)

// runProgram executes the main program with the given inputs and returns its output.
func runProgram(t *testing.T, inputs string) string {
    cmd := exec.Command("go", "run", "main.go")
    cmd.Stdin = strings.NewReader(inputs)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        t.Fatalf("Failed to run program: %v", err)
    }
    return out.String()
}

// extractSelectLines extracts lines from the output that represent SELECT results.
func extractSelectLines(output string) []string {
    var lines []string
    for _, line := range strings.Split(output, "\n") {
        trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "BootsDB> ") {
			trimmed = strings.ReplaceAll(trimmed, "BootsDB> ", "")
		}
        if strings.HasPrefix(trimmed, "Key: ") {
            lines = append(lines, trimmed)
        }
    }
    return lines
}

// TestSinglePageOperations contains end-to-end tests for single-page functionality.
func TestSinglePageOperations(t *testing.T) {
    // Test Case 1: Insert and Select a few records
    os.Remove("Boots.db")
    inputs := `INSERT user1 user1@example.com
INSERT user2 user2@example.com
SELECT
.exit
`
    output := runProgram(t, inputs)
    selectLines := extractSelectLines(output)
    expected := []string{
        "Key: 0, Username: user1, Email: user1@example.com",
        "Key: 1, Username: user2, Email: user2@example.com",
    }
    if len(selectLines) != len(expected) {
        t.Errorf("Expected %d select lines, got %d", len(expected), len(selectLines))
    }
    for i, line := range selectLines {
        if line != expected[i] {
            t.Errorf("Expected: %s, Got: %s", expected[i], line)
        }
    }

    //Test Case 2: Insert 12 records and Select (page limit)
    os.Remove("Boots.db")
    var insertCommands strings.Builder
    for i := 0; i < 12; i++ {
        insertCommands.WriteString(fmt.Sprintf("INSERT user%d user%d@example.com\n", i+1, i+1))
    }
    inputs = insertCommands.String() + "SELECT\n.exit\n"
    output = runProgram(t, inputs)
    selectLines = extractSelectLines(output)
    if len(selectLines) != 12 {
        t.Errorf("Expected 12 select lines, got %d", len(selectLines))
    }
    for i := 0; i < 12; i++ {
        expectedLine := fmt.Sprintf("Key: %d, Username: user%d, Email: user%d@example.com", i, i+1, i+1)
        if selectLines[i] != expectedLine {
            t.Errorf("Expected: %s, Got: %s", expectedLine, selectLines[i])
        }
    }

    // Test Case 3: Persistence across program restarts
    os.Remove("Boots.db")
    // First run: insert records and exit
    inputs = `INSERT user1 user1@example.com
INSERT user2 user2@example.com
.exit
`
    runProgram(t, inputs)
    // Second run: select records and exit
    inputs = `SELECT
.exit
`
    output = runProgram(t, inputs)
    selectLines = extractSelectLines(output)
    expected = []string{
        "Key: 0, Username: user1, Email: user1@example.com",
        "Key: 1, Username: user2, Email: user2@example.com",
    }
    if len(selectLines) != len(expected) {
        t.Errorf("Expected %d select lines after restart, got %d", len(expected), len(selectLines))
    }
    for i, line := range selectLines {
        if line != expected[i] {
            t.Errorf("Expected: %s, Got: %s", expected[i], line)
        }
    }
}