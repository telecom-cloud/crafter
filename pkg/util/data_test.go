package util

import "testing"

func TestUniqueName(t *testing.T) {
	type UniqueName struct {
		Name         string
		ExpectedName string
		ActualName   string
	}

	nameList := []UniqueName{
		{
			Name:         "aaa",
			ExpectedName: "aaa",
		},
		{
			Name:         "aaa",
			ExpectedName: "aaa0",
		},
		{
			Name:         "aaa0",
			ExpectedName: "aaa00",
		},
		{
			Name:         "aaa0",
			ExpectedName: "aaa01",
		},
		{
			Name:         "aaa00",
			ExpectedName: "aaa000",
		},
		{
			Name:         "aaa",
			ExpectedName: "aaa1",
		},
		{
			Name:         "aaa",
			ExpectedName: "aaa2",
		},
		{
			Name:         "aaa",
			ExpectedName: "aaa3",
		},
		{
			Name:         "aaa",
			ExpectedName: "aaa4",
		},
	}
	for _, name := range nameList {
		name.ActualName, _ = getUniqueName(name.Name, uniquePackageName)
		if name.ActualName != name.ExpectedName {
			t.Errorf("%s name expected unique name '%s', actually get '%s'", name.Name, name.ExpectedName, name.ActualName)
		}
	}
	for _, name := range nameList {
		name.ActualName, _ = getUniqueName(name.Name, uniqueMiddlewareName)
		if name.ActualName != name.ExpectedName {
			t.Errorf("%s name expected unique name '%s', actually get '%s'", name.Name, name.ExpectedName, name.ActualName)
		}
	}
}
