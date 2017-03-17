package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/marshome/apis/spec"
	"os"
	"github.com/marshome/apis/generate"
	"path/filepath"
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

func build_commets(c string, buffer *bytes.Buffer, tab int) {
	if strings.Contains(c, "\n\n") {
		commet_lines := strings.Split(c, "\n\n")
		for commet_index, commet := range commet_lines {
			for i := 0; i < tab; i++ {
				buffer.WriteString("    ")
			}
			buffer.WriteString("//" + commet + "\n")
			if commet_index < len(commet_lines)-1 {
				for i := 0; i < tab; i++ {
					buffer.WriteString("    ")
				}
				buffer.WriteString("//\n")
			}
		}
	} else {
		for i := 0; i < tab; i++ {
			buffer.WriteString("    ")
		}
		buffer.WriteString("//" + c + "\n")
	}
}

func object_to_struct(obj *spec.APIObject, buffer *bytes.Buffer) (err error) {
	if obj.Type != "object" {
		return errors.New("obj is not a struct " + obj.Id)
	}

	build_commets(obj.Description, buffer, 0)
	buffer.WriteString("type " + obj.Id + " struct{\n")
	for property_name, property := range obj.Properties {
		build_commets(property.Description, buffer, 1)
		property_name_go := strings.ToUpper(string(property_name[0]))
		property_name_go += property_name[1:]
		if property.Ref != "" {
			buffer.WriteString("    " + property_name_go + " ")
			buffer.WriteString("*" + property.Ref + "`json:\"" + property_name + "\"`\n")
		} else {
			if property.Type == "object" {
				if property.AdditionalProperties != nil {
					buffer.WriteString("    " + property_name_go + " ")
					buffer.WriteString("map[string]*")
					if property.AdditionalProperties.Ref != "" {
						buffer.WriteString(property.AdditionalProperties.Ref)
					}
				} else {
					buffer.WriteString("    " + property_name_go + " ")
					buffer.WriteString("struct{\n")
					buffer.WriteString("    } `json:\"" + property_name + "\"`")
				}
			} else {
				buffer.WriteString("    " + property_name_go + " ")
				buffer.WriteString("string `json:\"" + property_name + "\"`\n")
			}
		}
	}
	buffer.WriteString("}\n")
	buffer.WriteString("\n")

	return nil
}

func schemas_to_go(schemas map[string]*spec.APIObject) (src string, err error) {
	var buffer bytes.Buffer
	buffer.WriteString("package schemas\n")
	buffer.WriteString("\n")

	for _, schema := range schemas {
		err = object_to_struct(schema, &buffer)
		if err != nil {
			return "", err
		}
	}

	return buffer.String(), nil
}

func resource_recursive(m map[string]*spec.APIResource) {
	for _, v := range m {
		//for _, meth := range v.Methods {
		//fmt.Println(meth.Id + " " + " " + meth.Path)
		//}
		resource_recursive(v.Resources)
	}
}

func output_file(filePath string,data []byte)(err error) {
	dir := filepath.Dir(filePath)
	dirInfo, err := os.Stat(dir)
	if dirInfo == nil || os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(filePath, data, os.ModePerm | os.ModeExclusive)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	dir_data, err := ioutil.ReadFile("../testdata/api-list.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	dir := spec.APIDirectory{}
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

		api_dir := "../testdata/" + strings.ToLower(name) + "/" + renameVersion(version) + "/"
		json_path := api_dir + strings.ToLower(name) + "-api.json"

		docData, err := ioutil.ReadFile(json_path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("<ApiSpec>", json_path)

		genParam := &generate.ServerGenerateParams{}
		genParam.Namespace = strings.Replace(name + "_" + version, ".", "_", -1)
		code, codeErr := generate.GenerateServer(string(docData), genParam)
		if codeErr != nil {
			fmt.Println(codeErr)
		}

		serverFile := filepath.Dir(json_path) + "/gen/server/api/api.go"

		output_file(serverFile, []byte(code))

		if codeErr != nil {
			return
		}
	}
}
