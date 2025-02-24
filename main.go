package main

import (
	"BootsDB/storage_manager"
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

	storage_engine := storage_manager.InitializeStorage("Boots.db")
	pager := storage_manager.InitializePager(storage_engine)
	btree := storage_manager.InitializeBtree(pager)
	for {
		fmt.Print("BootsDB> ")
		scanner.Scan()
		input := scanner.Text()

		if input == ".exit" {
			pager.FlushCache()
			break
		}

		switch scanQuery(input) {
		case INSERT:
			parts := strings.Fields(input)
			username := parts[1]
			email := parts[2]
			btree.Insert(username, email)

		case SELECT:
			btree.Select()

		default:
			fmt.Println("Unknown command")
		}
	}
}

//DONE:
//Work on select functionality: DONE
//Work on marking pages dirty when we change cached pages DONE
//Work on writing back to disk DONE
//Ensure we have correct functionality for when root gets full DONE
//Ensure tests work DONE

//TODO:
//Implement splitting algorithm in insert when page gets full

//TODO LATER:
//Implement LRU Cache or similart to update cache
//Implement journaling/write ahead logging(WAL) which saves data when db crashes

