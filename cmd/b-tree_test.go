package main

import (
	"encoding/binary"
	"testing"

	"github.com/abode1234/golang/key-value/utils-go"
)

func TestGetKeyMultiple(t *testing.T) {
	key0, val0 := "aa", "1"
	key1, val1 := "bb", "22"
	key2, val2 := "cc", "333"

	b := BNode{data: make([]byte, BTREE_PAGE_SIZE)}
	b.setHeader(BNODE_LEAF, 3) // nkeys = 3, btype = 2 (LEAF)

	// نحسب المواقع الصحيحة يدوياً (ground truth) بدون الاعتماد على kvPos المشكوك فيها
	correctBase := HEADER + 8*3 + 2*3 // = 34, هذا الحساب الصحيح المتفق عليه
	
	kv0len := 4 + len(key0) + len(val0)
	kv1len := 4 + len(key1) + len(val1)

	// نكتب البيانات يدوياً بالمواقع الصحيحة
	pos0 := correctBase
	binary.LittleEndian.PutUint16(b.data[pos0:], uint16(len(key0)))
	binary.LittleEndian.PutUint16(b.data[pos0+2:], uint16(len(val0)))
	copy(b.data[pos0+4:], key0)
	copy(b.data[pos0+4+len(key0):], val0)

	pos1 := correctBase + kv0len
	binary.LittleEndian.PutUint16(b.data[pos1:], uint16(len(key1)))
	binary.LittleEndian.PutUint16(b.data[pos1+2:], uint16(len(val1)))
	copy(b.data[pos1+4:], key1)
	copy(b.data[pos1+4+len(key1):], val1)

	pos2 := correctBase + kv0len + kv1len
	binary.LittleEndian.PutUint16(b.data[pos2:], uint16(len(key2)))
	binary.LittleEndian.PutUint16(b.data[pos2+2:], uint16(len(val2)))
	copy(b.data[pos2+4:], key2)
	copy(b.data[pos2+4+len(key2):], val2)

	// نخزن الـ offsets الصحيحة (بالنسبة لبداية منطقة الـ KV)
	b.setOffset(1, uint16(kv0len))
	b.setOffset(2, uint16(kv0len+kv1len))

	// الآن نطلب المفتاح رقم 0 عبر getKey العادية (اللي تستخدم kvPos بداخلها)
	got := string(b.getKey(0))
	utils.Assert(got == key0, "expected "+key0+" but got "+got)
}
func TestGetKey(t *testing.T) {
	key := "key1"
	val := "val1"

	b := BNode{data: make([]byte, BTREE_PAGE_SIZE)}
	b.setHeader(BNODE_LEAF, 1)
	b.setPtr(0, 1)

	pos := kvPos(b, 0)  // ← idx=0 بدل 1
	binary.LittleEndian.PutUint16(b.data[pos:], uint16(len(key)))
	binary.LittleEndian.PutUint16(b.data[pos+2:], uint16(len(val)))
	copy(b.data[pos+4:], key)
	copy(b.data[pos+4+uint16(len(key)):], val)
	utils.Assert(string(b.getKey(0)) == key, "getKey failed")  // ← idx=0
}
