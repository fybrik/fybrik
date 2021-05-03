package taxonomy

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

func main() {

	// Read M4D application yaml file
	applicationYaml, err := ioutil.ReadFile("../../samples/kubeflow/m4dapplication.yaml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("The M4D application spec is: \n")
	fmt.Println(string(applicationYaml))

	// Convert M4D application yaml to json
	applicationJson, err := yaml.YAMLToJSON(applicationYaml)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	// Validate against taxonomy
	path, err := filepath.Abs("application.values.schema.json")
	if err != nil {
		panic(err.Error())
	}

	taxonomyLoader := gojsonschema.NewReferenceLoader("file://" + path)
	documentLoader := gojsonschema.NewStringLoader(string(applicationJson))
	result, err := gojsonschema.Validate(taxonomyLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		fmt.Printf("This M4D application is valid\n")
	} else {
		fmt.Printf("This M4D application is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
}
