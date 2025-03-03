package golang

import "text/template"

type Backend struct{}

func (gb *Backend) Template() (*template.Template, error) {
	return Template()
}

func (gb *Backend) Element() map[string]string {
	return Element()
}

func (gb *Backend) SetOption(opts string) error {
	return SetOption(opts)
}

func (gb *Backend) GetOptions() []string {
	return GetOptions()
}

func (gb *Backend) RegisterFuncs(name string, fn interface{}) error {
	return RegisterFuncs(name, fn)
}
