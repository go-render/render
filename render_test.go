package render

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type test_model struct {
	FieldOne string
	FieldTwo int
}

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

	expectedLayout := "desktop/layout"

	bs := []byte(fmt.Sprintf("{{ extends %q}} ", expectedLayout))

	actualLayout := string(getLayoutFile(bs))

	if expectedLayout != actualLayout {
		t.Fatalf("Layout file incorrectly read. Expected %q. Actual %q", expectedLayout, actualLayout)
	}
}

func TestNothingResponse(t *testing.T) {

	w := httptest.NewRecorder()

	expectedBody := ""
	expectedCode := http.StatusOK

	Nothing(w, expectedCode)

	actualBody := w.Body.String()
	actualCode := w.Code

	if expectedBody != actualBody {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expectedBody, actualBody)
	}

	if expectedCode != actualCode {
		t.Fatalf("Codes don't match'. Expected %q. Actual %q", expectedCode, actualCode)
	}
}

func TestTextPlainResponse(t *testing.T) {

	w := httptest.NewRecorder()

	expectedBody := "QWERTY"

	Plain(w, expectedBody)

	actualBody := w.Body.String()

	if expectedBody != actualBody {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expectedBody, actualBody)
	}
}

func TestEncodeXML(t *testing.T) {

	w := &bytes.Buffer{}

	expected := "<test_model><FieldOne>fieldOne</FieldOne><FieldTwo>222</FieldTwo></test_model>"

	var m = test_model{
		FieldOne: "fieldOne",
		FieldTwo: 222,
	}

	EncodeXML(w, m)

	actual := string(w.Bytes())

	if expected != actual {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expected, actual)
	}
}

func TestXMLResponse(t *testing.T) {

	w := httptest.NewRecorder()

	expectedBody := "<test_model><FieldOne>fieldOne</FieldOne><FieldTwo>2</FieldTwo></test_model>"
	expectedContentType := DefaultXMLContentType

	var m = test_model{
		FieldOne: "fieldOne",
		FieldTwo: 2,
	}

	XML(w, m)

	actualBody := w.Body.String()
	actualContentType := w.Header().Get(ContentType)

	if expectedBody != actualBody {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expectedBody, actualBody)
	}

	if expectedContentType != actualContentType {
		t.Fatalf("Content types don't match'. Expected %q. Actual %q", expectedContentType, actualContentType)
	}
}

func TestEncodeJSON(t *testing.T) {

	w := &bytes.Buffer{}

	expected := "{\"FieldOne\":\"fieldOne1\",\"FieldTwo\":2}\n"

	var m = test_model{
		FieldOne: "fieldOne1",
		FieldTwo: 2,
	}

	EncodeJSON(w, m)

	actual := string(w.Bytes())

	if expected != actual {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expected, actual)
	}
}

func TestJSONResponse(t *testing.T) {

	w := httptest.NewRecorder()

	expectedBody := "{\"FieldOne\":\"fieldOne1111\",\"FieldTwo\":22222}\n"
	expectedContentType := DefaultJSONContentType

	var m = test_model{
		FieldOne: "fieldOne1111",
		FieldTwo: 22222,
	}

	JSON(w, m)

	actualBody := w.Body.String()
	actualContentType := w.Header().Get(ContentType)

	if expectedBody != actualBody {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expectedBody, actualBody)
	}

	if expectedContentType != actualContentType {
		t.Fatalf("Content types don't match'. Expected %q. Actual %q", expectedContentType, actualContentType)
	}
}

func TestFile(t *testing.T) {

	f, err := ioutil.TempFile(os.TempDir(), "render-pkg-test-")

	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatal(err)
	}

	expectedBody := "The file content"
	expectedContentType := ContentTypeBinary

	_, err = f.Write([]byte(expectedBody))

	if err != nil {
		t.Fatal(err)
	}

	File(w, r, f.Name())

	actualBody := w.Body.String()
	actualContentType := w.Header().Get(ContentType)

	if expectedBody != actualBody {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expectedBody, actualBody)
	}

	if expectedContentType != actualContentType {
		t.Fatalf("Content types don't match'. Expected %q. Actual %q", expectedContentType, actualContentType)
	}
}

