package generator

import (
	"path/filepath"

	tpl "github.com/telecom-cloud/crafter/pkg/template"
)

type ErrorFile struct {
}

func (pkgGen *HttpPackageGenerator) genError(modelDir string) error {
	errorPath := filepath.Join(modelDir, tpl.ErrorTplName)
	input := map[string]interface{}{
		"PackageName": "types",
	}
	err := pkgGen.TemplateGenerator.Generate(input, tpl.ErrorTplName, errorPath, false)
	if err != nil {
		return err
	}
	return nil
}
