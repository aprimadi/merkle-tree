package merkle

import (
  "bytes"
)

// Token range from (left, right]
type Range struct {
  Left  []byte
  Right []byte
}

func (r *Range) Equal(other Range) bool {
  return bytes.Compare(r.Left, other.Left) == 0 && bytes.Compare(r.Right, other.Right) == 0
}
