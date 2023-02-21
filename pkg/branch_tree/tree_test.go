package branch_tree

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type testDecisionNode struct {
	Id          int
	Rule        bool
	ChildrenArr []string
}

func (t testDecisionNode) ID() string {
	return fmt.Sprintf("%d", t.Id)
}

func (t testDecisionNode) DecisionRule(param interface{}) bool {
	return t.Rule
}

func (t testDecisionNode) ChildrenList() []string {
	return t.ChildrenArr
}

type testExecNode struct {
	Id          int
	DoSomeThing func(param interface{}) error
}

func (t testExecNode) ID() string {
	return fmt.Sprintf("%d", t.Id)
}

func (t testExecNode) Do(param interface{}) error {
	return t.DoSomeThing(param)
}

func TestDecisionTree_AddNode(t *testing.T) {
	tests := []struct {
		name     string
		args     []Node
		wantErrs []error
		wantIds  []string
	}{
		{
			name: "TEST_Wrong_Type",
			args: []Node{
				testDecisionNode{Id: 1, Rule: true, ChildrenArr: []string{"11", "12"}},
				testNode{Id: 2},
				testDecisionNode{Id: 11, Rule: false, ChildrenArr: []string{"111"}},
				testExecNode{Id: 21},
				testDecisionNode{Id: 2, Rule: false, ChildrenArr: []string{"21"}},
				testExecNode{Id: 111},
				testExecNode{Id: 12},
			},
			wantErrs: []error{nil, errors.New("must set Node to")},
			wantIds:  []string{"1", "2"},
		},
		{
			name:     "TEST_Empty",
			args:     []Node{},
			wantErrs: []error{},
			wantIds:  []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := &DecisionTree{root: &rootDecisionNode{}}

			for i := range tt.args {
				err := tree.AddNode(tt.args[i])

				if err != nil && !strings.HasPrefix(err.Error(), tt.wantErrs[i].Error()) {
					t.Errorf("error not wantErrs : %v - %v", err, tt.wantErrs[i])
				}
			}
			ids := tree.root.ChildrenList()
			assert.Equal(t, tt.wantIds, ids)

		})
	}
}

func TestDecisionTree_BuildTree(t *testing.T) {
	tests := []struct {
		name    string
		args    []Node
		wantIds [][]string
	}{
		{
			name: "TEST_BuildTree1",
			args: []Node{
				testDecisionNode{Id: 1, Rule: true, ChildrenArr: []string{"11", "12"}},
				testDecisionNode{Id: 11, Rule: false, ChildrenArr: []string{"111"}},
				testExecNode{Id: 21},
				testDecisionNode{Id: 2, Rule: false, ChildrenArr: []string{"21"}},
				testExecNode{Id: 111},
				testExecNode{Id: 12},
			},
			wantIds: [][]string{
				{"ROOT_NODE"},
				{"1", "2"},
				{"11", "12", "21"},
				{"111"},
			},
		},
		{
			name: "TEST_BuildTree2",
			args: []Node{
				testDecisionNode{Id: 1, Rule: true, ChildrenArr: []string{"11"}},
				testDecisionNode{Id: 11, Rule: false, ChildrenArr: []string{"111"}},
				testExecNode{Id: 21},
				testDecisionNode{Id: 2, Rule: false, ChildrenArr: []string{"21"}},
				testExecNode{Id: 111},
				testExecNode{Id: 12},
			},
			wantIds: [][]string{
				{"ROOT_NODE"},
				{"1", "2", "12"},
				{"11", "21"},
				{"111"},
			},
		},
		{
			name: "TEST_BuildTree2_1",
			args: []Node{
				testDecisionNode{Id: 1, Rule: true, ChildrenArr: []string{"11"}},
				testExecNode{Id: 21},
				testDecisionNode{Id: 2, Rule: false, ChildrenArr: []string{"21"}},
				testDecisionNode{Id: 11, Rule: false, ChildrenArr: []string{"111"}},
				testExecNode{Id: 111},
				testExecNode{Id: 12},
			},
			wantIds: [][]string{
				{"ROOT_NODE"},
				{"1", "2", "12"},
				{"11", "21"},
				{"111"},
			},
		},
		{
			name: "TEST_BuildTree3",
			args: []Node{
				testDecisionNode{Id: 1, Rule: true, ChildrenArr: []string{"11", "12", "21"}},
				testExecNode{Id: 21},
				testDecisionNode{Id: 11, Rule: false, ChildrenArr: []string{"111"}},
				testExecNode{Id: 111},
				testExecNode{Id: 12},
			},
			wantIds: [][]string{
				{"ROOT_NODE"},
				{"1"},
				{"11", "12", "21"},
				{"111"},
			},
		},
		{
			name: "TEST_BuildTree4",
			args: []Node{
				testDecisionNode{Id: 1, Rule: true, ChildrenArr: []string{"11"}},
				testDecisionNode{Id: 11, Rule: true, ChildrenArr: []string{"111"}},
				testDecisionNode{Id: 111, Rule: true, ChildrenArr: []string{"1111"}},
				testDecisionNode{Id: 1111, Rule: true, ChildrenArr: []string{"11111"}},
				testDecisionNode{Id: 11111, Rule: true, ChildrenArr: []string{"111111"}},
				testExecNode{Id: 111111},
			},
			wantIds: [][]string{
				{"ROOT_NODE"},
				{"1"},
				{"11"},
				{"111"},
				{"1111"},
				{"11111"},
				{"111111"},
			},
		},
		{
			name: "TEST_BuildTree5",
			args: []Node{
				testExecNode{Id: 11},
				testExecNode{Id: 21},
				testExecNode{Id: 22},
				testDecisionNode{Id: 1, Rule: true, ChildrenArr: []string{"11"}},
				testDecisionNode{Id: 2, Rule: true, ChildrenArr: []string{"21", "22"}},
			},
			wantIds: [][]string{
				{"ROOT_NODE"},
				{"1", "2"},
				{"11", "21", "22"},
			},
		},
		{
			name: "TEST_BuildTree6",
			args: []Node{
				testDecisionNode{Id: 1, Rule: false, ChildrenArr: []string{"11", "12"}},
				testDecisionNode{Id: 11, Rule: false, ChildrenArr: []string{"111"}},
				testDecisionNode{Id: 2, Rule: true, ChildrenArr: []string{"21", "22"}},
				testDecisionNode{Id: 21, Rule: false, ChildrenArr: []string{"211", "212"}},
				testDecisionNode{Id: 22, Rule: true, ChildrenArr: []string{"221", "222"}},
				testExecNode{Id: 211},
				testExecNode{Id: 212},
				testExecNode{Id: 111},
				testExecNode{Id: 221},
				testExecNode{Id: 222},
				testExecNode{Id: 12},
			},
			wantIds: [][]string{
				{"ROOT_NODE"},
				{"1", "2"},
				{"11", "12", "21", "22"},
				{"111", "211", "212", "221", "222"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := &DecisionTree{root: &rootDecisionNode{}}

			for i := range tt.args {
				err := tree.AddNode(tt.args[i])
				assert.NoError(t, err)
			}

			ids := tree.PrintTree()
			assert.Equal(t, tt.wantIds, ids)
		})
	}
}

func TestDecisionTree_Decision(t *testing.T) {
	NotThis := func(param interface{}) error {
		return errors.New("Not this. ")
	}

	tests := []struct {
		name      string
		args      []Node
		param     int
		wantParam int
	}{
		{
			name: "TEST1",
			args: []Node{
				testDecisionNode{Id: 1, Rule: false, ChildrenArr: []string{"11", "12"}},
				testDecisionNode{Id: 11, Rule: false, ChildrenArr: []string{"111"}},
				testDecisionNode{Id: 2, Rule: true, ChildrenArr: []string{"21"}},
				testExecNode{Id: 21, DoSomeThing: func(param interface{}) error {
					valueOfA := reflect.ValueOf(param)
					valueOfA = valueOfA.Elem()
					valueOfA.SetInt(2)
					return nil
				}},
				testExecNode{Id: 111, DoSomeThing: NotThis},
				testExecNode{Id: 12, DoSomeThing: NotThis},
			},
			param:     1,
			wantParam: 2,
		},
		{
			name: "TEST2",
			args: []Node{
				testDecisionNode{Id: 1, Rule: false, ChildrenArr: []string{"11", "12"}},
				testDecisionNode{Id: 11, Rule: false, ChildrenArr: []string{"111"}},
				testDecisionNode{Id: 2, Rule: true, ChildrenArr: []string{"21", "22"}},
				testDecisionNode{Id: 21, Rule: false, ChildrenArr: []string{"211", "212"}},
				testDecisionNode{Id: 22, Rule: true, ChildrenArr: []string{"221"}},
				testExecNode{Id: 211, DoSomeThing: NotThis},
				testExecNode{Id: 212, DoSomeThing: NotThis},
				testExecNode{Id: 111, DoSomeThing: NotThis},
				testExecNode{Id: 221, DoSomeThing: func(param interface{}) error {
					valueOfA := reflect.ValueOf(param)
					valueOfA = valueOfA.Elem()
					valueOfA.SetInt(3)
					return nil
				}},
				testExecNode{Id: 12, DoSomeThing: NotThis},
			},
			param:     1,
			wantParam: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := &DecisionTree{root: &rootDecisionNode{}}

			for i := range tt.args {
				err := tree.AddNode(tt.args[i])
				assert.NoError(t, err)
			}

			err := tree.Decision(&tt.param)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantParam, tt.param)
		})
	}
}
