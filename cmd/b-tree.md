# B-Tree node encoding

A `BNode` is not a struct with fields — it's a thin wrapper around a raw byte
slice (`b-tree.go:8-10`):

```go
type BNode struct {
	data []byte
}
```

Everything else in `b-tree.go` is just functions that read/write specific
byte ranges of `data` by hand, using `encoding/binary`. There's no
serialization step: the in-memory representation of a node *is* its on-disk
page layout, so a `BNode` can be written straight to disk and read straight
back.

## Page layout

Each node lives in one fixed-size page (`BTREE_PAGE_SIZE = 4096` bytes,
`b-tree.go:31`) and is laid out like this:

```
[0-3]     [4 .. 4+8n)      [4+8n .. 4+10n)   [kv region...]
HEADER    pointers          offsets           key-value pairs
(4B)      (8B * nkeys)      (2B * nkeys)
```

- **HEADER (4B)** — 2 bytes node type + 2 bytes key count (`setHeader`,
  `btype`, `nkeys`).
- **pointers (8B × nkeys)** — one `uint64` disk page number per key, used
  only on internal nodes to point at child nodes (`getPtr`/`setPtr`).
- **offsets (2B × nkeys)** — cumulative byte offsets into the KV region,
  used to jump straight to the *n*-th key-value pair without scanning
  every entry before it (`offsetPos`/`getOffset`/`setOffset`).
- **KV region** — each entry is `klen(2B) + vlen(2B) + key + value`
  (`kvPos`/`getKey`).

`init()` (`b-tree.go:49-52`) asserts at startup that one KV entry, at its
absolute worst case (`BTREE_MAX_KEYS_SIZE=1000` + `BTREE_MAX_VALUES_SIZE=3000`
bytes), still fits inside a single 4096-byte page alongside the header,
one pointer, and one offset.

## Reading the header

```go
func (b BNode) btype() uint16 { return binary.LittleEndian.Uint16(b.data) }
func (b BNode) nkeys() uint16 { return binary.LittleEndian.Uint16(b.data[2:4]) }
func (b BNode) setHeader(btype, nkeys uint16) { ... }
```

`btype` reads bytes `[0:2]`, `nkeys` reads bytes `[2:4]`. Every other
function in the file calls `b.nkeys()` to know how far the pointer and
offset arrays extend — the header has to be set (`setHeader`) before
anything else touches the node.

## Pointers — 0-indexed

```go
func (b BNode) getPtr(idx uint16) uint64 {
	utils.Assert(idx < b.nkeys(), "index out of range")
	return binary.LittleEndian.Uint64(b.data[HEADER+idx*8:])
}
```

`idx` ranges over `0 .. nkeys-1`, one pointer per key, stored right after
the 4-byte header. `setPtr` mirrors this for writes.

## Offsets — 1-indexed, boundary-style

```go
func offsetPos(b BNode, idx uint16) uint16 {
	utils.Assert(1 <= idx && idx <= b.nkeys(), "index out of range")
	return HEADER + 8*b.nkeys() + 2*(idx-1)
}

func (b BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}
	return binary.LittleEndian.Uint16(b.data[offsetPos(b, idx):])
}
```

This is the part that trips people up: offsets are indexed `1 .. nkeys`,
*not* `0 .. nkeys-1`. `getOffset(idx)` means "how many bytes into the KV
region does key `idx` start" — so `getOffset(0)` is always `0` (the first
key starts right at the beginning of the KV region) and is never actually
stored; it's hardcoded. Offsets for `idx = 1..nkeys` are the ones physically
written into the offset array, each one `(idx-1)` slots into that array.

## Key-value pairs — 0-indexed again

```go
func kvPos(b BNode, idx uint16) uint16 {
	utils.Assert(idx <= b.nkeys(), "index out of range")
	return HEADER + 8*b.nkeys() + 2*b.nkeys() + b.getOffset(idx)
}

func (b BNode) getKey(idx uint16) []byte {
	utils.Assert(idx < b.nkeys(), "index out of range")
	pos := kvPos(b, idx)
	klen := binary.LittleEndian.Uint16(b.data[pos:])
	return b.data[pos+4:][:klen]
}
```

`kvPos(idx)` = end of the offset array (`HEADER + 8*nkeys + 2*nkeys`) +
`getOffset(idx)`. Because `getOffset(0) == 0`, `kvPos(0)` always lands
exactly at the start of the KV region — no lookup needed for the first
key. `kvPos` itself is allowed one index past the last key (`idx <= nkeys`,
not `idx < nkeys`) because it's also used internally to find *where the KV
region ends* (`kvPos(nkeys)` = end-of-data, needed when appending a new
key or computing the node's total size). `getKey`, on the other hand, only
makes sense for an index that actually holds a key, so it asserts the
strict `idx < nkeys`.

At `pos`, the layout is `klen(2B) | vlen(2B) | key bytes | value bytes`.
`getKey` reads `klen` from `pos`, then slices the key out starting at
`pos+4` (skipping over both the 2-byte `klen` and 2-byte `vlen` fields).
There's no `getVal` yet — that's the obvious next function to add,
symmetric to `getKey` but starting at `pos+4+klen`.

## Worked example (with numbers)

Take a leaf node with **3 keys**: `"a"→"1"`, `"bb"→"22"`, `"ccc"→"333"`.

`nkeys = 3`, `HEADER = 4`.

**1. Array sizes**

```
pointers = 8 * nkeys = 8 * 3 = 24 bytes
offsets  = 2 * nkeys = 2 * 3 = 6 bytes
```

So the KV region starts at:

```
HEADER + pointers + offsets = 4 + 24 + 6 = 34
```

That `34` is exactly the constant term in `kvPos`: `HEADER + 8*nkeys + 2*nkeys`.

**2. Size of each KV entry** (`4B` for `klen+vlen`, then the raw bytes)

```
entry 0: "a"  →"1"    → 4 + 1 + 1 = 6 bytes
entry 1: "bb" →"22"   → 4 + 2 + 2 = 8 bytes
entry 2: "ccc"→"333"  → 4 + 3 + 3 = 10 bytes
```

**3. Offsets — cumulative, 1-indexed**

`getOffset(idx)` is "how many bytes into the KV region has entry `idx` *started*",
i.e. the running total of every entry *before* it:

```
getOffset(0) = 0                          (hardcoded, entry 0 starts at the region's start)
getOffset(1) = size(entry 0) = 6          (entry 1 starts after entry 0)
getOffset(2) = size(entry 0)+size(entry 1) = 6+8 = 14
getOffset(3) = size(entry 0..2)           = 6+8+10 = 24   (= end of the KV region)
```

Only `getOffset(1)`, `getOffset(2)`, `getOffset(3)` are physically stored
(3 slots, one per key beyond the implicit 0th) — that matches `1 <= idx <= nkeys`
in `offsetPos`. Where they live in the offset array:

```
offsetPos(1) = HEADER + 8*nkeys + 2*(1-1) = 4 + 24 + 0  = 28   // slot 0
offsetPos(2) = HEADER + 8*nkeys + 2*(2-1) = 4 + 24 + 2  = 30   // slot 1
offsetPos(3) = HEADER + 8*nkeys + 2*(3-1) = 4 + 24 + 4  = 32   // slot 2
```

(`28, 30, 32` — 3 consecutive `uint16` slots, right before the KV region at
byte `34`. Checks out: `32 + 2 = 34`.)

**4. Absolute KV positions**

```
kvPos(0) = 34 + getOffset(0) = 34 + 0  = 34   →  entry "a"→"1"    lives at bytes [34, 40)
kvPos(1) = 34 + getOffset(1) = 34 + 6  = 40   →  entry "bb"→"22"  lives at bytes [40, 48)
kvPos(2) = 34 + getOffset(2) = 34 + 14 = 48   →  entry "ccc"→"333" lives at bytes [48, 58)
kvPos(3) = 34 + getOffset(3) = 34 + 24 = 58   →  end of the KV region (no entry here — this
                                                   is the "idx == nkeys" case kvPos allows
                                                   but getKey rejects)
```

**5. Reading `getKey(1)`** ("bb"), step by step:

```
pos   := kvPos(b, 1)                     = 40
klen  := data[40:42] as uint16           = 2
key   := data[40+4 : 40+4+2] = data[44:46] = "bb"
```

That's why `getKey` reads `pos+4:` for the key — it's skipping the 2-byte
`klen` and 2-byte `vlen` fields that sit at `[pos:pos+4)`.

## Workflow end-to-end

To read the *i*-th key out of a raw page:

1. `b.nkeys()` — how many keys are in this node (from the header).
2. `b.getOffset(idx)` — how far into the KV region key `idx` starts
   (`0` if `idx == 0`, otherwise a read from the offset array).
3. `kvPos(b, idx)` — turn that into an absolute byte position by adding
   past the header, pointer array, and offset array.
4. `b.getKey(idx)` — at that position, read `klen` and slice out the key
   bytes.

Building a node (as in `b-tree_test.go`) runs the same steps in reverse:
`setHeader` first (so `nkeys()` is correct for everything else), then
`setPtr`/`setOffset` for each slot, then write `klen`/`vlen`/key/value
bytes directly at `kvPos(idx)`.

## What's still missing

- `getVal` / `setOffset` writers for a full insert path.
- `BTree.get`/`new`/`del` (`b-tree.go:22-24`) are declared as callback
  fields but nothing wires them up to real page storage yet — this is
  where the node would actually get read from and written to disk.
- No split/merge logic — currently this file only knows how to lay out
  and read a single node's bytes, not how a tree of nodes grows.
