package storage_engine

//TODO:
//1. Write the select functionality that prints out every row
//2. Generate tests using io functionalitu fir e2e testing
//3. Rewrite anything you want for the program, ensure tests still past, and continue to tree splitting

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
	"fmt"
)

type NodeType int

const (
	ROOT_NODE_TYPE        = 0
	LEAF_NODE_KV_STARTING = 11
	MAX_ROW_ID_ROOT       = 2
)

type Page struct {
	/*
		This is a 4kb array that will hold our data
	*/
	array     [4096]byte
	has_space bool
}

type Pager struct {
	/*
		Pager acts as buffer pool manager by hanlding blocks of data,
		writing to disk, and checking cache
	*/
	num_pages    int
	cached_pages []*Page
}

type Cursor struct {
	pager     Pager
	file_name string
}

func (c *Cursor) create_root() {
	root_page := Page{
		array:     [4096]byte{},
		has_space: true,
	}
	var node_type uint32 = uint32(ROOT_NODE_TYPE)
	var is_root uint32 = 0
	var num_cells uint32 = 1
	var max_row_id uint32 = 0
	root_page.array[0] = uint8(node_type)
	root_page.array[1] = uint8(is_root)
	root_page.array[MAX_ROW_ID_ROOT] = uint8(max_row_id)
	binary.BigEndian.PutUint32(root_page.array[7:11], num_cells)

	c.pager.cached_pages = append(c.pager.cached_pages, &root_page)
	c.pager.num_pages += 1
}

func (c *Cursor) Initialize(file_name string) {
	c.file_name = file_name
	fileInfo, err := os.Stat(file_name)
	if os.IsNotExist(err) {
		file, err := os.Create(file_name)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fileInfo, _ = file.Stat()
	}
	size := fileInfo.Size()
	c.pager.num_pages = int(size) / 4096
	if c.pager.num_pages == 0 {
		c.create_root()
	} else {
		file, err := os.Open(c.file_name)
		if err != nil {
			return
		}
		defer file.Close()
		buffer := make([]byte, len(c.pager.cached_pages[c.pager.num_pages-1].array))
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			return
		}
		var array [4096]byte
		copy(array[:], buffer[:]) // Copy bytes from buffer[n:] into array
		root_page := &Page{
			array:     array,
			has_space: array[MAX_ROW_ID_ROOT] <= 12,
		}
		if len(c.pager.cached_pages) == 0 {
			c.pager.cached_pages = make([]*Page, 1) // Allocate space
		}
		c.pager.cached_pages[0] = root_page
	}
}

func (c *Cursor) Insert(username string, email string) {
	/*
	   Insert adds a new username/email record into the B-tree.
	   Order of operations:
	   1. Get reference to root page
	   2. Traverse tree to find target page for insertion:
	      - If current page has space, this is our target
	      - If current page is full:
	        - If no child pages exist, create a new page
	        - If child pages exist, compare keys to find correct child page
	      - Repeat until we find a page with space or need to split
	   3. Once target page is found:
	      - If page has space, insert the record
	      - If page is full, perform page split:
	        - Create new page
	        - Redistribute records between pages
	        - Update parent page with new split info
	   4. Update any necessary page metadata (number of cells, keys, etc)
	*/

	insert_data := func(page *Page) {
		username32 := make([]byte, 32)
		email255 := make([]byte, 255)
		copy(username32, []byte(username))
		copy(email255, []byte(email))
		tuple := append(username32, email255...)
		curr_key := uint32(page.array[MAX_ROW_ID_ROOT])
		position := LEAF_NODE_KV_STARTING + (curr_key * 291)
		binary.BigEndian.PutUint32(page.array[position:position+4], curr_key)
		copy(page.array[position+4:position+291], tuple)
		page.array[MAX_ROW_ID_ROOT] += 1
	}

	root_page := c.pager.cached_pages[0]

	curr_page := root_page
	for {
		if curr_page.has_space {
			break
		}

		// if !hasChildren(curr_page) {
		//     // Create new page since we're at leaf with no space
		//     new_page := createNewPage()
		//     // Setup new page relationships
		//     break
		// }

		// // Move to correct child page based on key comparison
		// curr_page = findChildPage(curr_page, key)
	}

	insert_data(curr_page)

}

func (c *Cursor) Select() {
    root_page := c.pager.cached_pages[0]
    
    num_records := root_page.array[MAX_ROW_ID_ROOT]
    
    for key := 0; key < int(num_records); key++ {
        position := LEAF_NODE_KV_STARTING + key*291
        
        record_key := binary.BigEndian.Uint32(root_page.array[position : position+4])
        
        username_bytes := root_page.array[position+4 : position+36]
        username := strings.TrimRight(string(username_bytes), "\x00")
        
        email_bytes := root_page.array[position+36 : position+291]
        email := strings.TrimRight(string(email_bytes), "\x00")
        
        fmt.Printf("Key: %d, Username: %s, Email: %s\n", record_key, username, email)
    }
}

func (c *Cursor) Flush() {

	os.WriteFile(c.file_name, c.pager.cached_pages[c.pager.num_pages-1].array[:], 0644)
}

func NewStorageEngine(file_name string) *Cursor {
	engine := &Cursor{}
	engine.Initialize(file_name)
	return engine
}
