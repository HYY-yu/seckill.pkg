package branch_tree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testNode struct {
	Id int
}

func (t testNode) ID() string {
	return fmt.Sprintf("%d", t.Id)
}

func Test_rootDecisionNode_AddChildren(t *testing.T) {
	tests := []struct {
		name    string
		args    []Node
		wantIds []string
	}{
		{
			name: "TEST1",
			args: []Node{
				testNode{Id: 1},
				testNode{Id: 2},
				testNode{Id: 3},
				testNode{Id: 4},
			},
			wantIds: []string{"1", "2", "3", "4"},
		},
		{
			name:    "TEST2",
			args:    []Node{},
			wantIds: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &rootDecisionNode{}

			for i := range tt.args {
				r.AddChildren(tt.args[i])
			}
			ids := r.ChildrenList()

			assert.Equal(t, tt.wantIds, ids)
			assert.Equal(t, len(tt.wantIds), len(r.childList))
			assert.Equal(t, len(tt.wantIds), len(r.childMap))
		})
	}
}

func Test_rootDecisionNode_DeleteChildren(t *testing.T) {
	tests := []struct {
		name     string
		args     []Node
		deleteId string
		wantIds  []string
	}{
		{
			name: "TEST1",
			args: []Node{
				testNode{Id: 1},
				testNode{Id: 2},
				testNode{Id: 3},
				testNode{Id: 4},
			},
			deleteId: "1",
			wantIds:  []string{"2", "3", "4"},
		},
		{
			name: "TEST2",
			args: []Node{
				testNode{Id: 1},
				testNode{Id: 2},
				testNode{Id: 3},
				testNode{Id: 4},
			},
			deleteId: "2",
			wantIds:  []string{"1", "3", "4"},
		}, {
			name: "TEST3",
			args: []Node{
				testNode{Id: 1},
				testNode{Id: 2},
				testNode{Id: 3},
				testNode{Id: 4},
			},
			deleteId: "4",
			wantIds:  []string{"1", "2", "3"},
		}, {
			name: "TEST4",
			args: []Node{
				testNode{Id: 1},
				testNode{Id: 2},
				testNode{Id: 3},
				testNode{Id: 4},
			},
			deleteId: "0",
			wantIds:  []string{"1", "2", "3", "4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &rootDecisionNode{}

			for i := range tt.args {
				r.AddChildren(tt.args[i])
			}

			r.DeleteChildren(tt.deleteId)
			ids := r.ChildrenList()
			assert.Equal(t, tt.wantIds, ids)
		})
	}
}
