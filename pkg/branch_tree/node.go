package branch_tree

type Node interface {
	ID() string
}

// DecisionNode 决策节点
type DecisionNode interface {
	Node
	// DecisionRule 决策规则
	// return true 代表进入本节点的子节点，不再遍历本节点的兄弟节点。
	// return false 代表跳过此节点，继续遍历本节点的兄弟节点。
	DecisionRule(param interface{}) bool
	// ChildrenList 子节点 ID 列表
	ChildrenList() []string
}

// ExecNode 叶子节点
type ExecNode interface {
	Node
	// Do 当决策走到最深一层的叶子节点，调用 Do ，并结束这次决策。
	Do(param interface{}) error
}

// 根节点
// 内部虚拟根节点，为整颗决策树配置默认根。
type rootDecisionNode struct {
	childList []Node
	childMap  map[string]int
}

func (r *rootDecisionNode) AddChildren(n Node) {
	if r.childList == nil {
		r.childList = make([]Node, 0)
	}
	if r.childMap == nil {
		r.childMap = make(map[string]int)
	}

	r.childList = append(r.childList, n)
	r.childMap[n.ID()] = 1
}

func (r *rootDecisionNode) HasChildren(id string) bool {
	_, ok := r.childMap[id]
	return ok
}

func (r *rootDecisionNode) DeleteChildren(id string) {
	if r.HasChildren(id) {
		delete(r.childMap, id)
		for i, v := range r.childList {
			if v.ID() == id {
				r.childList = append(r.childList[:i], r.childList[i+1:]...)
				return
			}
		}
	}
}

func (*rootDecisionNode) ID() string {
	return "ROOT_NODE"
}

func (*rootDecisionNode) DecisionRule(param interface{}) bool {
	return true
}

func (r *rootDecisionNode) ChildrenList() []string {
	result := make([]string, len(r.childList))
	for i, v := range r.childList {
		result[i] = v.ID()
	}
	return result
}
