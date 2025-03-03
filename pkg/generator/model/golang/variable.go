package golang

var variables = `
{{- define "Variables"}}
var {{.Name}} {{.Type.ResolveName ROOT}} = {{.Value.Expression}}
{{end}}
`
