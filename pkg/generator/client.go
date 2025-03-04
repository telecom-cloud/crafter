package generator

import (
	"path/filepath"
	"strings"

	"github.com/telecom-cloud/crafter/pkg/generator/model"
	"github.com/telecom-cloud/crafter/pkg/meta"
	tpl "github.com/telecom-cloud/crafter/pkg/template"
	"github.com/telecom-cloud/crafter/pkg/util"
)

const ServiceSuffix = "Service"

type ClientMethod struct {
	*HttpMethod
	BodyParamsCode   string
	QueryParamsCode  string
	PathParamsCode   string
	HeaderParamsCode string
	FormValueCode    string
	FormFileCode     string
	DecodeCustomKey  string
}

type ClientConfig struct {
	QueryEnumAsInt bool
}

type ClientGenerator interface {
}

type ClientFile struct {
	Config        ClientConfig
	FilePath      string
	PackageName   string
	ServiceName   string
	BaseDomain    string
	Imports       map[string]*model.Model
	ClientMethods []*ClientMethod
}

func (pkgGen *HttpPackageGenerator) genClient(pkg *PackageDescription, clientDir string) error {
	module, _, _ := util.SearchGoMod(".", true)
	baseDomain := pkgGen.BaseDomain
	generatedJson := &meta.GeneratedJSON{}
	serviceGroupDir := filepath.Join(clientDir, pkgGen.ServiceGroup)
	generatedJsonFile := filepath.Join(serviceGroupDir, "generated.json")
	for _, s := range pkg.Services {
		if baseDomain == "" {
			baseDomain = s.BaseDomain
		}
		cliDir := serviceGroupDir
		if len(pkgGen.ForceClientDir) != 0 {
			cliDir = pkgGen.ForceClientDir
		}
		idlClientFilePath := filepath.Join(cliDir, strings.ToLower(s.Name+".go"))
		client := ClientFile{
			FilePath:      idlClientFilePath,
			PackageName:   pkgGen.ServiceGroup,
			ServiceName:   util.ToCamelCase(s.Name),
			ClientMethods: s.ClientMethods,
			BaseDomain:    baseDomain,
			Config: ClientConfig{
				QueryEnumAsInt: pkgGen.QueryEnumAsInt,
			},
		}
		client.Imports = make(map[string]*model.Model, len(client.ClientMethods)+1)
		for _, m := range client.ClientMethods {
			// Iterate over the request and return parameters of the method to get import path.
			for key, mm := range m.Models {
				if v, ok := client.Imports[mm.PackageName]; ok && v.Package != mm.Package {
					client.Imports[key] = mm
					continue
				}
				client.Imports[mm.PackageName] = mm
			}
		}
		if len(pkgGen.UseDir) != 0 {
			oldModelDir := filepath.Clean(filepath.Join(pkgGen.ProjPackage, pkgGen.ModelDir))
			newModelDir := filepath.Clean(pkgGen.UseDir)
			for _, m := range client.ClientMethods {
				for _, mm := range m.Models {
					mm.Package = strings.Replace(mm.Package, oldModelDir, newModelDir, 1)
				}
			}
		}
		err := pkgGen.TemplateGenerator.Generate(client, tpl.IdlClientTplName, client.FilePath, false)
		if err != nil {
			return err
		}
		serviceJson := &meta.GeneratedJSON{
			ServiceGroup: pkgGen.ServiceGroup,
			Module:       module,
			Clients: []string{
				strings.TrimSuffix(s.Name, ServiceSuffix),
			},
		}
		generatedJson, err = meta.LoadGeneratedJson(generatedJsonFile, serviceJson)
		if err != nil {
			return err
		}
	}

	return pkgGen.genServiceGroup(serviceGroupDir, baseDomain, generatedJson)
}

func (pkgGen *HttpPackageGenerator) genServiceGroup(serviceGroupDir, baseDomain string, generatedJson *meta.GeneratedJSON) error {
	return pkgGen.TemplateGenerator.Generate(map[string]interface{}{
		"ServiceGroup": generatedJson.ServiceGroup,
		"Module":       generatedJson.Module,
		"BaseDomain":   baseDomain,
		"Clients":      generatedJson.Clients,
	}, tpl.IdlGroupClientTplName, filepath.Join(serviceGroupDir, strings.ToLower(pkgGen.ServiceGroup))+".go", false)
}

func (pkgGen *HttpPackageGenerator) genHttpClient(clientDir, serviceGroupDir string) error {
	httpClientPath := filepath.Join(clientDir, serviceGroupDir, tpl.HttpClientTplName)
	isExist, err := util.PathExist(httpClientPath)
	if err != nil {
		return err
	}
	// generate http client once

	httpClient := map[string]interface{}{
		"PackageName":    serviceGroupDir,
		"QueryEnumAsInt": pkgGen.QueryEnumAsInt,
	}
	if !isExist || pkgGen.ForceUpdateClient {
		err := pkgGen.TemplateGenerator.Generate(httpClient, tpl.HttpClientTplName, httpClientPath, false)
		if err != nil {
			return err
		}
	}
	return nil
}
