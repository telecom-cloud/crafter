package golang

// Enum .
var enum = `
{{define "Enum"}}
{{- $EnumType := (Identify .Name)}}
type {{$EnumType}} {{.GoType}}

const (
	{{- range $i, $e := .Values}}
	{{$EnumType}}_{{$e.Name}} {{$EnumType}} = {{$e.Value.Expression}}
	{{- end}}
)

func (p {{$EnumType}}) String() string {
	switch p {
	{{- range $i, $e := .Values}}
	case {{$EnumType}}_{{$e.Name}}:
		return "{{printf "%s%s" $EnumType $e.Name | SnakeCase}}"
	{{- end}}
	}
	return "<UNSET>"
}

func {{$EnumType}}FromString(s string) ({{$EnumType}}, error) {
	switch s {
	{{- range $i, $e := .Values}}
	case "{{printf "%s%s" $EnumType $e.Name | SnakeCase}}":
		return {{$EnumType}}_{{$e.Name}}, nil
	{{- end}}
	}
	return {{$EnumType}}(0), fmt.Errorf("not a valid {{$EnumType}} string")
}

{{- if Features.MarshalEnumToText}}

func (p {{$EnumType}}) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *{{$EnumType}}) UnmarshalText(text []byte) error {
	q, err := {{$EnumType}}FromString(string(text))
	if err != nil {
		return err
	}
	*p = q
	return nil
}
{{- end}}
{{end}}
`
