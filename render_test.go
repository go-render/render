package render

import (
	"testing"
	"fmt"
)

func TestInit(t *testing.T) {

	opts := &Options{
		RootDirectory:    "root_dir",
		DefaultLayout:    "def_layout",
		DefaultExtension: ".ext",
	}

	Init(opts)

	if options.RootDirectory != opts.RootDirectory {
		t.Fatalf("RootDirectory incorrect. Expected %s. Actual %s", opts.RootDirectory, options.RootDirectory)
	}

	if options.DefaultLayout != opts.DefaultLayout {
		t.Fatalf("DefaultLayout incorrect. Expected %s. Actual %s", opts.DefaultLayout, options.DefaultLayout)
	}

	if options.DefaultExtension != opts.DefaultExtension {
		t.Fatalf("DefaultExtension incorrect. Expected %s. Actual %s", opts.DefaultExtension, options.DefaultExtension)
	}
}

func TestGetLayoutFile(t *testing.T) {

    expected := "../../layout"
   
	bs := []byte(fmt.Sprintf("{{ extends %q}} ", expected))

	actual := string(getLayoutFile(bs))

	if expected != actual {
		t.Fatalf("Layout file incorrectly read. Expected %s. Actual %s", expected, actual)
	}
}

