## Disk-Based Binary Search Tree (BST) for LSM Key-Value Store

This implementation provides a simple disk-based binary search tree (BST) that can be used in an LSM key-value store to serialize a memtable to disk. 

### 1. Node Structure (`node.go`)
- Each node in the BST contains:
    - `key`: The key as a byte slice.
    - `value`: The value associated with the key.
    - `leftChild`: Offset in the file where the left child node is stored.
    - `rightChild`: Offset in the file where the right child node is stored.
- Nodes can be serialized to a byte slice for storage on disk and deserialized from a byte slice when reading from disk.

### 2. Writer (`writer.go`)
- The writer handles inserting key-value pairs into the BST and writing them to a file.
- **Key Functions:**
    - `OpenWriter(pathName string)`: Opens or creates a new file for writing the BST. The file is validated with a magic number to ensure it's a valid BST file.
    - `Put(key []byte, value []byte)`: Inserts a key-value pair into the BST. It writes the new node to the disk and links it to its parent.
    - `findPos(key []byte)`: Finds the position in the file where a new node should be inserted.
    - `Close()`: Closes the file after writing.

### 3. Reader (`reader.go`)
- The reader uses memory-mapped I/O for fast access to the data, allowing the file to be accessed through a byte slice in memory
- **Key Functions:**
    - `OpenReader(pathName string)`: Opens the file for reading and validates it with a magic number.
    - `Get(key []byte)`: Searches the BST for a key and returns the associated value.
    - `Close()`: Unmaps the memory-mapped file.

### Example usage:

``` go
package main

import (
	"fmt"
	"github.com/ashwaniYDV/diskbst"
)

func main() {
	memtable := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	// Open a new file to store the tree
	writer, _ := diskbst.OpenWriter("./test.bst")

	// Store all key-value pairs in the tree
	for key, val := range memtable {
		_ = writer.Put([]byte(key), val)
	}

	writer.Close()

	// Open a reader for the tree
	// the tree file will be memory-mapped
	reader, _ := diskbst.OpenReader("./test.bst")
	defer reader.Close()

	val, _ := reader.Get([]byte("key1"))
	fmt.Println(string(val)) // value1

	_, err := reader.Get([]byte("non_existent_key"))
	fmt.Println(err.Error()) // key not found
}
```

### Limitations of current approach (Further improvements)
* If a key already exists, then it's not overwritten
* Binary tree is not self-balanced