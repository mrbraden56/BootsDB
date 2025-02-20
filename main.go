package main

import (
	"BootsDB/storage_engine"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type QueryType string

const (
	INSERT QueryType = "INSERT"
	SELECT QueryType = "SELECT"
)

func scanQuery(input string) QueryType {
	cmd := strings.ToUpper(strings.TrimSpace(input))
	if strings.HasPrefix(cmd, string(INSERT)) {
		return INSERT
	}
	if strings.HasPrefix(cmd, string(SELECT)) {
		return SELECT
	}
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	storage_engine := storage_engine.NewStorageEngine("Boots.db")
	for {
		fmt.Print("BootsDB> ")
		scanner.Scan()
		input := scanner.Text()

		if input == ".exit" {
			storage_engine.Flush()
			break
		}

		switch scanQuery(input) {
		case INSERT:
			parts := strings.Fields(input)
			username := parts[1]
			email := parts[2]
			storage_engine.Insert(username, email)

		case SELECT:
			storage_engine.Select()

		default:
			fmt.Println("Unknown command")
		}
	}
}
