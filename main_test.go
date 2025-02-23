package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
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
        // Remove all leading "BootsDB> " prefixes
        for strings.HasPrefix(trimmed, "BootsDB> ") {
            trimmed = strings.TrimPrefix(trimmed, "BootsDB> ")
            trimmed = strings.TrimSpace(trimmed) // Handle any extra spaces after prefix removal
        }
        if strings.HasPrefix(trimmed, "Key: ") {
            lines = append(lines, trimmed)
        }
    }
    return lines
}

// TestBTreeFunctionality contains end-to-end tests for BTree functionality.
func TestBTreeFunctionality(t *testing.T) {
    const maxN = 13 // Maximum number of records before overflow, calculated based on page size and tuple size

    // **Test Case 1: Insert and Select a few records**
    // Tests basic insert and select functionality with a small number of records.
    t.Run("InsertAndSelectFewRecords", func(t *testing.T) {
        os.Remove("Boots.db") // Start with a fresh database
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
    })

    // **Test Case 2: Insert maximum number of records without overflow and Select**
    // Tests select functionality with the maximum number of records that fit in a page (13).
    t.Run("InsertMaxRecordsWithoutOverflow", func(t *testing.T) {
        os.Remove("Boots.db")
        var insertCommands strings.Builder
        for i := 0; i < maxN; i++ {
            insertCommands.WriteString(fmt.Sprintf("INSERT user%d user%d@example.com\n", i+1, i+1))
        }
        inputs := insertCommands.String() + "SELECT\n.exit\n"
        output := runProgram(t, inputs)
        selectLines := extractSelectLines(output)
        if len(selectLines) != maxN {
            t.Errorf("Expected %d select lines, got %d", maxN, len(selectLines))
        }
        for i := 0; i < maxN; i++ {
            expectedLine := fmt.Sprintf("Key: %d, Username: user%d, Email: user%d@example.com", i, i+1, i+1)
            if selectLines[i] != expectedLine {
                t.Errorf("Expected: %s, Got: %s", expectedLine, selectLines[i])
            }
        }
    })

    // **Test Case 3: Try to insert beyond capacity, check for overflow message, and select to ensure only maxN records are present**
    // Tests page overflow detection by attempting to insert one record beyond capacity (14th record).
    // Verifies that "Page Overflow" is printed and only 13 records are stored.
    t.Run("InsertBeyondCapacity", func(t *testing.T) {
        os.Remove("Boots.db")
        var insertCommands strings.Builder
        // Attempt to insert maxN + 1 records
        for i := 0; i < maxN+1; i++ {
            insertCommands.WriteString(fmt.Sprintf("INSERT user%d user%d@example.com\n", i+1, i+1))
        }
        inputs := insertCommands.String() + "SELECT\n.exit\n"
        output := runProgram(t, inputs)
        // Check for overflow message
        if !strings.Contains(output, "Page Overflow") {
            t.Errorf("Expected 'Page Overflow' in output, but not found")
        }
        // Verify that only maxN records were inserted
        selectLines := extractSelectLines(output)
        if len(selectLines) != maxN {
            t.Errorf("Expected %d select lines, got %d", maxN, len(selectLines))
        }
        for i := 0; i < maxN; i++ {
            expectedLine := fmt.Sprintf("Key: %d, Username: user%d, Email: user%d@example.com", i, i+1, i+1)
            if selectLines[i] != expectedLine {
                t.Errorf("Expected: %s, Got: %s", expectedLine, selectLines[i])
            }
        }
    })

    // **Test Case 4: Persistence across program restarts**
    // Tests that inserted records persist after the program is closed and reopened.
    // Inserts 5 records, exits, reopens, and verifies the records are still present.
    t.Run("PersistenceAcrossRestarts", func(t *testing.T) {
        os.Remove("Boots.db")
        inputs := `INSERT user1 user1@example.com
INSERT user2 user2@example.com
INSERT user3 user3@example.com
INSERT user4 user4@example.com
INSERT user5 user5@example.com
.exit
`
        runProgram(t, inputs)
        inputs = `SELECT
.exit
`
        output := runProgram(t, inputs)
        selectLines := extractSelectLines(output)
        expected := []string{
            "Key: 0, Username: user1, Email: user1@example.com",
            "Key: 1, Username: user2, Email: user2@example.com",
            "Key: 2, Username: user3, Email: user3@example.com",
            "Key: 3, Username: user4, Email: user4@example.com",
            "Key: 4, Username: user5, Email: user5@example.com",
        }
        if len(selectLines) != len(expected) {
            t.Errorf("Expected %d select lines after restart, got %d", len(expected), len(selectLines))
        }
        for i, line := range selectLines {
            if line != expected[i] {
                t.Errorf("Expected: %s, Got: %s", expected[i], line)
            }
        }
    })

    // **Test Case 5: Insert record with maximum-sized username and email**
    // Tests select functionality with records that have maximum-sized fields (32 bytes for username, 255 bytes for email).
    t.Run("InsertMaxSizedRecord", func(t *testing.T) {
        os.Remove("Boots.db")
        maxUsername := strings.Repeat("a", 32)
        maxEmail := strings.Repeat("b", 255)
        inputs := fmt.Sprintf("INSERT %s %s\nSELECT\n.exit\n", maxUsername, maxEmail)
        output := runProgram(t, inputs)
        selectLines := extractSelectLines(output)
        expectedLine := fmt.Sprintf("Key: 0, Username: %s, Email: %s", maxUsername, maxEmail)
        if len(selectLines) != 1 {
            t.Errorf("Expected 1 select line, got %d", len(selectLines))
        }
        if selectLines[0] != expectedLine {
            t.Errorf("Expected: %s, Got: %s", expectedLine, selectLines[0])
        }
    })
}