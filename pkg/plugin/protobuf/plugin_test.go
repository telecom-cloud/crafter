package protobuf

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/telecom-cloud/crafter/pkg/meta"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestPlugin_Handle(t *testing.T) {
	in, err := ioutil.ReadFile("../testdata/request_protoc.out")
	if err != nil {
		t.Fatal(err)
	}

	req := &pluginpb.CodeGeneratorRequest{}
	err = proto.Unmarshal(in, req)
	if err != nil {
		t.Fatalf("unmarshal stdin request error: %v", err)
	}

	// prepare args
	plu := &Plugin{}
	plu.setLogger()
	args, _ := plu.parseArgs(*req.Parameter)

	plu.Handle(req, args)
	plu.warningLogger()
}

func TestFixModelPathAndPackage(t *testing.T) {
	plu := &Plugin{}
	plu.Package = "telecom-cloud/crafter"
	plu.ModelDir = meta.ModelDir
	// default model dir
	ret1 := [][]string{
		{"a/b/c", "telecom-cloud/crafter/biz/model/a/b/c"},
		{"biz/model/a/b/c", "telecom-cloud/crafter/biz/model/a/b/c"},
		{"telecom-cloud/crafter/a/b/c", "telecom-cloud/crafter/biz/model/a/b/c"},
		{"telecom-cloud/crafter/biz/model/a/b/c", "telecom-cloud/crafter/biz/model/a/b/c"},
	}
	for _, r := range ret1 {
		tmp := r[0]
		if !strings.Contains(tmp, plu.Package) {
			if strings.HasPrefix(tmp, "/") {
				tmp = plu.Package + tmp
			} else {
				tmp = plu.Package + "/" + tmp
			}
		}
		result, _ := plu.fixModelPathAndPackage(tmp)
		if result != r[1] {
			t.Fatalf("want go package: %s, but get: %s", r[1], result)
		}
	}

	plu.ModelDir = "model_test"
	// customized model dir
	ret2 := [][]string{
		{"a/b/c", "telecom-cloud/crafter/model_test/a/b/c"},
		{"model_test/a/b/c", "telecom-cloud/crafter/model_test/a/b/c"},
		{"telecom-cloud/crafter/a/b/c", "telecom-cloud/crafter/model_test/a/b/c"},
		{"telecom-cloud/crafter/model_test/a/b/c", "telecom-cloud/crafter/model_test/a/b/c"},
	}
	for _, r := range ret2 {
		tmp := r[0]
		if !strings.Contains(tmp, plu.Package) {
			if strings.HasPrefix(tmp, "/") {
				tmp = plu.Package + tmp
			} else {
				tmp = plu.Package + "/" + tmp
			}
		}
		result, _ := plu.fixModelPathAndPackage(tmp)
		if result != r[1] {
			t.Fatalf("want go package: %s, but get: %s", r[1], result)
		}
	}
}
