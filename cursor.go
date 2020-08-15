package merkle

import (
	"fmt"
)

// Cursor is used to traverse the leaf node of a Merkle Tree
type Cursor struct {
	tree *MerkleTree
	pos  *Node // Current node position, always store a leaf node
}

// First find the first leaf node
func (c *Cursor) First() *Node {
	// Start from the root
	node := c.tree.root

	for {
		if node.isLeaf() {
			c.pos = node
			return node
		}

		// Always select the left node
		node = node.lchild
	}
}

// Seek find leaf node with range matching the given key
// A seek operation always start from the current position of the cursor, this
// is to take advantage of program that insert data by key order.
func (c *Cursor) Seek(key []byte) *Node {
	// Traverse up the tree until it finds a node with range that match the key
	node := c.pos
	if node == nil {
		node = c.tree.root
	}
	for !node.inRange(key) && node.parent != nil {
		node = node.parent
	}

	if !node.inRange(key) {
		panic(fmt.Sprintf("The given key is not in the merkle tree: %v", key))
	}

	// Traverse down the tree to find a leaf node with range that match the key
	for {
		if node.isLeaf() {
			if node.inRange(key) {
				c.pos = node
				return node
			}
			panic("Tree isn't build correctly")
		}

		switch {
		case node.lchild.inRange(key):
			node = node.lchild
		case node.rchild.inRange(key):
			node = node.rchild
		default:
			// No child is in range, return nil, should not reach here
			panic("Tree isn't build correctly")
		}
	}
}

// Next get to the next leaf node.
// return nil if the current node is the last node
func (c *Cursor) Next() *Node {
	node := c.pos
	if node == nil {
		return c.First()
	}
  assert(node.isLeaf(), "Invariant violation: pos doesn't store leaf node")

	// Traverse up a tree until we're no longer the last child
	for node.parent != nil && node.parent.rchild == node {
		node = node.parent
	}

  if node.parent == nil {
    // This node is the last child
    return nil
  }

	// Get sibling node
	node = node.parent.rchild

	// Then, always take the first child until we reach a leaf node
	for !node.isLeaf() {
		node = node.lchild
	}

  c.pos = node
  return node
}
