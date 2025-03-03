package template

var (
	MiddlewareTplName       = "middleware.go"
	MiddlewareSingleTplName = "middleware_single.go"
	ModelTplName            = "model.go"
	HttpClientTplName       = "httpclient.go" // underlying client for client command
	ErrorTplName            = "errors.go"
	IdlClientTplName        = "idl_client.go" // client of service for quick call
	IdlGroupClientTplName   = "idl_group_client.go"
)

var templateNameSet = map[string]string{
	MiddlewareTplName:       MiddlewareTplName,
	MiddlewareSingleTplName: MiddlewareSingleTplName,
	ModelTplName:            ModelTplName,
	IdlClientTplName:        IdlClientTplName,
	HttpClientTplName:       HttpClientTplName,
	IdlGroupClientTplName:   IdlGroupClientTplName,
}

func IsDefaultPackageTpl(name string) bool {
	if _, exist := templateNameSet[name]; exist {
		return true
	}

	return false
}

var DefaultPkgConfig = Config{
	Layouts: []Template{
		// Model tpl is imported by model generator. Here only decides model directory.
		{
			Path: defaultModelDir + sp + ModelTplName,
			Body: ``,
		},
		// Client tpl is imported by client generator. Here only decides client directory.
		{
			Path:   defaultClientDir + sp + HttpClientTplName,
			Delims: [2]string{"{{", "}}"},
			Body:   httpClientTpl,
		},
		{
			Path:   defaultClientDir + sp + IdlGroupClientTplName,
			Delims: [2]string{"{{", "}}"},
			Body:   idlGroupClientTpl,
		},
		{
			Path:   defaultClientDir + sp + IdlClientTplName,
			Delims: [2]string{"{{", "}}"},
			Body:   idlClientTpl,
		},
		//{
		//	Path:   defaultClientDir + sp + ErrorTplName,
		//	Delims: [2]string{"{{", "}}"},
		//	Body:   errorTpl,
		//},
	},
}
