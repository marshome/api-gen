package main

import (
	"github.com/go-openapi/loads"
	"fmt"
	"github.com/marshome/api-gen/pkg/generator"
)

func main() {
	doc, err := loads.JSONSpec("./swagger.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	code, err := generator.GenerateServer(doc.Spec(), "api")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(code)
}