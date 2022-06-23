package main

import (
	"fmt"
	_ "github.com/distribution/distribution/v3/registry/auth/htpasswd"           // used for docker test registry
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory" // used for docker test registry
	"io/ioutil"
)

func main() {
	body, _ := ioutil.ReadFile("2021.10.json")
	tags, _ := unmarshallUnstructured(body)
	fmt.Printf("%v\n", tags)

	tags, _ = getTypedTagList(body)
	fmt.Printf("%v\n", tags)

}
