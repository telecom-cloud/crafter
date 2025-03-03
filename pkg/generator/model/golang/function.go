package golang

var function = `
{{define "Function"}}
func {{template "FuncBody" . -}}
{{end}}{{/* define "Function" */}}

{{define "FuncBody"}}
{{- .Name -}}(
{{- range $i, $arg := .Args -}}
{{- if gt $i 0}}, {{end -}}
{{$arg.Name}} {{$arg.Type.ResolveName ROOT}}
{{- end -}}{{/* range */}})
{{- if gt (len .Rets) 0}} ({{end -}}
{{- range $i, $ret := .Rets -}}
{{- if gt $i 0}}, {{end -}}
{{$ret.Type.ResolveName ROOT}}
{{- end -}}{{/* range */}}
{{- if gt (len .Rets) 0}}) {{end -}}{
{{.Code}}
}
{{end}}{{/* define "FuncBody" */}}
`

var method = `
{{define "Method"}}
func ({{.ReceiverName}} {{.ReceiverType.ResolveName ROOT}}) 
{{- template "FuncBody" .Function -}}
{{end}}
`
