// The pager is resposible for managing the buffer pool and interacting with
// the storage engine and b+ tree on what data needs to be written
package storage_manager

import (
	"container/list"
	"sync"
)

type Page struct {
	slotted_array [4096]byte
	dirty         bool
	page_number   int
}

type PageCache struct {
	content map[int]*Page
	lruList *list.List   // Doubly-linked list for LRU eviction order (front = most recent, back = least recent).
	maxSize int          // Maximum number of pages the cache can hold.
	lock    sync.RWMutex // Mutex for thread-safe access (read-write lock for concurrency).
}

type Pager struct {
	cache   *PageCache
	storage *Storage
}

func (pager *Pager) get_root() *Page {
	var page *Page

	_, root_in_cache := pager.cache.content[0]
	if root_in_cache {
		page = pager.cache.content[0]
	} else {
		root_slotted_array := pager.storage.get_root_from_disk()
		page = &Page{
			slotted_array: *root_slotted_array,
			dirty:         false,
		}
		pager.cache.content[0] = page
	}

	return page
}

func (pager *Pager) FlushCache(){
	root := pager.cache.content[0]
	if root.dirty{
		pager.storage.write_page_to_disk(root)
	}
}

func InitializePager(storage_struct *Storage) *Pager {
	cache := &PageCache{
		content: make(map[int]*Page), // Initialize the map to avoid nil map panics.
		lruList: list.New(),          // Create a new empty doubly-linked list for LRU tracking.
		maxSize: 500,                 // Set the maximum size as provided.
		lock:    sync.RWMutex{},      // Initialize the RWMutex (zero value is usable, included for clarity).
	}

	pager_struct := &Pager{
		cache:   cache,
		storage: storage_struct,
	}
	return pager_struct
}
