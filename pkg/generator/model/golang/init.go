package golang

import (
	"fmt"
	"strings"
	"text/template"
)

var tpls *template.Template

var element = map[string]string{
	"file":      file,
	"typedef":   typedef,
	"constants": constants,
	"variables": variables,
	"function":  function,
	"enum":      enum,
	"struct":    structLike,
	"method":    method,
	"oneof":     oneof,
}

/**************************Export API*******************************/

func Template() (*template.Template, error) {
	if tpls != nil {
		return tpls, nil
	}
	tpls = new(template.Template)

	tpls = tpls.Funcs(funcMap)

	var err error
	for k, e := range element {
		tpls, err = tpls.Parse(e)
		if err != nil {
			return nil, fmt.Errorf("parse template '%s' failed, err: %v", k, err.Error())
		}
	}
	return tpls, nil
}

func Element() map[string]string {
	return element
}

/***********************Template Functions**************************/

var funcMap = template.FuncMap{
	"Features":            getFeatures,
	"Identify":            identify,
	"CamelCase":           camelCase,
	"SnakeCase":           snakeCase,
	"GetTypedefReturnStr": getTypedefReturnStr,
}

func RegisterFuncs(name string, fn interface{}) error {
	if _, ok := funcMap[name]; ok {
		return fmt.Errorf("duplicate function: %s has been registered", name)
	}
	funcMap[name] = fn
	return nil
}

func identify(name string) string {
	return name
}

func camelCase(name string) string {
	return name
}

func snakeCase(name string) string {
	return name
}

func getTypedefReturnStr(name string) string {
	if strings.Contains(name, ".") {
		idx := strings.LastIndex(name, ".")
		return name[:idx] + "." + "New" + name[idx+1:] + "()"

	}
	return "New" + name + "()"
}

/************************Template Options**************************/

type feature struct {
	MarshalEnumToText  bool
	TypedefAsTypeAlias bool
}

var features = feature{}

func getFeatures() feature {
	return features
}

func SetOption(opt string) error {
	switch opt {
	case "MarshalEnumToText":
		features.MarshalEnumToText = true
	case "TypedefAsTypeAlias":
		features.TypedefAsTypeAlias = true
	}
	return nil
}

var Options = []string{
	"MarshalEnumToText",
	"TypedefAsTypeAlias",
}

func GetOptions() []string {
	return Options
}
