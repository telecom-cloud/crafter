package generator

import (
	"errors"
	"fmt"
	"github.com/telecom-cloud/crafter/pkg/util/logs"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"

	"github.com/telecom-cloud/crafter/pkg/generator/model"
	"github.com/telecom-cloud/crafter/pkg/meta"
	tpl "github.com/telecom-cloud/crafter/pkg/template"
)

type PackageDescription struct {
	IdlName  string
	Package  string
	Services []*Service
	Models   []*model.Model
}

type Service struct {
	Name             string
	Version          string
	Methods          []*HttpMethod
	ClientMethods    []*ClientMethod
	DependencyModels []*model.Model
	ServiceGroup     string
	ServiceGenDir    string
	BaseDomain       string // base domain for client code
}

type HttpMethod struct {
	Name               string
	HTTPMethod         string
	Comment            string
	Path               string
	Serializer         string
	OutputDir          string
	RefPackage         string // handler import dir
	RefPackageAlias    string // handler import alias
	RequestTypeName    string
	RequestTypePackage string
	RequestTypeRawName string
	ReturnTypeName     string
	ReturnTypePackage  string
	ReturnTypeRawName  string
	ModelPackage       map[string]string
	GenHandler         bool // Whether to generate one handler, when an idl interface corresponds to multiple http method
	// Annotations     map[string]string
	Models map[string]*model.Model
}

// HttpPackageGenerator is used to record the configuration related to generating crafter http code.
type HttpPackageGenerator struct {
	ServiceGroup string

	ConfigPath     string // package template path
	CmdType        string
	Backend        meta.Backend // model template
	Options        []Option
	ProjPackage    string // go module for project
	ModelDir       string
	UseDir         string // model dir for third repo
	ClientDir      string // client dir for "new"/"update" command
	IdlClientDir   string // client dir for "client" command
	ForceClientDir string // client dir without namespace for "client" command
	BaseDomain     string // request domain for "client" command
	QueryEnumAsInt bool   // client code use number for query parameter
	ServiceGenDir  string

	NeedModel            bool
	SnakeStyleMiddleware bool // use snake name style for middleware
	ForceUpdateClient    bool // force update 'crafter_client.go'

	loadedBackend   ModelBackend
	curModel        *model.Model
	processedModels map[*model.Model]bool

	tpl.TemplateGenerator
}

func (pkgGen *HttpPackageGenerator) Init() error {
	defaultConfig := packageConfig
	customConfig := tpl.Config{}
	// unmarshal from user-defined config file if it exists
	if pkgGen.ConfigPath != "" {
		cdata, err := os.ReadFile(pkgGen.ConfigPath)
		if err != nil {
			return fmt.Errorf("read layout config from  %s failed, err: %v", pkgGen.ConfigPath, err.Error())
		}
		if err = yaml.Unmarshal(cdata, &customConfig); err != nil {
			return fmt.Errorf("unmarshal layout config failed, err: %v", err.Error())
		}
		if reflect.DeepEqual(customConfig, tpl.Config{}) {
			return errors.New("empty config")
		}
	}

	if pkgGen.TemplateGenerator.Templates == nil {
		logs.Info("templates is nil")
		logs.Flush()
		pkgGen.TemplateGenerator.Templates = make(map[string]*tpl.TemplateInfo, len(defaultConfig.Layouts))
	}

	// load default template
	for _, layout := range defaultConfig.Layouts {
		// default template use "fileName" as template name
		path := filepath.Base(layout.Path)
		err := pkgGen.LoadLayout(layout, path, true)
		if err != nil {
			return err
		}
	}

	// override the default template, other customized file template will be loaded by "TemplateGenerator.Init"
	for _, layout := range customConfig.Layouts {
		if !tpl.IsDefaultPackageTpl(layout.Path) {
			continue
		}
		err := pkgGen.LoadLayout(layout, layout.Path, true)
		if err != nil {
			return err
		}
	}

	pkgGen.Config = &customConfig
	// load Model tpl if you need
	if pkgGen.Backend != "" {
		if err := pkgGen.LoadBackend(pkgGen.Backend); err != nil {
			return fmt.Errorf("load model template failed, err: %v", err.Error())
		}
	}

	pkgGen.processedModels = make(map[*model.Model]bool)
	pkgGen.TemplateGenerator.IsPackageTpl = true

	return pkgGen.TemplateGenerator.Init()
}

func (pkgGen *HttpPackageGenerator) checkInit() (bool, error) {
	if pkgGen.TemplateGenerator.Templates == nil {
		if err := pkgGen.Init(); err != nil {
			return false, fmt.Errorf("init layout config failed, err: %v", err.Error())
		}
	}
	return pkgGen.ConfigPath == "", nil
}

func (pkgGen *HttpPackageGenerator) GeneratePackage(pkg *PackageDescription) error {
	if _, err := pkgGen.checkInit(); err != nil {
		return err
	}
	if len(pkg.Models) != 0 {
		for _, m := range pkg.Models {
			if err := pkgGen.GenModel(m, pkgGen.NeedModel); err != nil {
				return fmt.Errorf("generate model %s failed, err: %v", m.FilePath, err.Error())
			}
		}
	}

	switch pkgGen.CmdType {
	case meta.CmdClient:
		// default client dir
		clientDir := pkgGen.IdlClientDir
		// user specify client dir
		if len(pkgGen.ClientDir) != 0 {
			clientDir = pkgGen.ClientDir
		}
		if err := pkgGen.genClient(pkg, clientDir); err != nil {
			return err
		}
		if err := pkgGen.genHttpClient(pkgGen.ClientDir, pkgGen.ServiceGroup); err != nil {
			return err
		}
		if err := pkgGen.genCustomizedFile(pkg); err != nil {
			return err
		}
		return nil
	case meta.CmdError:
		if err := pkgGen.genError(pkgGen.ModelDir); err != nil {
			return err
		}
		return nil
	case meta.CmdDoc:

	}

	if err := pkgGen.genCustomizedFile(pkg); err != nil {
		return err
	}

	return nil
}
