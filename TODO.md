# the node sturcture

the node sturcture is include the following sturcture

# 1. Header
-  [x] Type (1 bytes): Indicates whether the node is a leaf node or an internal node.
- [x] nkeys (2 bytes): Represents the number of keys stored in the node.


# 2. Pointer (List of number of keys in the node * 8 bytes)
- [x] Present only in internal nodes.
- [x] Each pointer (8 bytes) corresponds to a child node. Internal nodes use these pointers to navigate through the tree structure.


# 3. Offset (List of the number of keys in the node * 2 bytes)
- [x] Each offset points to the location of the corresponding key-value pair within the key-values section.
- [x] The offset is relative to the start of the key-values section.
- [x] The first offset is always 0, as it points to the beginning of the key-values section.

# 4.key value(packed of key and value pairs)
- [x] Pairs of keys and values data .
- [x] The Klen(2 bytes): length of the key.
- [ ] The Vlen(2 bytes): length of the value.
- [ ] The key(klen bytes): The actual key data.
- [ ] The value(vlen bytes): The actual value data.
- [ ] These pairs are packed together in the key-values section without any separators.
