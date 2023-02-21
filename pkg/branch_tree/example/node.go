package main

import "fmt"

type People struct {
	Age    int
	Height int
}

type AgeDecisionNode struct {
	Start int
	End   int

	Childs []string
}

func NewAgeDecisionNode(s, e int) *AgeDecisionNode {
	return &AgeDecisionNode{
		Start: s,
		End:   e,
	}
}

func (a *AgeDecisionNode) ID() string {
	return fmt.Sprintf("S%d-E%d", a.Start, a.End)
}

func (a *AgeDecisionNode) DecisionRule(param interface{}) bool {
	if p, ok := param.(*People); ok {
		return p.Age >= a.Start && p.Age < a.End
	}
	return false
}

func (a *AgeDecisionNode) SetChilds(cl []string) {
	a.Childs = cl
}

func (a *AgeDecisionNode) ChildrenList() []string {
	return a.Childs
}

type HeightDecisionNode struct {
	Id           string
	DecisionFunc func(param interface{}) bool
	Childs       []string
}

func (h *HeightDecisionNode) ID() string {
	return h.Id
}

func (h *HeightDecisionNode) DecisionRule(param interface{}) bool {
	return h.DecisionFunc(param)
}

func (h *HeightDecisionNode) ChildrenList() []string {
	return h.Childs
}

type EnterPark struct {
}

func (e *EnterPark) ID() string {
	return "EnterPark"
}

func (e *EnterPark) Do(param interface{}) error {
	fmt.Println("I will enter park. ")
	return nil
}

type BuyTicket struct {
}

func (e *BuyTicket) ID() string {
	return "BuyTicket"
}

func (e *BuyTicket) Do(param interface{}) error {
	fmt.Println("I will buy ticket. ")
	return nil
}

type DiscountTicket struct {
}

func (e *DiscountTicket) ID() string {
	return "DiscountTicket"
}

func (e *DiscountTicket) Do(param interface{}) error {
	fmt.Println("I will discount. ")
	return nil
}
