package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/marshome/apis/generate"
)

var (
	specFile        = flag.String("spec", "", "")
	clientFile     = flag.String("client", "", "")
	serverFile     = flag.String("server", "", "")
	apiPackageBase = flag.String("api_pkg_base", "google.golang.org/api", "Go package prefix to use for all generated APIs.")
)

func main() {
	flag.Parse()

	if *specFile == "" {
		fmt.Println("need flag spec")
		return
	}

	docData, err := ioutil.ReadFile(*specFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	//client code
	if *clientFile != "" {
		params := generate.ClientGenerateParams{}
		params.ApiPackageBase = *apiPackageBase

		code, err := generate.GenerateClient(string(docData), &params)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = ioutil.WriteFile(*clientFile, []byte(code), 0)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	//server code
	if *serverFile != "" {
		params := generate.ServerGenerateParams{}
		params.ApiPackageBase = *apiPackageBase

		code, err := generate.GenerateServer(string(docData), &params)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = ioutil.WriteFile(*serverFile, []byte(code), 0)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
