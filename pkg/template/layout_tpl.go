package template

import "path/filepath"

//-----------------------------------Default Layout-----------------------------------------

const (
	sp = string(filepath.Separator)

	defaultModelDir  = "biz" + sp + "model"
	defaultScriptDir = "script"
	defaultClientDir = "biz" + sp + "client"
)

var DefaultLayoutConfig = Config{
	Layouts: []Template{
		{
			Path: defaultModelDir + sp,
		},
		{
			Path:   "go.mod",
			Delims: [2]string{"{{", "}}"},
			Body:   `module {{.GoModule}}`,
		},
		{
			Path: ".gitignore",
			Body: `*.o
*.a
*.so
_obj
_test
*.[568vq]
[568vq].out
*.cgo1.go
*.cgo2.c
_cgo_defun.c
_cgo_gotypes.go
_cgo_export.*
_testmain.go
*.exe
*.exe~
*.test
*.prof
*.rar
*.zip
*.gz
*.psd
*.bmd
*.cfg
*.pptx
*.log
*nohup.out
*settings.pyc
*.sublime-project
*.sublime-workspace
!.gitkeep
.DS_Store
/.idea
/.vscode
/output
*.local.yml
dumped_crafter_remote_config.json
		  `,
		},
		{
			Path: "build.sh",
			Body: `#!/bin/bash
RUN_NAME={{.ServiceName}}
mkdir -p output/bin
cp script/* output 2>/dev/null
chmod +x output/bootstrap.sh
go build -o output/bin/${RUN_NAME}`,
		},
		{
			Path: defaultScriptDir + sp + "bootstrap.sh",
			Body: `#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)
BinaryName={{.ServiceName}}
echo "$CURDIR/bin/${BinaryName}"
exec $CURDIR/bin/${BinaryName}`,
		},
	},
}
