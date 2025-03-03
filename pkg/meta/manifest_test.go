package meta

import (
	"fmt"
	"testing"

	gv "github.com/hashicorp/go-version"
)

func TestValidate(t *testing.T) {
	_, err := gv.NewVersion(Version)
	if err != nil {
		t.Fatalf("not a valid version: %s", err)
	}
}

func TestGeneratedJson(t *testing.T) {
	original := &GeneratedJSON{
		ServiceGroup: "eci",
		Module:       "github/telecom-cloud/statecloud-sdk-go",
		Clients: []string{
			"DataCache",
			"Region",
		},
	}
	filename := "generated.json"
	data, err := LoadGeneratedJson(filename, original)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(data)
}
