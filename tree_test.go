package merkle

import (
  "bytes"
  "reflect"
  "testing"
)

func TestDiff(t *testing.T) {
  min := []byte{0}
  max := []byte{255}
  left := []byte{16}
  right := []byte{31}
  tree1 := NewMerkleTree(2, min, max, left, right)
  tree2 := NewMerkleTree(2, min, max, left, right)

  var node *Node

  c1 := tree1.Cursor()
  node = c1.Seek([]byte{17})
  node.hash = []byte("test")
  node = c1.Seek([]byte{31})
  node.hash = []byte("foo")

  c2 := tree2.Cursor()
  node = c2.Seek([]byte{17})
  node.hash = []byte("test")
  node = c2.Seek([]byte{31})
  node.hash = []byte("bar")

  diff := tree1.Diff(tree2)
  if len(diff) != 1 {
    t.Errorf("Diff length differs. Expect %v, got %v", 1, len(diff))
  }
  r := Range{Left: []byte{27}, Right: []byte{31}}
  if !diff[0].Equal(r) {
    t.Errorf("Expected: %v, got: %v", r, diff[0])
  }
}

func TestSerialize(t *testing.T) {
  min := []byte{0}
  max := []byte{255}
  left := []byte{16}
  right := []byte{31}
  tree := NewMerkleTree(2, min, max, left, right)

  var node *Node
  c := tree.Cursor()
  node = c.Seek([]byte{17})
  node.hash = []byte("test")
  node = c.Seek([]byte{31})
  node.hash = []byte("foo")

  tree.Fill()

  expected := &MerkleTreeSerializable{
    Min: min,
    Max: max,
    FullRange: Range{Left: left, Right: right},
    Depth: 2,
    FlatTree: [][]byte{
      []byte{43, 75, 75, 233, 233, 220, 128, 100},
      []byte{116, 101, 115, 116},
      []byte{102, 111, 111},
      []byte{116, 101, 115, 116},
      []byte{},
      []byte{},
      []byte{102, 111, 111},
    },
  }
  got := Serialize(tree)
  if (bytes.Compare(expected.Min, got.Min) != 0 ||
      bytes.Compare(expected.Max, got.Max) != 0 ||
      !reflect.DeepEqual(expected.FullRange, got.FullRange) ||
      expected.Depth != got.Depth) {
    t.Errorf(
      "Expected: (Min: %v, Max: %v, FullRange: %v, Depth: %v), got: (Min: %v, Max: %v, FullRange: %v, Depth: %v)",
      expected.Min, expected.Max, expected.FullRange, expected.Depth,
      got.Min, got.Max, got.FullRange, got.Depth,
    )
  }
  for i := range expected.FlatTree {
    if bytes.Compare(expected.FlatTree[i], got.FlatTree[i]) != 0 {
      t.Errorf("FlatTree(i) expected: %v, got: %v", expected.FlatTree[i], got.FlatTree[i])
    }
  }
}

func TestDeserialize(t *testing.T) {
  // Expected tree
  min := []byte{0}
  max := []byte{255}
  left := []byte{16}
  right := []byte{31}
  tree := NewMerkleTree(2, min, max, left, right)

  var node *Node
  c := tree.Cursor()
  node = c.Seek([]byte{17})
  node.hash = []byte("test")
  node = c.Seek([]byte{31})
  node.hash = []byte("foo")

  s := &MerkleTreeSerializable{
    Min: min,
    Max: max,
    FullRange: Range{Left: left, Right: right},
    Depth: 2,
    FlatTree: [][]byte{
      []byte{43, 75, 75, 233, 233, 220, 128, 100},
      []byte{116, 101, 115, 116},
      []byte{102, 111, 111},
      []byte{116, 101, 115, 116},
      []byte{},
      []byte{},
      []byte{102, 111, 111},
    },
  }

  tree2 := Deserialize(s)
  if len(tree2.Diff(tree)) != 0 {
    t.Errorf("Deserialized tree is not equal to the original tree")
  }
}
