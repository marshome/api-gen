package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/marshome/i-api/spec/googlespec"
	"github.com/marshome/i-api/spec/goolge2swagger"
	"github.com/marshome/x/filesystem"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var oddVersionRE = regexp.MustCompile(`^(.+)_(v[\d\.]+)$`)

func renameVersion(version string) string {
	if version == "alpha" || version == "beta" {
		return "v0." + version
	}
	if m := oddVersionRE.FindStringSubmatch(version); m != nil {
		return m[1] + "/" + m[2]
	}
	return version
}

var testingCode = bytes.NewBufferString("")

func P(format string, args ...interface{}) {
	testingCode.WriteString(fmt.Sprintf(format, args...))
}

func Pn(format string, args ...interface{}) {
	testingCode.WriteString(fmt.Sprintf(format+"\n", args...))
}

type ServiceInfo struct {
	Namespace   string
	ServiceName string
}

func main() {
	dir_data, err := ioutil.ReadFile("./testing/data/api-list.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	dir := googlespec.APIDirectory{}
	err = json.Unmarshal(dir_data, &dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range dir.Items {
		tokens := strings.Split(v.Id, ":")
		if len(tokens) != 2 {
			fmt.Println("error len(tokens)!=2")
			return
		}

		name := tokens[0]
		version := tokens[1]

		api_dir := "./testing/data/" + strings.ToLower(name) + "/" + renameVersion(version) + "/"
		json_path := api_dir + strings.ToLower(name) + "-api.json"

		docData, err := ioutil.ReadFile(json_path)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("<ApiSpec>", json_path)

		googleSpec := googlespec.APIDocument{}
		err = json.Unmarshal(docData, &googleSpec)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		swagerSpec, err := goolge2swagger.Convert(&googleSpec)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		swaggerJson, err := swagerSpec.MarshalJSON()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		swaggerJsonIndentBuffer:=bytes.NewBufferString("")
		json.Indent(swaggerJsonIndentBuffer,swaggerJson,"","  ")

		filesystem.NewFile(api_dir+"/swagger.json", []byte(swaggerJsonIndentBuffer.String()))
	}
}
