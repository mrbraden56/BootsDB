// Implementation of the B+ Tree algorithm
package storage_manager

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type BTree struct {
	pager *Pager
}

func (btree *BTree) Insert(username string, email string) {
	root := btree.pager.get_root()
	header_size := 100

	// Initialize header if not already done
	if binary.BigEndian.Uint32(root.slotted_array[0:4]) == 0 {
		header := [100]byte{
			0x00, 0x00, 0x00, 0x01, // Bytes 0-3: Page number 1
			0x00, 0x02, // Bytes 4-5: Flags (table leaf node)
			0x00, 0x00, // Bytes 6-7: Number of cells (0)
			0x00, 0x00, 0x10, 0x00, // Bytes 8-11: Cell content offset (4096)
			0x00, 0x00, 0x0F, 0x9C, // Bytes 12-15: Number of free bytes (4096 - 100 = 3996)
			// Bytes 16-99: Zeros
		}
		copy(root.slotted_array[:header_size], header[:])
	}

	// Get current number of cells(These are pointers to tuples)
	N := binary.BigEndian.Uint16(root.slotted_array[6:8])

	// Get current cell content offset
	C := binary.BigEndian.Uint32(root.slotted_array[8:12])

	payloadSize := uint16(32 + 255)      // 287 bytes
	tupleSize := uint32(2 + payloadSize) // 289 bytes total

	// Calculate tuple offset (cells grow backwards from the end)
	tupleOffset := C - tupleSize

	// Build tuple
	tuple := make([]byte, tupleSize)
	binary.BigEndian.PutUint16(tuple[0:2], payloadSize)
	username32 := make([]byte, 32)
	email255 := make([]byte, 255)
	copy(username32, []byte(username))
	copy(email255, []byte(email))
	copy(tuple[2:34], username32)
	copy(tuple[34:], email255)

	// Update cell pointer array (starts at 100, 2 bytes per pointer)
	pointerOffset := uint32(header_size) + 2*uint32(N)
	binary.BigEndian.PutUint16(root.slotted_array[pointerOffset:pointerOffset+2], uint16(tupleOffset))

	// Copy tuple into page
	copy(root.slotted_array[tupleOffset:tupleOffset+tupleSize], tuple)

	// Update header
	binary.BigEndian.PutUint16(root.slotted_array[6:8], N+1)            // Increment number of cells
	binary.BigEndian.PutUint32(root.slotted_array[8:12], tupleOffset)   // Update cell content offset
	newFreeBytes := tupleOffset - (uint32(header_size) + 2*uint32(N+1)) // Calculate remaining free bytes
	binary.BigEndian.PutUint32(root.slotted_array[12:16], newFreeBytes) // Update free bytes
}

func (btree *BTree) Select() {
    root := btree.pager.get_root()
    header_size := 100
    tupleSize := 289

    if binary.BigEndian.Uint32(root.slotted_array[0:4]) == 0 {
        fmt.Println("No data in database")
    } else {
        N := binary.BigEndian.Uint16(root.slotted_array[6:8])
        for i := 0; i < int(N); i++ {
            pointerOffset := uint32(header_size) + uint32(2*i)
            pointerOffsetValue := binary.BigEndian.Uint16(root.slotted_array[pointerOffset:pointerOffset+2])
            disk_tuple := root.slotted_array[int(pointerOffsetValue):int(pointerOffsetValue)+tupleSize]
            username := strings.TrimRight(string(disk_tuple[2:34]), "\x00")
            email := strings.TrimRight(string(disk_tuple[34:34+255]), "\x00")
            fmt.Printf("Key: %d, Username: %s, Email: %s\n", i, username, email)
        }
    }
}

func InitializeBtree(pager_struct *Pager) *BTree {
	btree_struct := &BTree{
		pager: pager_struct,
	}

	return btree_struct
}
