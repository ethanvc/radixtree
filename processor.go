package radixtree

type PlainProcessor struct{}

func (p PlainProcessor) SplitPattern(pattern string) []PatternNode {
	return []PatternNode{
		PatternNode{
			NodeVal: pattern,
		},
	}
}

func (p PlainProcessor) GetParam(node PatternNode, pattern string) Param {
	panic("never come here")
}
