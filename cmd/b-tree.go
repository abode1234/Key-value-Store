package main

import "encoding/binary"

type BNode struct {
	data []byte
}

const (
	BNODE_NODE = 1
	BNODE_LEAF = 2
)

// The page size is defined to be 4K bytes. A larger page size such as 8K or 16K also works.
// We also add some constraints on the size of the keys and values. So that a node with a
// single KV pair always fits on a single page. If you need to support bigger keys or bigger
// values, you have to allocate extra pages for them and that adds complexity.

type BTree struct {
	root uint64
	get  func(uint64) BNode
	new  func(BNode) uint64
	del  func(uint64)
}

const HEADER_SIZE = 4
const BTREE_PAGE_SIZE = 4096
const BTREE_MAX_KEY_SIZE = 1000
const BTREE_MAX_VALUE_SIZE = 1000

func assert(b bool) {
	panic("unimplemented")
}

func init() {
	node1max := HEADER_SIZE + 8 + 2 + 4 + BTREE_MAX_KEY_SIZE + BTREE_MAX_VALUE_SIZE
	assert(node1max <= BTREE_PAGE_SIZE)
}

//Since a node is just an array of bytes, we’ll add some helper functions to access its contents.


// Heder

func (node BNode) btype() uint16 {
	return binary.LittleEndian.Uint16(node.data)
}

func (node BNode) nkeys() uint16{
	return binary.LittleEndian.Uint16(node.data[2:4])
}

func (node BNode) setHeader(btype uint16, nkeys uint16) {
	binary.LittleEndian.PutUint16(node.data[0:2], btype)
	binary.LittleEndian.PutUint16(node.data[2:4], nkeys)
}

// Pointers

func (node BNode) getPtr(idx uint16) uint64{
	assert(idx < node.nkeys())
	pos := HEADER_SIZE + 2 + idx*8
	return binary.LittleEndian.Uint64(node.data[pos:])
}

func (node BNode) setPtr(idx uint16, val uint64) {
	assert(idx < node.nkeys())
	pos := HEADER_SIZE + 8*idx
	binary.LittleEndian.PutUint64(node.data[pos:], val)
}

// • The offset is relative to the position of the first KV pair.
// • The offset of the first KV pair is always zero, so it is not stored in the list.
// • We store the offset to the end of the last KV pair in the offset list, which is used to
// determine the size of the node


func offsetPos(node BNode, idx uint16) uint16 {
	assert(1 <= idx && idx <= node.nkeys())
	return HEADER_SIZE + 8*node.nkeys() + 2*(idx-1)
}


func (node BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}
	return binary.LittleEndian.Uint16(node.data[offsetPos(node, idx):])
}

func (node BNode) setOffset(idx uint16, offset uint16) {
	binary.LittleEndian.PutUint16(node.data[offsetPos(node, idx):], offset)
}
