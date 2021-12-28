package validate

import (
	"path/filepath"

	"emperror.dev/errors"
	"github.com/xeipuuv/gojsonschema"
)

// IsDraft4 returns an error if the input file is not a valid draft4 JSON schema
func IsDraft4(path string) error {
	schemaLoader := gojsonschema.NewSchemaLoader()
	schemaLoader.Draft = gojsonschema.Draft4

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrapf(err, "could not get absolute path for %s", path)
	}

	_, err = schemaLoader.Compile(gojsonschema.NewReferenceLoader("file://" + absPath))
	return err
}
