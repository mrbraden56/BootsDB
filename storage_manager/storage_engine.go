// The storage engine is responsible for all I/O of the database
package storage_manager

import "os"

type Storage struct {
	fileSize int64
	file     *os.File
}

func (storage *Storage) get_root_from_disk() *[4096]byte {
	storage.file.Seek(0, 0)

	var buffer [4096]byte
	storage.file.Read(buffer[:])
	return &buffer

}

func InitializeStorage(file_name string) *Storage {
	file, err := os.OpenFile(file_name, os.O_RDWR|os.O_CREATE, 0666)

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
	}
	storage_struct := &Storage{
		file:     file,
		fileSize: fileInfo.Size(),
	}

	return storage_struct
}
