package diskbst

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// writer struct implements Writer interface
type writer struct {
	cursor *os.File
	tail   int64
}

type Writer interface {
	Put(key []byte, value []byte) error
	Close()
}

func OpenWriter(pathName string) (Writer, error) {
	var w writer

	var justCreated bool

	cursor, err := os.OpenFile(pathName, os.O_RDWR, 0755)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		// create file
		cursor, err = os.OpenFile(pathName, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, err
		}

		// write the magic number to identify the file format
		_, err = cursor.Write(magicNumber)
		if err != nil {
			return nil, err
		}

		justCreated = true
	}

	if !justCreated {
		// validate magic number to ensure the file is a valid BST file
		mn := make([]byte, len(magicNumber))
		_, err = cursor.Read(mn)
		if err != nil {
			return nil, err
		}

		if bytes.Compare(mn, magicNumber) != 0 {
			return nil, errInvalidMagicNumber
		}
	}

	w.cursor = cursor

	// Determine the current size of the file (used to set the tail)
	info, err := cursor.Stat()
	if err != nil {
		return nil, err
	}

	w.tail = info.Size()

	return &w, nil
}

func (w *writer) Put(key []byte, value []byte) error {
	newNode := node{
		key:   key,
		value: value,
	}

	// write new node to disk at the end of file
	data := newNode.serialize()
	n, err := w.cursor.WriteAt(data, w.tail)
	if err != nil {
		return err
	}

	if n < len(data) {
		return fmt.Errorf("failed writing new node")
	}

	newTail := w.tail + int64(n)

	pos, err := w.findPos(key)
	if err != nil {
		return err
	}

	// If it's the root node, no linking is required
	if isRoot := pos == int64(len(magicNumber)); isRoot {
		w.tail = newTail
		return nil
	}

	// link new node to its parent
	nt := make([]byte, 8)
	binary.LittleEndian.PutUint64(nt, uint64(w.tail))
	n, err = w.cursor.WriteAt(nt, pos)
	if err != nil {
		return err
	}

	if n < len(nt) {
		return fmt.Errorf("failed writing new node offset to parent")
	}

	w.tail = newTail

	return nil
}

func (w *writer) Close() {
	_ = w.cursor.Close()
}

// returns location in the parent node where the new node’s offset should be written
func (w *writer) findPos(key []byte) (int64, error) {
	var (
		currNode       node
		childPtrOffset int64
	)
	nodeSize := make([]byte, 8)

	// when BST is empty
	if w.tail == int64(len(magicNumber)) {
		// first node
		return w.tail, nil
	}

	currPos := int64(len(magicNumber))

	for {
		n, err := w.cursor.ReadAt(nodeSize, currPos)
		if err != nil {
			return 0, err
		}

		if n < len(nodeSize) {
			// corruption
			return 0, fmt.Errorf("failed reading node size")
		}

		currPos += 8

		s := binary.LittleEndian.Uint64(nodeSize)

		nd := make([]byte, s)
		n, err = w.cursor.ReadAt(nd, currPos)
		if err != nil {
			return 0, err
		}

		if n < len(nd) {
			// corruption
			return 0, fmt.Errorf("failed reading node")
		}

		var next int64
		currNode.deserialize(nd)
		currPos += int64(8 + len(currNode.key) + 8 + len(currNode.value))

		if bytes.Compare(key, currNode.key) <= 0 {
			// key <= currNode.Key
			next = int64(currNode.leftChild)
			childPtrOffset = currPos
		} else {
			// key > currNode.key
			next = int64(currNode.rightChild)
			childPtrOffset = currPos + int64(8)
		}

		if next == 0 {
			// parent of the new node found
			break
		}

		currPos = next
	}

	return childPtrOffset, nil
}

var (
	magicNumber           = []byte{0xD, 0xB, 0xD}
	errInvalidMagicNumber = errors.New("invalid magic number")
)
