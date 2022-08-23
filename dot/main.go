package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/diff"
)

type Object struct {
	Foo string
	Bar int
}

func main() {
	o1 := &Object{
		Foo: "",
		Bar: 0,
	}
	o2 := &Object{
		Foo: "",
		Bar: 0,
	}

	s := diff.ObjectGoPrintSideBySide(o1, o2)
	fmt.Printf("%s\n", s)
}
