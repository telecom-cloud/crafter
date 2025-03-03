package golang

// StructLike is the code template for struct, union, and exception.
var structLike = `
{{define "Struct"}}
{{- $TypeName := (Identify .Name) -}}
{{$MessageLeadingComments := .LeadingComments}}
{{if ne (len $MessageLeadingComments) 0}}
//{{$MessageLeadingComments}}
{{end -}}
type {{$TypeName}} struct {
{{- range $i, $f := .Fields}}
{{- $FieldLeadingComments := $f.LeadingComments}}
{{$FieldTrailingComments := $f.TrailingComments -}}
{{- if ne (len $FieldLeadingComments) 0 -}}
    //{{$FieldLeadingComments}}
{{end -}}
{{- if $f.IsPointer -}}
	{{$f.Name}} *{{$f.Type.ResolveName ROOT}} {{$f.GenGoTags}}{{if ne (len $FieldTrailingComments) 0}} //{{$FieldTrailingComments}}{{end -}}
{{- else -}}
	{{$f.Name}} {{$f.Type.ResolveName ROOT}} {{$f.GenGoTags}}{{if ne (len $FieldTrailingComments) 0}} //{{$FieldTrailingComments}}{{end -}}
{{- end -}}
{{- end}}
}

func New{{$TypeName}}() *{{$TypeName}} {
	return &{{$TypeName}}{
		{{template "StructLikeDefault" .}}
	}
}

{{template "FieldGetOrSet" .}}

{{if eq .Category 14}}
func (p *{{$TypeName}}) CountSetFields{{$TypeName}}() int {
	count := 0
	{{- range $i, $f := .Fields}}
	{{- if $f.Type.IsSettable}}
	if p.IsSet{{$f.Name}}() {
		count++
	}
	{{- end}}
	{{- end}}
	return count
}
{{- end}}

func (p *{{$TypeName}}) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{{$TypeName}}(%+v)", *p)
}

{{- if eq .Category 15}}
func (p *{{$TypeName}}) Error() string {
	return p.String()
}
{{- end}}
{{- end}}{{/* define "StructLike" */}}

{{- define "StructLikeDefault"}}
{{- range $i, $f := .Fields}}
	{{- if $f.IsSetDefault}}
		{{$f.Name}}: {{$f.DefaultValue.Expression}},
	{{- end}}
{{- end}}
{{- end -}}{{/* define "StructLikeDefault" */}}

{{- define "FieldGetOrSet"}}
{{- $TypeName := (Identify .Name)}}
{{- range $i, $f := .Fields}}
{{$FieldName := $f.Name}}
{{$FieldTypeName := $f.Type.ResolveName ROOT}}

{{- if $f.Type.IsSettable}}
func (p *{{$TypeName}}) IsSet{{$FieldName}}() bool {
	return p.{{$FieldName}} != nil
}
{{- end}}{{/* IsSettable . */}}

func (p *{{$TypeName}}) Get{{$FieldName}}() {{$FieldTypeName}} {
	{{- if $f.Type.IsSettable}}
	if !p.IsSet{{$FieldName}}() {
		return {{with $f.DefaultValue}}{{$f.DefaultValue.Expression}}{{else}}nil{{end}}
	}
	{{- end}}
{{- if $f.IsPointer}}
	return *p.{{$FieldName}}
{{else}}
	return p.{{$FieldName}}
{{- end -}}
}

func (p *{{$TypeName}}) Set{{$FieldName}}(val {{$FieldTypeName}}) {
{{- if $f.IsPointer}}
	*p.{{$FieldName}} = val
{{else}}
	p.{{$FieldName}} = val
{{- end -}}
}
{{- end}}{{/* range .Fields */}}
{{- end}}{{/* define "FieldGetOrSet" */}}
`
