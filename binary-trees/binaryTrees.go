package main

import (
	"fmt"
	"golang.org/x/tour/tree"
)

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan int) {
	left := t.Left
	right := t.Right
	if left != nil {
		Walk(left, ch)
	}
	ch <- t.Value
	if right != nil {
		Walk(right, ch)
	}
	
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *tree.Tree) bool {
	ch1, ch2 := make(chan int), make(chan int)
	go Walk(t1, ch1)
	go Walk(t2, ch2)
	for i :=0 ; i < 10; i++{
		if <-ch1 != <-ch2 {
			return false
		}
	}
	return true
}

func main() {
	ch := make(chan int)
	testTree := tree.New(1)
	go Walk(testTree, ch)
	for i :=0 ; i < 10; i++{
	 	fmt.Println(<-ch)
	}
	close(ch)
	theSameTest := Same(tree.New(1), tree.New(1))
	if (theSameTest) {
		fmt.Println("the same")
	}
	differntTest := Same(tree.New(1), tree.New(2))
	if (!differntTest) {
		fmt.Println("Different")
	}
}