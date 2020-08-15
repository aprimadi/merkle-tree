package merkle

import (
  "reflect"
  "testing"
)

func TestIteration(t *testing.T) {
  min := []byte{0}
  max := []byte{255}
  left := []byte{255}
  right := []byte{3}
  tree := NewMerkleTree(1, min, max, left, right)

  ranges := []Range{
    Range{Left: []byte{255}, Right: []byte{1}},
    Range{Left: []byte{1}, Right: []byte{3}},
  }
  c := tree.Cursor()
  i := 0
  for n := c.First(); n != nil; n = c.Next() {
    rng := n.Range
    if !reflect.DeepEqual(rng, ranges[i]) {
      t.Errorf("Expected: %v, got: %v", ranges[i], rng)
    }
    i += 1
  }
}
