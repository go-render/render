package render

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type testModel struct {
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
		t.Fatalf("RootDirectory incorrect. Expected %q. Actual %q", opts.RootDirectory, options.RootDirectory)
	}

	if options.DefaultLayout != opts.DefaultLayout {
		t.Fatalf("DefaultLayout incorrect. Expected %q. Actual %q", opts.DefaultLayout, options.DefaultLayout)
	}

	if options.DefaultExtension != opts.DefaultExtension {
		t.Fatalf("DefaultExtension incorrect. Expected %q. Actual %q", opts.DefaultExtension, options.DefaultExtension)
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

func TestExecuteHTML(t *testing.T) {

	opts := &Options{
		RootDirectory:    ".",
		DefaultExtension: ".html",
	}

	Init(opts)

	buffer := &bytes.Buffer{}

	m := testModel{
		FieldOne: "Title",
		FieldTwo: 2,
	}

	err := ExecuteHTML(buffer, "view_test", m)

	if err != nil {
		t.Fatal(err)
	}

	expected, _ := ioutil.ReadFile("view_result_test.html")
	expectedBody := string(expected)

	actual := buffer.Bytes()
	actualBody := string(actual)

	if expectedBody != actualBody {
		t.Fatalf("Expected %q. Actual %q.", expectedBody, actualBody)
	}
}

func TestHTMLResponse(t *testing.T) {

	opts := &Options{
		RootDirectory:    ".",
		DefaultExtension: ".html",
	}

	Init(opts)

	m := testModel{
		FieldOne: "Title",
		FieldTwo: 2,
	}

	w := httptest.NewRecorder()

	HTML(w, "view_test", m)

	expected, _ := ioutil.ReadFile("view_result_test.html")
	expectedBody := string(expected)

	expectedContentType := DefaultHTMLContentType

	actualBody := w.Body.String()
	actualContentType := w.Header().Get(ContentType)

	if expectedBody != actualBody {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expectedBody, actualBody)
	}

	if expectedContentType != actualContentType {
		t.Fatalf("Content types don't match'. Expected %q. Actual %q", expectedContentType, actualContentType)
	}
}

func TestHTMLResponseWithCaching(t *testing.T) {

	opts := &Options{
		RootDirectory:    ".",
		DefaultExtension: ".html",
		UseCache: true,
	}

	Init(opts)

	m := testModel{
		FieldOne: "Title",
		FieldTwo: 2,
	}

	w := httptest.NewRecorder()

	HTML(w, "view_test", m)

	expected, _ := ioutil.ReadFile("view_result_test.html")
	expectedBody := string(expected)

	expectedContentType := DefaultHTMLContentType

	actualBody := w.Body.String()
	actualContentType := w.Header().Get(ContentType)

	if expectedBody != actualBody {
		t.Fatalf("Bodies don't match'. Expected %q. Actual %q", expectedBody, actualBody)
	}

	if expectedContentType != actualContentType {
		t.Fatalf("Content types don't match'. Expected %q. Actual %q", expectedContentType, actualContentType)
	}
}

func TestUnknownViewHTMLResponse(t *testing.T) {

	opts := &Options{
		RootDirectory:    ".",
		DefaultExtension: ".html",
	}

	Init(opts)

	m := testModel{
		FieldOne: "Title",
		FieldTwo: 2,
	}

	w := httptest.NewRecorder()
	
	view_unknown := "view_unknown"

	HTML(w, view_unknown, m)

	expectedCode := http.StatusInternalServerError
	actualCode := w.Code

	if expectedCode != actualCode {
		t.Fatalf("Expected status %d. Actual Status %d", expectedCode, actualCode)
	}

	expectedBodyEnd := fmt.Sprintf("/%s%s: no such file or directory\n", view_unknown, opts.DefaultExtension)
	actualBody := w.Body.String()

	if !strings.HasSuffix(actualBody, expectedBodyEnd) {
		t.Fatalf("Bodies don't match. Expected body end %q. Actual body %q", expectedBodyEnd, actualBody)
	}
}

func TestExecuteHTMLChildLayout(t *testing.T) {

	opts := &Options{
		RootDirectory:    ".",
		DefaultExtension: ".html",
	}

	Init(opts)

	buffer := &bytes.Buffer{}

	m := testModel{
		FieldOne: "Title",
		FieldTwo: 2,
	}

	err := ExecuteHTML(buffer, "view_child_test", m)

	if err != nil {
		t.Fatal(err)
	}

	expected, _ := ioutil.ReadFile("view_result_test.html")
	expectedBody := string(expected)

	actual := buffer.Bytes()
	actualBody := string(actual)

	if expectedBody != actualBody {
		t.Fatalf("Expected %q. Actual %q.", expectedBody, actualBody)
	}
}

func TestEncodeXML(t *testing.T) {

	w := &bytes.Buffer{}

	expected := "<testModel><FieldOne>fieldOne</FieldOne><FieldTwo>222</FieldTwo></testModel>"

	var m = testModel{
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

	expectedBody := "<testModel><FieldOne>fieldOne</FieldOne><FieldTwo>2</FieldTwo></testModel>"
	expectedContentType := DefaultXMLContentType

	var m = testModel{
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

	var m = testModel{
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

	var m = testModel{
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

func TestFormatTime(t *testing.T) {

	tm := time.Date(2014, time.September, 9, 22, 25, 17, 45, time.Local)

	layout := "02 January 2006 15:04"

	expected := "09 September 2014 22:25"
	actual := formatTime(tm, layout)

	if expected != actual {
		t.Fatalf("Expected '%s'. Actual '%s'", expected, actual)
	}
}

func TestFormatFloat(t *testing.T) {

	f := 1254.35987

	expected := "1254.36"
	actual := formatFloat(f, 2)

	if expected != actual {
		t.Fatalf("Expected '%s'. Actual '%s'", expected, actual)
	}
}

func TestFormatInt(t *testing.T) {

	var i int64 = 987654

	expected := "11110001001000000110"
	actual := formatInt(i, 2)

	if expected != actual {
		t.Fatalf("Expected '%s'. Actual '%s'", expected, actual)
	}
}

