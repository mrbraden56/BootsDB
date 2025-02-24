/**
 * B+ Tree Node Structure for SQLite Variant with Doubly Linked Nodes
 * Page size: 4096 bytes (configurable in database header)
 *
 * **Database Metadata Header (100 bytes, file start):**
 * - 0-15: Magic string (16 bytes, "SQLite format 3\0")
 * - 16-17: Page size (uint16_t, 512-65536, e.g., 4096)
 * - 18-19: File format write version (uint16_t, typically 1 or 2)
 * - 20-21: File format read version (uint16_t, typically 1 or 2)
 * - 22: Reserved space (uint8_t, usually 0)
 * - 23: Max payload fraction (uint8_t, default 64)
 * - 24: Min payload fraction (uint8_t, default 32)
 * - 25: Leaf payload fraction (uint8_t, default 32)
 * - 26-29: File change counter (uint32_t, increments on writes)
 * - 30-33: Database size in pages (uint32_t)
 * - 34-37: First freelist trunk page (uint32_t, 0 if none)
 * - 38-41: Freelist page count (uint32_t)
 * - 42-45: Schema cookie (uint32_t)
 * - 46-49: Schema format number (uint32_t, typically 4)
 * - 50-53: Default cache size (uint32_t)
 * - 54-57: Largest root page (uint32_t, 0 if no auto-vacuum)
 * - 58-61: Text encoding (uint32_t, 1=UTF-8, 2=UTF-16le, 3=UTF-16be)
 * - 62-65: User version (uint32_t)
 * - 66-69: Incremental vacuum mode (uint32_t, non-zero enables)
 * - 70-73: Application ID (uint32_t)
 * - 74-91: Reserved (18 bytes, zeros)
 * - 92-95: Version valid for (uint32_t)
 * - 96-99: SQLite version number (uint32_t, e.g., 9999999 for this variant)
 *
 * **Common Node Header (20 bytes):**
 * - 0-3: Page number (uint32_t, unique ID, root init 1)
 * - 4: Flags (uint8_t, 0x01 leaf, 0x00 internal)
 * - 5-6: Cell count (uint16_t)
 * - **Leaf Only:**
 *   - 7-8: Cell content offset (uint16_t, init 20)
 *   - 9-10: Free bytes (uint16_t)
 *   - 11-12: Total cell content bytes (uint16_t)
 * - **Internal Only:**
 *   - 7-10: Rightmost child pointer (uint32_t, leftmost child < first key)
 *   - 11-12: Free bytes (uint16_t)
 * - **Both:**
 *   - 13-16: Next sibling pointer (uint32_t, 0 if none)
 *   - 17-20: Previous sibling pointer (uint32_t, 0 if none)
 *
 * **Leaf Node:**
 * - Flags: 0x01
 * - Cells: Data records (variable size, e.g., key-value tuples)
 * - Sibling Pointers: Next (13-16), Previous (17-20)
 *
 * **Internal Node:**
 * - Flags: 0x00
 * - Cells: (pointer, key) pairs (variable size, e.g., 4-byte child pointer, key)
 * - Rightmost Child: 7-10 (leftmost child in B+-tree convention)
 * - Sibling Pointers: Next (13-16), Previous (17-20)
 *
 * **Root Node:**
 * - Leaf or internal, per tree height
 * - On page 1, starts at byte 100 after metadata
 *
 * **Slotted Array:**
 * - Pointers: 2-byte offsets from byte 20 (or 100 for root on page 1) up
 * - Data: Cells from cell content offset down, sorted by key
 * - Overflow: Pointers and data meet
 */

package storage_manager

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type BTree struct {
	pager *Pager
	order int
}

// func (btree *BTree) binary_search(page *Pager) {

// }

// func (btree *BTree) find_leaf_node(root *Page, key uint32) *Page {

// }

// func (btree *BTree) Insert(username string, email string) {
// 	root := btree.pager.get_root()
// 	key := binary.BigEndian.Uint32(root.slotted_array[16:20])
// 	leafNode := btree.find_leaf_node(root, key)             // Traverse to leaf using the key
// 	N := binary.BigEndian.Uint16(leafNode.slotted_array[6:8]) // Number of cells in the leaf
// 	pageFull := N >= uint16(btree.order-1)                    // Max keys = order - 1
// 	if pageFull {
// 		btree.split_and_insert(leafNode, username, email)
// 	} else {
// 		btree.insert_into_page_slot(leafNode, username, email)
// 	}
// }

func (btree *BTree) Insert(username string, email string) {
    root := btree.pager.get_root()

    // Check if this is an initialized page by looking for SQLite header
    isInitialized := string(root.slotted_array[0:16]) == "BootsDB format 3"

    node_header_offset := 100 // For root page
	common_node_header := 20
	var free_bytes int
    
    if !isInitialized {
        // Initialize database metadata header (bytes 0-99)
        header := make([]byte, 100)
        copy(header[0:16], []byte("BootsDB format 3\000"))
        binary.BigEndian.PutUint16(header[16:18], 4096) // Page size
        header[18] = 1 // Write version
        header[19] = 1 // Read version
        header[22] = 0 // Reserved
        header[23] = 64 // Max payload fraction
        header[24] = 32 // Min payload fraction
        header[25] = 32 // Leaf payload fraction
        binary.BigEndian.PutUint32(header[26:30], 0) // File change counter
        binary.BigEndian.PutUint32(header[30:34], 1) // Database size in pages
        binary.BigEndian.PutUint32(header[34:38], 0) // First freelist trunk page
        binary.BigEndian.PutUint32(header[38:42], 0) // Freelist page count
        binary.BigEndian.PutUint32(header[42:46], 0) // Schema cookie
        binary.BigEndian.PutUint32(header[46:50], 4) // Schema format number
        binary.BigEndian.PutUint32(header[50:54], 0) // Default cache size
        binary.BigEndian.PutUint32(header[54:58], 0) // Largest root page
        binary.BigEndian.PutUint32(header[58:62], 1) // Text encoding (UTF-8)
        binary.BigEndian.PutUint32(header[62:66], 0) // User version
        binary.BigEndian.PutUint32(header[66:70], 0) // Incremental vacuum mode
        binary.BigEndian.PutUint32(header[70:74], 0) // Application ID
        // Reserved bytes 74-91 are zeros
        binary.BigEndian.PutUint32(header[92:96], 0) // Version valid for
        binary.BigEndian.PutUint32(header[96:100], 9999999) // SQLite version number
        copy(root.slotted_array[0:100], header)

        // Initialize node header (bytes 100-120)
        node_header := make([]byte, 21)
        binary.BigEndian.PutUint32(node_header[0:4], 1) // Page number = 1 for root
        node_header[4] = 0x01 // Flags: leaf node
        binary.BigEndian.PutUint16(node_header[5:7], 0) // Cell count
        binary.BigEndian.PutUint16(node_header[7:9], 4096) // Cell content offset
        free_bytes = 4096 - 120 // 100 metadata + 20 node header
        binary.BigEndian.PutUint16(node_header[9:11], uint16(free_bytes))
        binary.BigEndian.PutUint16(node_header[11:13], 0) // Total cell content bytes
        binary.BigEndian.PutUint32(node_header[13:17], 0) // Next sibling
        binary.BigEndian.PutUint32(node_header[17:21], 0) // Previous sibling
        copy(root.slotted_array[100:121], node_header)
    }    
    // Get current number of cells
    N := binary.BigEndian.Uint16(root.slotted_array[node_header_offset+5:node_header_offset+7])
    
    // Get current cell content offset
    C := binary.BigEndian.Uint16(root.slotted_array[node_header_offset+7:node_header_offset+9])

    payloadSize := uint16(32 + 255)      // 287 bytes
    tupleSize := uint32(2 + payloadSize) // 289 bytes

    // Calculate tuple offset
    tupleOffset := uint32(C) - tupleSize

    // Calculate pointer offset for the new cell
    pointerOffset := uint32(node_header_offset + common_node_header + 2*int(N))

    // Check if page is full
    if tupleOffset < pointerOffset+2 {
        fmt.Println("Page Overflow")
        return
    }

    // Build tuple
    tuple := make([]byte, tupleSize)
    binary.BigEndian.PutUint16(tuple[0:2], payloadSize)
    username32 := make([]byte, 32)
    email255 := make([]byte, 255)
    copy(username32, []byte(username))
    copy(email255, []byte(email))
    copy(tuple[2:34], username32)
    copy(tuple[34:], email255)

	order := free_bytes / (2 + 255 + 32)
	btree.order = order

    // Set pointer
    binary.BigEndian.PutUint16(root.slotted_array[pointerOffset:pointerOffset+2], uint16(tupleOffset))

    // Copy tuple
    copy(root.slotted_array[tupleOffset:tupleOffset+tupleSize], tuple)

    // Update header
    binary.BigEndian.PutUint16(root.slotted_array[node_header_offset+5:node_header_offset+7], N+1)
    binary.BigEndian.PutUint16(root.slotted_array[node_header_offset+7:node_header_offset+9], uint16(tupleOffset))
    
    // Calculate new free bytes
    newFreeBytes := tupleOffset - (uint32(node_header_offset+20) + 2*uint32(N+1))
    binary.BigEndian.PutUint16(root.slotted_array[node_header_offset+9:node_header_offset+11], uint16(newFreeBytes))
    
    // Update total cell content bytes
    totalCellContent := binary.BigEndian.Uint16(root.slotted_array[node_header_offset+11:node_header_offset+13]) + uint16(tupleSize)
    binary.BigEndian.PutUint16(root.slotted_array[node_header_offset+11:node_header_offset+13], totalCellContent)

    // Mark page as dirty
    root.dirty = true
}

func (btree *BTree) Select() {
    root := btree.pager.get_root()
    node_header_offset := 100
    pageNumber := binary.BigEndian.Uint32(root.slotted_array[node_header_offset:node_header_offset+4])
    fmt.Println("Reading from page number:", pageNumber)
    N := binary.BigEndian.Uint16(root.slotted_array[node_header_offset+5:node_header_offset+7])
    tupleSize := 289

    // Iterate through cells with proper bounds checking
    for i := 0; i < int(N); i++ {
        pointerOffset := uint32(node_header_offset + 20 + 2*i)
        pointerOffsetValue := binary.BigEndian.Uint16(root.slotted_array[pointerOffset:pointerOffset+2])
        disk_tuple := root.slotted_array[int(pointerOffsetValue):int(pointerOffsetValue)+tupleSize]
        username := strings.TrimRight(string(disk_tuple[2:34]), "\x00")
        email := strings.TrimRight(string(disk_tuple[34:34+255]), "\x00")
        fmt.Printf("Key: %d, Username: %s, Email: %s\n", i, username, email)
    }
}

func InitializeBtree(pager_struct *Pager) *BTree {
	btree_struct := &BTree{
		pager: pager_struct,
	}
	return btree_struct
}
