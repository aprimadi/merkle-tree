merkle-tree
===========

Implementation of Merkle Tree for anti entropy data synchronization protocol in Golang. The focus of this library is in providing a merkle tree implementation that is serializable/deserializable over the network and quickly computing the range of difference between two Merkle tree.

## Background

Amazon Dynamo uses Merkle Tree for anti entropy replica synchronization in the case of permanent failures as described in the paper: [LINK TO PAPER HERE]. However, the paper provides very minimal detail. Thus, this project was developed to gain insight into the detail of such protocol, as well as providing building block for building highly available, fault tolerant distributed systems in Go. This project draw heavy inspiration from the Apache Cassandra Merkle Tree implementation.

## Hash Function

As the main purpose of this library is to use Merkle Tree for data synchronization under non-Byzantine environment, it uses Murmur3 hash function by default which is a non-cryptographic hash function. Murmur3 should not be used for verifying the authenticity of data such as in Blockchain. Its advantage compared to cryptographic hash function is that it's reversable and can be computed very quickly.

## Usage

```go
package main

import (
  "github.com/aprimadi/merkle-tree"
)

func main() {
  // Create a merkle tree with depth 2
  min := []byte{0}
  max := []byte{255}
  left := []byte{255}
  right := []byte{15}
  tree := merkle.NewMerkleTree(2, min, max, left, right)

  // Iterate to the tree leaf nodes
  c := tree.Cursor()
  for n := c.First(); n != nil; n = c.Next() {
    range := n.Range

    // Collect data for a given range and compute its hash
    // ...

    n.hash = ...
  }

  // Deserialize a tree from the network
  // ...
  tree2 := merkle.Deserialize(ts)

  // Compare the tree with another tree deserialized from the network
  var rangeDiffs []Range
  rangeDiffs = tree.Diff(tree2)

  // Send data in the diff range to synchronize
  // ...
}
```
