package render

import (
    _ "fmt"
	"io/ioutil"
	_ "strings"
	"testing"
)

func TestExtendMustReturnAnEmptyString(t *testing.T) {

	s := "test_layout"

	if n := extends(s); n != "" {
		t.Fatalf("`extends` should return an empty string. Actual %q", n)
	}
}

func TestRender(t *testing.T) {
    
	opts := &Options{
		RootDirectory:    ".",
		DefaultExtension: ".html",
	}
	
    Init(opts)

	m := testModel{
		FieldOne: "Name",
		FieldTwo: 18,
	}
	
	actual, err := render("partial_test", m)
	
	if err != nil {
	    t.Fatalf("Expected no error. Actual error: %s", err.Error())
	}
	
	expected, _ := ioutil.ReadFile("partial_result_test.html")
	expectedContent := string(expected)

	actualContent := string(actual)

	if expectedContent != actualContent {
		t.Fatalf("Expected %q. Actual %q.", expectedContent, actualContent)
	}
}

func TestRenderWithCache(t *testing.T) {
    
	opts := &Options{
		RootDirectory:    ".",
		DefaultExtension: ".html",
		UseCache: true,
	}
	
    Init(opts)

	m := testModel{
		FieldOne: "Name",
		FieldTwo: 18,
	}
	
	actual, err := render("partial_test", m)
	
	if err != nil {
	    t.Fatalf("Expected no error. Actual error: %s", err.Error())
	}
	
	expected, _ := ioutil.ReadFile("partial_result_test.html")
	expectedContent := string(expected)

	actualContent := string(actual)

	if expectedContent != actualContent {
		t.Fatalf("Expected %q. Actual %q.", expectedContent, actualContent)
	}
}

func TestRenderUnknownPartial(t *testing.T) {
    
	opts := &Options{
		RootDirectory:    ".",
		DefaultExtension: ".html",
	}
	
    Init(opts)

	m := testModel{
		FieldOne: "Name",
		FieldTwo: 18,
	}
	
	partial_unknown := "partial_unknown"
	
	actualContent, err := render(partial_unknown, m)
	
	if err == nil {
	    t.Fatalf("Expected an error.")
	}

	if actualContent != "" {
	    t.Fatalf("Expected no content. Actual content %q", actualContent)
	}
}


