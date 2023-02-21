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

// AddNode 将节点添加至决策树，自动识别节点位置。
func (t *DecisionTree) AddNode(n Node) error {
	if t.root == nil {
		t.root = &rootDecisionNode{}
	}
	if t.nodeMap == nil {
		t.nodeMap = make(map[string]Node)
		t.nodeMap[t.root.ID()] = t.root
	}
	switch nn := n.(type) {
	case DecisionNode:
		for _, v := range nn.ChildrenList() {
			if t.root.HasChildren(v) {
				t.root.DeleteChildren(v)
			}
		}
	case ExecNode:
	case Node:
		return fmt.Errorf("must set Node to DecisionNode or ExecNode")
	default:
		return fmt.Errorf("not support type: %v", reflect.TypeOf(n))
	}

	if !t.dfsTree(t.root, t.root, n) {
		t.root.AddChildren(n)
	}
	t.nodeMap[n.ID()] = n
	return nil
}

func (t *DecisionTree) PrintTree() [][]string {
	return t.bfsTreePrint(t.root)
}

func (t *DecisionTree) bfsTreePrint(root Node) [][]string {
	result := make([][]string, 0)
	p := root
	queue := list.New()
	queue.PushBack(p.ID())
	L := 0

	for queue.Len() != 0 {
		r := make([]string, 0)
		// 处理当前层
		size := queue.Len()
		for i := 0; i < size; i++ {
			e := queue.Front()
			pid, _ := e.Value.(string)
			p, ok := t.nodeMap[pid]
			if !ok {
				continue
			}
			r = append(r, p.ID())
			// DEBUG
			fmt.Printf("LEVEL: %d ,id:%s \n", L, p.ID())
			switch p := t.nodeMap[pid].(type) {
			case DecisionNode:
				fmt.Printf(" It has child: %v \n", p.ChildrenList())
				for _, child := range p.ChildrenList() {
					queue.PushBack(child)
				}
			case ExecNode:
				//queue.PushBack(p.ID())
			}
			queue.Remove(e)
		}
		L++
		result = append(result, r)
	}
	return result
}

func (t *DecisionTree) dfsTree(root DecisionNode, n DecisionNode, p Node) bool {
	if len(root.ChildrenList()) == 0 {
		// p 节点肯定不存在于树的某个节点中，因为树中没有节点（除根节点）
		// 此节点暂时添加到 root
		return false
	}
	cl := n.ChildrenList()
	for _, v := range cl {
		if v == p.ID() {
			// 尝试删除 root 节点子节点列表中暂存的节点（因为此节点已找到它的真正父节点）
			if t.root.HasChildren(v) && n != root {
				t.root.DeleteChildren(v)
			}
			return true
		}
		_, ok := t.nodeMap[v]
		if !ok {
			continue
		}
		switch node := t.nodeMap[v].(type) {
		case DecisionNode:
			ok := t.dfsTree(root, node, p)
			if ok {
				return ok
			}
		case ExecNode:
			continue
		}
	}
	return false
}

// Decision 开始决策
func (t *DecisionTree) Decision(param interface{}) error {
	if len(t.root.childList) == 0 {
		return fmt.Errorf("not have node")
	}

	return t.levelOrderTree(param, t.root)
}

func (t *DecisionTree) levelOrderTree(param interface{}, root Node) error {
	p := root
	queue := list.New()
	queue.PushBack(p.ID())

	for queue.Len() != 0 {
		// 处理当前层
		size := queue.Len()
		for i := 0; i < size; i++ {
			e := queue.Front()
			pid, _ := e.Value.(string)
			switch p := t.nodeMap[pid].(type) {
			case DecisionNode:
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
			case ExecNode:
				return p.Do(param)
			}
			queue.Remove(e)
		}
	}
	return nil
}
