package main

import (
	"flag"
	"github.com/marshome/3p-logrus"
	"io/ioutil"
	"github.com/marshome/i-api/codegen/golang"
	"github.com/marshome/x/filesystem"
	"fmt"
	"os"
)

func main() {
	specPath := flag.String("spec", "", "spec file path")
	serverPath := flag.String("server", "", "server file path")
	clientPath := flag.String("client", "", "client file path")
	codeLang := flag.String("lang", "go", "code lang,default is go")

	flag.Parse()

	if *specPath == "" {
		fmt.Fprintln(os.Stderr,"need spec")
		logrus.Exit(1)
	}

	if *serverPath==""&&*clientPath==""{
		fmt.Fprintln(os.Stderr,"need server or client")
		logrus.Exit(1)
	}

	specJson, err := ioutil.ReadFile(*specPath)
	if err != nil {
		fmt.Fprintln(os.Stderr,"read spec file faild,", err)
		logrus.Exit(1)
	}

	if *codeLang == "go" {
		if *serverPath != "" {
			code, err := codegen.GenerateServer(specJson)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				logrus.Exit(1)
			}

			err = filesystem.NewFile(*serverPath, code)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				logrus.Exit(1)
			}
		}

		if *clientPath != "" {
			code, err := codegen.GenerateClient(specJson)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				logrus.Exit(1)
			}

			err = filesystem.NewFile(*clientPath, code)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				logrus.Exit(1)
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "lang not support")
		logrus.Exit(1)
	}
}
