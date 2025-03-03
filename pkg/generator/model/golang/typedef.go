package golang

// Typedef .
var typedef = `
{{define "Typedef"}}
{{- $NewTypeName := (Identify .Alias)}}
{{- $OldTypeName := .Type.ResolveNameForTypedef ROOT}}
type {{$NewTypeName}} = {{$OldTypeName}}

{{if eq .Type.Kind 25}}{{if .Type.HasNew}}
func New{{$NewTypeName}}() *{{$NewTypeName}} {
	return {{(GetTypedefReturnStr $OldTypeName)}}
}
{{- end}}{{- end}}
{{- end}}
`
