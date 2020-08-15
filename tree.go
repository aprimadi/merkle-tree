package merkle

import (
	"bytes"
	"math/big"

	"github.com/spaolacci/murmur3"
)

type NodeType uint8

const (
	NodeInternal NodeType = iota
	NodeLeaf
)

type MerkleTree struct {
	// Whether the hash of the internal node is filled
	filled bool

	// Min and max possible value in the hash ring
	min []byte
	max []byte

	// fullRange represents a partition in the hash ring
	fullRange Range

	// depth of the merkle tree
	depth uint8

	root *Node
}

func NewMerkleTree(depth uint8, min []byte, max []byte, left []byte, right []byte) *MerkleTree {
	if depth == 0 {
		panic("Depth must be >= 1")
	}

	fullRange := Range{Left: left, Right: right}
	t := &MerkleTree{depth: depth, min: min, max: max, fullRange: fullRange}

	// Initialize the tree
	root := new(Node)
	root.Range = t.fullRange
	t.buildTree(root, 0, depth)
	t.root = root

	return t
}

func (t *MerkleTree) buildTree(node *Node, depth uint8, maxdepth uint8) {
	if depth == maxdepth {
		node.typ = NodeLeaf
		return
	}

	midpoint := t.midpoint(node.Range)

	// Build left subtree
	node.lchild = new(Node)
	node.lchild.parent = node
	node.lchild.Range = Range{Left: node.Range.Left, Right: midpoint}
	t.buildTree(node.lchild, depth+1, maxdepth)

	// Build right subtree
	node.rchild = new(Node)
	node.rchild.parent = node
	node.rchild.Range = Range{Left: midpoint, Right: node.Range.Right}
	t.buildTree(node.rchild, depth+1, maxdepth)
}

// Return midpoint between left and right
func (t *MerkleTree) midpoint(rng Range) []byte {
	var min big.Int
	var max big.Int
	var lt big.Int
	var rt big.Int
	var m big.Int

	min.SetBytes(t.min)
	max.SetBytes(t.max)
	lt.SetBytes(rng.Left)
	rt.SetBytes(rng.Right)

	if lt.Cmp(&rt) < 1 {
		m.Add(&lt, &rt)
		m.Rsh(&m, 1)
	} else {
		// Loop around case
		// The size of the range is (max - left) + (right - min)
		// Thus, the midpoint is (left + size / 2) mod (max + 1)
		var sz big.Int
		sz.Add(&max, &rt)
		sz.Sub(&sz, &lt)
		sz.Sub(&sz, &min)
		sz.Rsh(&sz, 1)
		m.Add(&lt, &sz)
		if m.Cmp(&max) == 1 {
		}
		m.Sub(&m, &max)
	}

	// Convert m from big int to byte slice
	midpoint := m.Bytes()

	// Prefix pad bytes with zeros so it has the same number of bytes as r.Max
	if len(midpoint) < len(t.max) {
		pad := len(t.max) - len(midpoint)
		midpoint = append(make([]byte, pad, len(t.max)), midpoint...)
	}

	return midpoint
}

func (t *MerkleTree) Diff(other *MerkleTree) []Range {
	t.Fill()
  other.Fill()

	d := []Range{}
	diff(t.root, other.root, &d)
	return d
}

func (t *MerkleTree) Fill() {
  t.fillInnerHash()
}

type consistency uint8

const (
	FullyConsistent consistency = iota
	PartiallyInconsistent
	FullyInconsistent
)

// Add largest contiguous ranges to diff
func diff(node1 *Node, node2 *Node, result *[]Range) consistency {
	assert(node1.typ == node2.typ, "Node type mismatch")
	assert(node1.Range.Equal(node2.Range), "Node range doesn't match")

	if bytes.Compare(node1.hash, node2.hash) == 0 {
		return FullyConsistent
	}
  // Both are node leaf with different hash
	if node1.typ == NodeLeaf && node2.typ == NodeLeaf {
		return FullyInconsistent
	}

	// Possible values left: PartiallyInconsistent, FullyInconsistent
	// Check the child
	lcon := diff(node1.lchild, node2.lchild, result)
	rcon := diff(node1.rchild, node2.rchild, result)
	if lcon == FullyInconsistent && rcon == FullyInconsistent {
		// If this is a root node, append diff to result
		if node1.parent == nil {
			*result = append(*result, node1.Range)
		}
		return FullyInconsistent
	} else if rcon == FullyInconsistent {
		*result = append(*result, node1.rchild.Range)
		return PartiallyInconsistent
	} else if lcon == FullyInconsistent {
		*result = append(*result, node1.lchild.Range)
		return PartiallyInconsistent
	} else {
		return PartiallyInconsistent
	}
}

// Fill hash of internal nodes
func (t *MerkleTree) fillInnerHash() {
	if t.filled {
		return
	}

	fillInnerHash(t.root)
	t.filled = true
}

func fillInnerHash(n *Node) {
	if n.typ == NodeLeaf {
		return
	}

	// Populate hash of child first
	fillInnerHash(n.lchild)
	fillInnerHash(n.rchild)

	// Compute the hash of this node
	if n.lchild.hash == nil && n.rchild.hash == nil {
		n.hash = nil
	} else if n.lchild.hash == nil {
		n.hash = n.rchild.hash
	} else if n.rchild.hash == nil {
		n.hash = n.lchild.hash
	} else {
		h := murmur3.New64()
		h.Write(n.lchild.hash)
		h.Write(n.rchild.hash)
		n.hash = h.Sum(nil)
	}
}

func (t *MerkleTree) Cursor() *Cursor {
	return &Cursor{tree: t}
}

// Merkle tree representation that can be serialized over the network using
// any protocol: msgpack, protobuf, json, etc.
type MerkleTreeSerializable struct {
	Min       []byte
	Max       []byte
	FullRange Range
	Depth     uint8
	FlatTree  [][]byte
}

// Serialize using breadth-first representation
func Serialize(t *MerkleTree) *MerkleTreeSerializable {
	s := new(MerkleTreeSerializable)
	s.Min = t.min
	s.Max = t.max
	s.FullRange = t.fullRange
	s.Depth = t.depth
	s.FlatTree = make([][]byte, 0)

	// Flatten tree starting from root node
	q := make([]*Node, 0)
	q = append(q, t.root)
	for len(q) > 0 {
		// Pop queue and append to flat tree
		node := q[0]
		q = q[1:]

		s.FlatTree = append(s.FlatTree, node.hash)

		// Push child of the node into queue
		if node.typ == NodeInternal {
			q = append(q, node.lchild, node.rchild)
		}
	}

	return s
}

// Deserialize breadth-first representation of the tree
func Deserialize(s *MerkleTreeSerializable) *MerkleTree {
	t := &MerkleTree{
		min:       s.Min,
		max:       s.Max,
		fullRange: s.FullRange,
		depth:     s.Depth,
	}

	// Build the tree
	root := &Node{level: 0, Range: t.fullRange}
	q := []*Node{}
	q = append(q, root)
	for _, hash := range s.FlatTree {
		node := q[0]
		q = q[1:]

		node.hash = hash

		if node.level < t.depth {
			// Push lchild and rchild
			midpoint := t.midpoint(node.Range)
			node.lchild = &Node{
				parent: node,
				level: node.level + 1,
				Range: Range{Left: node.Range.Left, Right: midpoint},
			}
			node.rchild = &Node{
				parent: node,
				level: node.level + 1,
				Range: Range{Left: midpoint, Right: node.Range.Right},
			}
			q = append(q, node.lchild, node.rchild)
		} else {
			node.typ = NodeLeaf
		}
	}

	// Set root after building the tree
	t.root = root

	return t
}

type Node struct {
	typ NodeType

	hash []byte

	parent *Node
	lchild *Node
	rchild *Node

  level uint8

	// Represent the range of this node
	Range Range
}

func (n *Node) isLeaf() bool {
	return n.typ == NodeLeaf
}

// Check if the given key is within range
func (n *Node) inRange(key []byte) bool {
	var l big.Int
	var r big.Int
	var k big.Int

	l.SetBytes(n.Range.Left)
	r.SetBytes(n.Range.Right)
	k.SetBytes(key)

	// l < k <= r
	return (l.Cmp(&k) == -1 && k.Cmp(&r) < 1)
}

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
