package radixtree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRadixTree_Insert(t *testing.T) {
	var tree RadixTree[PlainProcessor, int]
	tree.MustInsert("/abc", 3)
	tree.MustInsert("/abc/def", 4)
}

func Test_DuplicateInsert(t1 *testing.T) {
	var tree RadixTree[PlainProcessor, int]
	err := tree.Insert("/abc", 3)
	require.NoError(t1, err)
	err = tree.Insert("/abc", 4)
	require.EqualError(t1, err, "pattern already exist: /abc")
}

func TestRadixTree_1(t *testing.T) {
	var tree RadixTree[GinPatternProcessor, int]
	tree.MustInsert("/abc/:id", 3)
	tree.MustInsert("/abc/bcd", 4)
	n, params, err := tree.Search("/abc/bcde", nil)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(params))
	require.Equal(t, "", n.pattern)
}
