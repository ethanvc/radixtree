package radixtree

import (
	"bytes"
	"fmt"
)

type RadixTree[Processor PatternProcessor, Value any] struct {
	processor Processor
	node      *RadixNode[Value]
}

type RadixNode[Value any] struct {
	children    map[byte]*RadixNode[Value]
	paramChild  *RadixNode[Value]
	patternNode PatternNode
	val         Value
	valEnabled  bool
	pattern     string
}

type PatternNode struct {
	ParamType bool
	NodeVal   string
}

type Param struct {
	Key   string
	Value string
}

type Params []Param

type PatternProcessor interface {
	SplitPattern(p string) []PatternNode
	GetParam(node PatternNode, p string) Param
}

func (t *RadixTree[Processor, Value]) Insert(pattern string, val Value) error {
	n, err := t.insert(t.processor.SplitPattern(pattern))
	if err != nil {
		return err
	}
	n.pattern = pattern
	n.val = val
	n.valEnabled = true
	return nil
}

func (t *RadixTree[Processor, Value]) insert(patternNodes []PatternNode) (*RadixNode[Value], error) {
	if t.node == nil {
		head, last := patternNodesToRadixNodes[Value](patternNodes)
		t.node = head
		return last, nil
	}
	n := t.node
	var i int
	var patternNode PatternNode
	for i, patternNode = range patternNodes {
		if n.patternNode != patternNode {
			break
		}
		if i == len(patternNodes)-1 {
			if n.valEnabled {
				return nil, fmt.Errorf("pattern already exist: %s", getPattern(patternNodes))
			}
			return n, nil
		}
		nextPatternNode := patternNodes[i+1]
		if nextPatternNode.ParamType {
			if n.paramChild == nil {
				head, last := patternNodesToRadixNodes[Value](patternNodes[i+1:])
				n.paramChild = head
				return last, nil
			}
			n = n.paramChild
		} else {
			tmp := n.children[getFirstByte(nextPatternNode.NodeVal)]
			if tmp == nil {
				head, last := patternNodesToRadixNodes[Value](patternNodes[i+1:])
				if n.children == nil {
					n.children = make(map[byte]*RadixNode[Value])
				}
				n.children[getFirstByte(nextPatternNode.NodeVal)] = head
				return last, nil
			}
			n = tmp
		}
	}
	if n.patternNode.ParamType {
		if patternNode.ParamType {
			return nil, fmt.Errorf("param placeholder conflict, insert pattern is %s, exist placeholder is %s",
				getPattern(patternNodes), n.patternNode.NodeVal)
		} else {
			panic("never come here")
		}
	} else {
		if patternNode.ParamType {
			panic("never come here")
		} else {
			prefix := longestCommonPrefix(n.patternNode.NodeVal, patternNode.NodeVal)
			if prefix != n.patternNode.NodeVal {
				childOld := *n
				n.children = make(map[byte]*RadixNode[Value])
				n.paramChild = nil
				n.patternNode.NodeVal = prefix
				var defaultVal Value
				n.val = defaultVal
				n.valEnabled = false
				n.pattern = ""
				n.children[getFirstByte(childOld.patternNode.NodeVal)] = &childOld
			}
			patternNodes[i].NodeVal = patternNodes[i].NodeVal[len(prefix):]
			head, last := patternNodesToRadixNodes[Value](patternNodes[i:])
			if n.children == nil {
				n.children = make(map[byte]*RadixNode[Value])
			}
			n.children[getFirstByte(head.patternNode.NodeVal)] = head
			return last, nil
		}
	}
}

func (t *RadixTree[Processor, Value]) MustInsert(pattern string, val Value) {
	err := t.Insert(pattern, val)
	if err != nil {
		panic(err)
	}
}

func getPattern(patternNodes []PatternNode) string {
	buf := bytes.NewBuffer(nil)
	for _, patternNode := range patternNodes {
		buf.WriteString(patternNode.NodeVal)
	}
	return buf.String()
}

func patternNodesToRadixNodes[Value any](patternNodes []PatternNode) (head, last *RadixNode[Value]) {
	for _, patternNode := range patternNodes {
		if head == nil {
			head = &RadixNode[Value]{
				patternNode: patternNode,
			}
			last = head
			continue
		}
		if patternNode.ParamType {
			last.paramChild = &RadixNode[Value]{
				patternNode: patternNode,
			}
			last = last.paramChild
		} else {
			last.children = make(map[byte]*RadixNode[Value])
			newChild := &RadixNode[Value]{
				patternNode: patternNode,
			}
			last.children[patternNode.NodeVal[0]] = newChild
			last = newChild
		}
	}
	return
}

func (t *RadixTree[Processor, Value]) Search(p string, params Params) (*RadixNode[Value], Params, error) {
	n := t.node
	var traceInfo traceBackInfo[Value]
	traceInfo.Params = params
	for {
		if n == nil {
			if n = traceInfo.Pop(); n == nil {
				return nil, params, nil
			}
		}
		if n.patternNode.ParamType {
			param := t.processor.GetParam(n.patternNode, p)
			traceInfo.Params = append(traceInfo.Params, param)
			p = p[len(param.Value):]
		} else {
			nodeLen := len(n.patternNode.NodeVal)
			if nodeLen > len(p) {
				n = traceInfo.Pop()
				if n == nil {
					return nil, params, nil
				}
				continue
			}
			if n.patternNode.NodeVal != p[0:nodeLen] {
				n = traceInfo.Pop()
				if n == nil {
					return nil, params, nil
				}
				continue
			}
			p = p[nodeLen:]
		}
		if p == "" {
			if n.valEnabled {
				return n, traceInfo.Params, nil
			} else {
				n = traceInfo.Pop()
				if n == nil {
					return nil, params, nil
				}
				continue
			}
		}

		if traceInfo.MatchParamNode {
			n = n.paramChild
			traceInfo.MatchParamNode = false
		} else {
			traceInfo.Push(n)
			n = n.children[p[0]]
		}
	}
}

func getFirstByte(s string) byte {
	if s == "" {
		return 0
	}
	return s[0]
}

func longestCommonPrefix(s1, s2 string) string {
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			return s1[:i] // 返回从0到i-1的子串
		}
	}

	return s1[:minLen]
}

type backTraceNode[Value any] struct {
	ParamSize int
	Node      *RadixNode[Value]
}

type backTraceNodes[Value any] []backTraceNode[Value]

func (t backTraceNodes[Value]) Push(n *RadixNode[Value], size int) backTraceNodes[Value] {
	return append(t, backTraceNode[Value]{
		ParamSize: size,
		Node:      n,
	})
}

func (t backTraceNodes[Value]) Pop(params Params) (backTraceNodes[Value], *RadixNode[Value], Params) {
	if len(t) == 0 {
		return nil, nil, params
	}
	lastNode := t[len(t)-1]
	return t[0 : len(t)-1], lastNode.Node, params[0:lastNode.ParamSize]
}

type traceBackInfo[Value any] struct {
	MatchParamNode bool
	Params         Params
	TraceNodes     backTraceNodes[Value]
}

func (info *traceBackInfo[Value]) Push(n *RadixNode[Value]) {
	info.TraceNodes = append(info.TraceNodes, backTraceNode[Value]{
		ParamSize: len(info.Params),
		Node:      n,
	})
}

func (info *traceBackInfo[Value]) Pop() *RadixNode[Value] {
	if len(info.TraceNodes) == 0 {
		return nil
	}
	info.MatchParamNode = true
	last := info.TraceNodes[len(info.TraceNodes)-1]
	info.Params = info.Params[0:last.ParamSize]
	info.TraceNodes = info.TraceNodes[0 : len(info.TraceNodes)-1]
	return last.Node
}
