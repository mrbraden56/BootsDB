package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

const PAGE_SIZE = 4096

type QueryType string

const (
	INSERT QueryType = "INSERT"
	SELECT QueryType = "SELECT"
)

type Page struct {
	tuples [PAGE_SIZE]byte
	id     int
}

type File struct {
	pages   []Page
	pageDir map[int]int
}

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
	page := &Page{
		tuples: [PAGE_SIZE]byte{},
		id:     0, //tracks lastest id number
	}

	file := &File{
		pages:   make([]Page, 0),
		pageDir: make(map[int]int), //tracks latest entry for each page
	}
	file.pages = append(file.pages, *page)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("BootsDB> ")
		scanner.Scan()
		input := scanner.Text()

		if input == ".exit" {
			break
		}

		switch scanQuery(input) {
		case INSERT:
			if file.pageDir[len(file.pages)-1]+291 > PAGE_SIZE {
				newPage := Page{
					tuples: [PAGE_SIZE]byte{},
					id:     0,
				}
				file.pages = append(file.pages, newPage)
				file.pageDir[len(file.pages)-1] = 0
			}

			fields := strings.Split(scanner.Text(), " ")
			username := fields[1]
			email := fields[2]
			page_amount := len(file.pages) - 1
			id := file.pages[page_amount].id

			usernameBytes := []byte(username)
			emailBytes := []byte(email)
			idBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(idBytes, uint32(id))

			var usernameArray [255]byte
			var emailArray [32]byte
			var idArray [4]byte
			copy(usernameArray[:], usernameBytes)
			copy(emailArray[:], emailBytes)
			copy(idArray[:], idBytes)

			var tuple [291]byte
			copy(tuple[0:4], idArray[:])
			copy(tuple[4:259], usernameArray[:])
			copy(tuple[259:291], emailArray[:])
			curr := file.pageDir[page_amount]
			copy(file.pages[page_amount].tuples[curr:curr+291], tuple[:])

			file.pages[page_amount].id += 1
			file.pageDir[page_amount] += 291

		case SELECT:
			for _, page := range file.pages {
				for i := 0; i < page.id; i++ {
					offset := i * 291
					id := binary.LittleEndian.Uint32(page.tuples[offset : offset+4])
					username := string(bytes.Trim(page.tuples[offset+4:offset+36], "\x00"))
					email := string(bytes.Trim(page.tuples[offset+36:offset+291], "\x00"))
					fmt.Printf("id: %d, username: %s, email: %s\n", id, username, email)
				}
			}
		default:
			fmt.Println("Unknown command")
		}
	}
}
