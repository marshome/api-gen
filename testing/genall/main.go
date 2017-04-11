package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"os"
	"path/filepath"
	"github.com/marshome/i-api/spec/googlespec"
	"github.com/marshome/i-api/codegen/golang"
	"bytes"
	"go/format"
	"github.com/marshome/i-api/spec"
	"github.com/marshome/x/filesystem"
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
	testingCode.WriteString(fmt.Sprintf(format + "\n", args...))
}

type ServiceInfo struct {
	Namespace   string
	ServiceName string
}

func parseServiceInfo(r *spec.Resource, name string, namespace string) [] *ServiceInfo {
	l := make([]*ServiceInfo, 0)
	l = append(l, &ServiceInfo{
		Namespace:namespace,
		ServiceName:codegen.GoName(name, true),
	})

	for _, sub := range r.Resources {
		l = append(l, parseServiceInfo(sub, name + "." + sub.Name, namespace)...)
	}

	return l
}

func genTestAll(imports []string,services []*ServiceInfo) {
	Pn("package main")
	Pn("")
	Pn("import(")
	Pn("\"net/http\"")
	Pn("\"github.com/marshome/i-api/genlib\"")
	for _, v := range imports {
		Pn("\"%s\"", strings.Replace(v, "..", "github.com/marshome/i-api/testing", -1))
	}
	Pn(")")
	Pn("")

	Pn("func main(){")
	for _, v := range services {
		Pn("%s_client.New(http.DefaultClient)", v.Namespace)
		Pn("%s.Handle%sService(genlib.NewRouter(),&%s.Default%sService_{})", v.Namespace, v.ServiceName, v.Namespace, v.ServiceName)
	}
	Pn("}")

	clean, err := format.Source(testingCode.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		filesystem.NewFile("../testall/main.go", testingCode.Bytes())
	} else {
		filesystem.NewFile("../testall/main.go", clean)
	}
}

func main() {
	dir_data, err := ioutil.ReadFile("../data/api-list.json")
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

	imports := make([]string, 0)
	services := make([]*ServiceInfo, 0)

	for _, v := range dir.Items {
		tokens := strings.Split(v.Id, ":")
		if len(tokens) != 2 {
			fmt.Println("error len(tokens)!=2")
			return
		}

		name := tokens[0]
		version := tokens[1]

		api_dir := "../data/" + strings.ToLower(name) + "/" + renameVersion(version) + "/"
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

		//spec
		spec := googlespec.Convert(&googleSpec)
		jsonData, err := json.MarshalIndent(spec, "", "    ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		filesystem.NewFile(json_path + ".spec.json", jsonData)

		//server
		serverCode, serverCodeErr := codegen.GenerateServer(jsonData)
		if serverCodeErr != nil {
			fmt.Fprintln(os.Stderr, serverCodeErr)
		}

		serverFile := filepath.Dir(json_path) + "/gen/server/server.go"

		filesystem.NewFile(serverFile, serverCode)

		if serverCodeErr != nil {
			return
		}

		//client
		clientCode, clientCodeErr := codegen.GenerateClient(jsonData)
		if clientCodeErr != nil {
			fmt.Fprintln(os.Stderr, clientCodeErr)
		}

		clientFile := filepath.Dir(json_path) + "/gen/client/client.go"

		filesystem.NewFile(clientFile, clientCode)

		if clientCodeErr != nil {
			return
		}

		if spec.Resources != nil&&len(spec.Resources) > 0 {
			imports = append(imports, strings.Replace(filepath.Dir(json_path) + "/gen/server", "\\", "/", -1))
			imports = append(imports, strings.Replace(filepath.Dir(json_path) + "/gen/client", "\\", "/", -1))
		}

		namespace := codegen.Namespace(name, version)
		if spec.Resources != nil {
			for _, r := range spec.Resources {
				services = append(services, parseServiceInfo(r, r.Name, namespace)...)
			}
		}
	}

	genTestAll(imports, services)
}
