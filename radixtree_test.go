package radixtree

import "testing"

func TestRadixTree_Insert(t1 *testing.T) {
	var tree RadixTree[PlainProcessor, int]
	tree.MustInsert("/abc", 3)
	tree.MustInsert("/abc/def", 4)
}
