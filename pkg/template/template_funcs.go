package template

import (
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/telecom-cloud/crafter/pkg/util"
)

var FuncMap = func() template.FuncMap {
	m := template.FuncMap{
		"ToLowerCamelCase": util.ToLowerCamelCase,
		"TrimSuffix":       util.TrimSuffix,
		"ToSnakeCase":      util.ToSnakeCase,
		"Split":            strings.Split,
		"Trim":             strings.Trim,
		"EqualFold":        strings.EqualFold,
		"ToHttpMethod":     util.ToHttpMethod,
	}
	for key, f := range sprig.TxtFuncMap() {
		m[key] = f
	}
	return m
}()
