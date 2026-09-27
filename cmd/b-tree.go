package main

import (
	"encoding/binary"
	"github.com/abode1234/golang/key-value/utils-go"
)

type BNode struct {
	data []byte
}

// type of the node 
const (
	BNODE_NODE = 1 // internal node
	BNODE_LEAF = 2 // leaf node
)

// BTree structure
type BTree struct {
	root uint64 // disk page number
	// callbacks to manage the disk pages references
	get func(uint64)  BNode // to references the pointers
	new func(BNode) uint64  // to allocate a new page
	del func(uint64)		// to deallocate a page
}

// Header logical structure

const (
	HEADER          = 4    // header size in bytes
	BTREE_PAGE_SIZE = 4096 // page size in bytes
	// size of the constants in the header
	BTREE_MAX_KEYS_SIZE   = 1000 // max size (bytes) of a single key
	BTREE_MAX_VALUES_SIZE = 3000 // max size (bytes) of a single value
)

// init the constants
// the header is composed by the following elements 
// Offset:  [0-3]   [4-11]    [12-13]   [14-17]   [18...]        [...]
//          HEADER  pointer   offset    klen+vlen key            value
//          (4B)    (8B)      (2B)      (4B)      (up to 1000B)  (up to 3000B)
// 1. HEADER is 4 bytes size  type node btype [2B] nkeys [2B]
// 2. 8B is the pointer to the next node  uint64
// 3. 2B is the offset to the key value pair uint16
// 4. 4B is the key length(klen) and the value length(vlen) uint16
// 5. BTREE_MAX_KEYS_SIZE is the maximum number of keys in the node
// 6. BTREE_MAX_VALUES_SIZE is the maximum number of values in the node

func init() {
	node1max := HEADER + 8 + 2 + 4 + BTREE_MAX_KEYS_SIZE + BTREE_MAX_VALUES_SIZE
	utils.Assert(node1max <= BTREE_PAGE_SIZE, "exceeded the page size")
}

// Header functions
// return the node type
// the first two bytes of the header it's define the node type
// in memory digram 
// Byte:      [0] [1] [2] [3] [4] [5] ...
// Content:   |--type--|--nkeys--| ... rest of data
//            (2 bytes) (2 bytes)

func  (b BNode) btype() uint16 {
	return  binary.LittleEndian.Uint16(b.data) // read the btype 0 -> 1
}

// return the number of keys in the node
// the second two bytes of the header it's define the number of keys
// digram
//	Byte:      [0] [1] [2] [3] [4] [5] ...
//	Content:   |--type--|--nkeys--| ... rest of data
//	           (2 bytes) (2 bytes)
func (b BNode) nkeys() uint16 {
	return binary.LittleEndian.Uint16(b.data[2:4]) // read the nkeys 2 -> 3
}

// set a header with the first two bytes for the node type and the second two bytes for the number of keys
// digram
//	Byte:      [0] [1] [2] [3] [4] [5] ...
//	Content:   |--type--|--nkeys--| ... rest of data
//	           (2 bytes) (2 bytes)
func (b BNode) setHeader(btype uint16, nkeys uint16) {
	binary.LittleEndian.PutUint16(b.data, btype)      // set the node type (btype) 0 -> 1
	binary.LittleEndian.PutUint16(b.data[2:4], nkeys) // set the number of keys (nkeys) 2 -> 3
}

// Pointer functions
// return the pointer to the next node (child node) size 8 bytes uint64
// digram
// Offset:    [4-11]
//            pointer
//            (8B)

func (b BNode) getPtr(idx uint16) uint64 {
	utils.Assert(idx < b.nkeys(), "index out of range")  // check the index
	return binary.LittleEndian.Uint64(b.data[HEADER+idx*8:]) // read the pointer 4 -> 11
}

// update the pointer to the next node (child node) size 8 bytes uint64
// digram
// Offset:    [4-11]
//            pointer
//            (8B)

func (b BNode) setPtr(idx uint16, val uint64) {
	utils.Assert(idx < b.nkeys(), "index out of range")
	binary.LittleEndian.PutUint64(b.data[HEADER+idx*8:], val) // write the pointer 4 -> 11
}

// Offset functions
// returns the position of the offset entry in the header, size 2 bytes uint16
// digram
// OffsetPos:    [4-11] -> [12-13]
//               pointer	offset
//               (8B)		(2B)

func offsetPos(b BNode, idx uint16) uint16 {
	utils.Assert(1 <= idx && idx <= b.nkeys(), "index out of range")
	return HEADER + 8*b.nkeys() + 2*(idx-1) // position of the offset entry in header
}

// returns the value of the offset i.e. the location of the kv-pair at given index in the node, size 2 bytes uint16
// digram
// Offset:    [4-11] -> [12-13]
//            pointer	offset
//            (8B)		(2B)

func (b BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}
	return binary.LittleEndian.Uint16(b.data[offsetPos(b, idx):]) // read the offset 4 -> 13
}

// update the value of the offset i.e. the location of the kv-pair at given index in the node, size 2 bytes uint16
// digram
// Offset:    [4-11] -> [12-13]
//            pointer	offset
//            (8B)		(2B)

func (b BNode) setOffset(idx uint16, val uint16) {
	binary.LittleEndian.PutUint16(b.data[offsetPos(b, idx):], val) // write the offset 4 -> 13
}

// Key-value functions

// returns the position of the key-value pair in the node, size 2 bytes uint16
// digram
// KeyValuePos:[0-3] -> [4-11] -> [12-13] -> [14-17] -> [18...]
//             HEADER	pointer	offset	klen+vlen	key
//              (4B)		(8B)		(2B)		(4B)   (up to 1000B)

func kvPos(b BNode, idx uint16) uint16 {
	utils.Assert(idx <= b.nkeys(), "index out of range")
		return  HEADER + 8*b.nkeys() + 2*b.nkeys() + b.getOffset(idx) // position of the key-value pair in the node
}

// returns the key of the key-value pair at given index in the node, size 4 bytes uint32
// digram
// Key:    [0-3] -> [4-11] -> [12-13] -> [14-17] -> [18...]
//         HEADER	pointer	offset	klen+vlen	key
//          (4B)		(8B)		(2B)		(4B)   (up to 1000B)

func (b BNode) getKey(idx uint16) []byte {
	utils.Assert(idx < b.nkeys(), "index out of range")  // check the index
	pos := kvPos(b, idx) // position of the key-value pair in the node
	klen := binary.LittleEndian.Uint16(b.data[pos:]) // read the key length 4 -> 17
	return b.data[pos+4:][:klen] // read the key 4 -> 17
}
