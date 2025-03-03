package golang

var oneof = `
{{define "Oneof"}}
type {{$.InterfaceName}} interface {
	{{$.InterfaceName}}()
}

{{range $i, $f := .Choices}}
type {{$f.MessageName}}_{{$f.ChoiceName}} struct {
	{{$f.ChoiceName}} {{$f.Type.ResolveName ROOT}}
}
{{end}}

{{range $i, $f := .Choices}}
func (*{{$f.MessageName}}_{{$f.ChoiceName}}) {{$.InterfaceName}}() {}
{{end}}

{{range $i, $f := .Choices}}
func (p *{{$f.MessageName}}) Get{{$f.ChoiceName}}() {{$f.Type.ResolveName ROOT}} {
	if p, ok := p.Get{{$.OneofName}}().(*{{$f.MessageName}}_{{$f.ChoiceName}}); ok {
		return p.{{$f.ChoiceName}}
	}
	return {{$f.Type.ResolveDefaultValue}}
}
{{end}}

{{end}}
`
