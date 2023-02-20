package branch_tree

import (
	"container/list"
	"fmt"
	"reflect"
)

// DecisionTree 分支树(决策树)
type DecisionTree struct {
	root *rootDecisionNode

	nodeMap map[string]Node
}

func (t *DecisionTree) AddNode(n Node) error {
	if t.root == nil {
		t.root = &rootDecisionNode{}
	}
	if t.nodeMap == nil {
		t.nodeMap = make(map[string]Node)
	}
	switch n.(type) {
	case DecisionNode:
	case ExecNode:
	case Node:
		return fmt.Errorf("must set Node to DecisionNode or ExecNode")
	default:
		return fmt.Errorf("not support type: %v", reflect.TypeOf(n))
	}

	if decisionNode, ok := n.(DecisionNode); ok && len(t.root.childList) > 0 {
		for _, v := range decisionNode.ChildrenList() {
			if t.root.HasChildren(v) {
				t.root.DeleteChildren(v)
			}
		}
	}

	t.root.AddChildren(n)
	t.nodeMap[n.ID()] = n
	return nil
}

func (t *DecisionTree) Decision(param interface{}) error {
	if len(t.root.childList) == 0 {
		return fmt.Errorf("not have node")
	}

	return levelOrderTree(param, t.root)
}

func levelOrderTree(param interface{}, root Node) error {
	p := root
	queue := list.New()
	queue.PushBack(p)

	for queue.Len() != 0 {
		// 处理当前层
		size := queue.Len()
		for i := 0; i < size; i++ {
			e := queue.Front()
			p, ok := e.Value.(DecisionNode)
			if ok {
				enterNext := p.DecisionRule(param)
				if enterNext {
					// 本层遍历提前结束，后续节点不再访问
					if i < size-1 {
						for j := 0; j < size-1-i; j++ {
							queue.Remove(queue.Front())
						}
					}
					for _, child := range p.ChildrenList() {
						queue.PushBack(child)
					}
				}
			} else {
				if exec, ok := e.Value.(ExecNode); ok {
					return exec.Do(param)
				}
			}
			queue.Remove(e)
		}
	}
	return nil
}
