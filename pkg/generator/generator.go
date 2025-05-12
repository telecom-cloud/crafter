package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/telecom-cloud/crafter/pkg/generator/model"
	"github.com/telecom-cloud/crafter/pkg/generator/model/golang"
	"github.com/telecom-cloud/crafter/pkg/meta"
	tpl "github.com/telecom-cloud/crafter/pkg/template"
	"github.com/telecom-cloud/crafter/pkg/util"
)

//---------------------------------Backend----------------------------------

type Option string

const (
	OptionMarshalEnumToText  Option = "MarshalEnumToText"
	OptionTypedefAsTypeAlias Option = "TypedefAsTypeAlias"
)

type ModelBackend interface {
	Template() (*template.Template, error)
	Element() map[string]string
	SetOption(opts string) error
	GetOptions() []string
	RegisterFuncs(name string, fn interface{}) error
}

func switchBackend(backend meta.Backend) ModelBackend {
	switch backend {
	case meta.BackendGolang:
		return &golang.Backend{}
	case meta.BackendJava:
	case meta.BackendPython:
	case meta.BackendRuby:
	case meta.BackendRust:
	}
	return loadThirdPartyBackend(string(backend))
}

func loadThirdPartyBackend(backend string) ModelBackend {
	panic("no implement yet!")
}

/**********************Generating*************************/

func (pkgGen *HttpPackageGenerator) LoadBackend(backend meta.Backend) error {
	bd := switchBackend(backend)
	if bd == nil {
		return fmt.Errorf("no found backend '%s'", backend)
	}
	for _, opt := range pkgGen.Options {
		if err := bd.SetOption(string(opt)); err != nil {
			return fmt.Errorf("set option %s error, err: %v", opt, err.Error())
		}
	}

	err := bd.RegisterFuncs("ROOT", func() *model.Model {
		return pkgGen.curModel
	})
	if err != nil {
		return fmt.Errorf("register global function in model template failed, err: %v", err.Error())
	}

	generateTpl, err := bd.Template()
	if err != nil {
		return fmt.Errorf("load backend %s failed, err: %v", backend, err.Error())
	}

	if pkgGen.TemplateGenerator.Templates == nil {
		pkgGen.TemplateGenerator.Templates = map[string]*tpl.TemplateInfo{}
	}
	pkgGen.TemplateGenerator.Templates[tpl.ModelTplName].Template = generateTpl
	pkgGen.loadedBackend = bd
	return nil
}

func (pkgGen *HttpPackageGenerator) GenModel(data *model.Model, gen bool) error {
	if pkgGen.processedModels == nil {
		pkgGen.processedModels = map[*model.Model]bool{}
	}

	if _, ok := pkgGen.processedModels[data]; !ok {
		var path string
		var updatePackage bool
		if strings.HasPrefix(data.Package, pkgGen.ProjPackage) && data.PackageName != pkgGen.ProjPackage {
			path = data.Package[len(pkgGen.ProjPackage):]
		} else {
			path = data.Package
			updatePackage = true
		}
		modelDir := util.SubDir(pkgGen.ModelDir, path)
		if updatePackage {
			data.Package = util.SubPackage(pkgGen.ProjPackage, modelDir)
		}
		data.FilePath = filepath.Join(modelDir, util.BaseNameAndTrim(data.FilePath)+".go")

		pkgGen.processedModels[data] = true
	}

	for _, dep := range data.Imports {
		if err := pkgGen.GenModel(dep, false); err != nil {
			return fmt.Errorf("generate model %s failed, err: %v", dep.FilePath, err.Error())
		}
	}

	if gen && !data.IsEmpty() {
		pkgGen.curModel = data
		removeDuplicateImport(data)
		err := pkgGen.TemplateGenerator.Generate(data, tpl.ModelTplName, data.FilePath, false)
		pkgGen.curModel = nil
		return err
	}
	return nil
}

// Idls with the same PackageDescription do not need to refer to each other
func removeDuplicateImport(data *model.Model) {
	for k, v := range data.Imports {
		if data.Package == v.Package {
			delete(data.Imports, k)
		}
	}
}

type Generator interface {
	Generate(serviceName string)
	GenerateDocs(serviceName string)
}
