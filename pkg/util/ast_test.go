package util

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"golang.org/x/tools/go/ast/astutil"
)

func TestAddImport(t *testing.T) {
	inserts := [][]string{
		{
			"ctx",
			"context",
		},
		{
			"",
			"context",
		},
	}
	files := [][]string{
		{
			`package foo

import (
	"fmt"
	"time"
)
`,
			`package foo

import (
	ctx "context"
	"fmt"
	"time"
)
`,
		},
		{
			`package foo

import (
	"fmt"
	"time"
)
`,
			`package foo

import (
	"context"
	"fmt"
	"time"
)
`,
		},
	}
	for idx, file := range files {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", file[0], parser.ImportsOnly)
		if err != nil {
			t.Fatalf("can not parse ast for file")
		}
		astutil.AddNamedImport(fset, f, inserts[idx][0], inserts[idx][1])
		var output []byte
		buffer := bytes.NewBuffer(output)
		err = format.Node(buffer, fset, f)
		if err != nil {
			t.Fatalf("can add import for file")
		}
		if buffer.String() != file[1] {
			t.Fatalf("insert import fialed")
		}
	}
}
