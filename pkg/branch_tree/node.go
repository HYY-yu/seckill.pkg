package branch_tree

type Node interface {
	ID() string
}

// DecisionNode 决策节点
type DecisionNode interface {
	Node
	// DecisionRule 决策规则
	// return true 代表进入下一层
	// return false 代表跳过此节点
	DecisionRule(param interface{}) bool
	ChildrenList() []string
}

// ExecNode 叶子节点
type ExecNode interface {
	Node
	Do(param interface{}) error
}

// 根节点
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
