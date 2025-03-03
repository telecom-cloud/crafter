package golang

var constants = `
{{define "Constants"}}
const {{.Name}} {{.Type.ResolveName ROOT}} = {{.Value.Expression}}
{{end}}
`
