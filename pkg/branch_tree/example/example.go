package main

import (
	"github.com/HYY-yu/seckill.pkg/pkg/branch_tree"
	"math"
)

// 生成以下树：
//                                           ┌──────┐
//                                           │ Root │
//                                           └──────┘
//                                               ▲
//                  ┌────────────────────────────┼──────────────────────┐
//                  │                            │                      │
//           ┌────────────┐          ┌───────────────────────┐   ┌─────────────┐
//           │  age < 18  │          │ age >= 18 & age < 60  │   │  age >= 60  │
//           └────────────┘          └───────────────────────┘   └─────────────┘
//                  ▲                            ▲                      ▲
//          ╱──────╱ ╲──────╲                    │                      │
//         ╱                 ╲            ╔════════════╗       ╔═════════════════╗
// ┌───────────────┐ ┌───────────────┐    ║ buy ticket ║       ║ discount ticket ║
// │ height < 1.2  │ │ height >= 1.2 │    ╚════════════╝       ╚═════════════════╝
// └───────────────┘ └───────────────┘
//         ▲                 ▲
//         │                 │
//    ╔═════════╗     ╔════════════╗
//    ║  enter  ║     ║ buy ticket ║
//    ╚═════════╝     ╚════════════╝

func main() {
	tree := branch_tree.DecisionTree{}

	enterNode := &EnterPark{}
	buyTicketNode := &BuyTicket{}
	discountTicketNode := &DiscountTicket{}

	lessHeight := &HeightDecisionNode{
		Id: "H<120cm",
		DecisionFunc: func(param interface{}) bool {
			if p, ok := param.(*People); ok {
				return p.Height < 120
			}
			return false
		},
		Childs: []string{enterNode.ID()},
	}
	geHeight := &HeightDecisionNode{
		Id: "H>=120cm",
		DecisionFunc: func(param interface{}) bool {
			if p, ok := param.(*People); ok {
				return p.Height >= 120
			}
			return false
		},
		Childs: []string{buyTicketNode.ID()},
	}

	ageLess18 := NewAgeDecisionNode(0, 18)
	ageIn18And60 := NewAgeDecisionNode(18, 60)
	ageG60 := NewAgeDecisionNode(60, math.MaxInt)

	ageLess18.SetChilds([]string{lessHeight.ID(), geHeight.ID()})
	ageIn18And60.SetChilds([]string{buyTicketNode.ID()})
	ageG60.SetChilds([]string{discountTicketNode.ID()})

	err := tree.AddNode(enterNode)
	err = tree.AddNode(buyTicketNode)
	err = tree.AddNode(discountTicketNode)
	err = tree.AddNode(lessHeight)
	err = tree.AddNode(geHeight)
	err = tree.AddNode(ageLess18)
	err = tree.AddNode(ageIn18And60)
	err = tree.AddNode(ageG60)
	if err != nil {
		panic(err)
	}

	err = tree.Decision(&People{
		Age:    16,
		Height: 150,
	})
	if err != nil {
		panic(err)
	}
	// Buy ticket

	err = tree.Decision(&People{
		Age:    8,
		Height: 110,
	})
	if err != nil {
		panic(err)
	}
	// Enter

	err = tree.Decision(&People{
		Age:    61,
		Height: 160,
	})
	if err != nil {
		panic(err)
	}
	// Discount

	err = tree.Decision(&People{
		Age:    45,
		Height: 160,
	})
	if err != nil {
		panic(err)
	}
	// Buy ticket

	// Run result:
	// I will buy ticket.
	// I will enter park.
	// I will discount.
	// I will buy ticket.
}
