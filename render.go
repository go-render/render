package render

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Options struct {
	RootDirectory          string
	DefaultLayout          string
	DefaultExtension       string
	DefaultHTMLContentType string
	DefaultCharset         string

	Funcs template.FuncMap
}

type TemplateValue struct {
	name     string
	template *template.Template
}

type TemplateMap map[string]*TemplateValue

type TemplateCache struct {
	Map TemplateMap
	*sync.RWMutex
}

const (
	ContentType = "Content-Type"

	ContentTypeTEXT   = "text/plain"
	ContentTypeHTML   = "text/html"
	ContentTypeXHTML  = "application/xhtml+xml"
	ContentTypeXML    = "text/xml"
	ContentTypeJSON   = "application/json"
	ContentTypeBinary = "application/octet-stream"

	CharsetUTF8 = "UTF-8"

	DefaultHTMLContentType = ContentTypeHTML + "; charset=" + CharsetUTF8
	DefaultJSONContentType = ContentTypeJSON + "; charset=" + CharsetUTF8
	DefaultXMLContentType  = ContentTypeXML + "; charset=" + CharsetUTF8
)

var (
	layoutPattern = regexp.MustCompile(`\{\{\s*extends\s+"((?:\.{1,2}/)*[^"]+)"\s*\}\}`)

	options = &Options{
		RootDirectory:          "views",
		DefaultLayout:          "_layout",
		DefaultExtension:       ".html",
		DefaultHTMLContentType: ContentTypeHTML,
		DefaultCharset:         CharsetUTF8,

		Funcs: template.FuncMap{},
	}

	tmplCache = TemplateCache{Map: TemplateMap{}, RWMutex: &sync.RWMutex{}}
)

func Init(o *Options) {

	options.Funcs["extends"] = extends
	options.Funcs["partial"] = partial

	if o != nil {
		trySetOption(o.RootDirectory, &options.RootDirectory)
		trySetOption(o.DefaultLayout, &options.DefaultLayout)
		trySetOption(o.DefaultExtension, &options.DefaultExtension)
		trySetOption(o.DefaultCharset, &options.DefaultCharset)

		for k, v := range o.Funcs {
			options.Funcs[k] = v
		}
	}
}

func trySetOption(value string, option *string) {

	if s := strings.TrimSpace(value); s > "" {
		*option = s
	}
}

func getTemplatesPaths(templatePath string) []string {

	if ext := filepath.Ext(templatePath); ext == "" {
		templatePath += options.DefaultExtension
	}

	absPath, err := filepath.Abs(templatePath)

	if err != nil {
		log.Fatal(err)
	}

	content, err := ioutil.ReadFile(absPath)

	if err != nil {
		log.Fatal(err)
	}

	layoutFile := getLayoutFile(content)

	if layoutFile != nil {

		layoutFilePath := filepath.Join(filepath.Dir(absPath), string(layoutFile))

		if ext := filepath.Ext(layoutFilePath); ext == "" {
			layoutFilePath += options.DefaultExtension
		}

		lps := getTemplatesPaths(layoutFilePath)

		return append(lps, absPath)
	}

	return []string{absPath}
}

func getLayoutFile(bs []byte) []byte {

	match := layoutPattern.FindSubmatch(bs)

	if len(match) > 0 {
		return match[1]
	}

	return nil
}

func HTML(w http.ResponseWriter, filepath string, model interface{}) {

	if ct := w.Header().Get(ContentType); ct == "" {
		ct = options.DefaultHTMLContentType + "; charset=" + options.DefaultCharset
		w.Header().Set(ContentType, ct)
	}

	if err := ExecuteHTML(w, filepath, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ExecuteHTML(wr io.Writer, fpath string, model interface{}) error {

	tmplPath := filepath.Join(options.RootDirectory, fpath)

	var tmplVal *TemplateValue
	var ok bool

	tmplCache.RLock()
	tmplVal, ok = tmplCache.Map[tmplPath]
	tmplCache.RUnlock()

	if !ok {

		templatesPaths := getTemplatesPaths(tmplPath)

		tmpl := template.New("root")

		if options.Funcs != nil {
			tmpl = tmpl.Funcs(options.Funcs)
		}

		tmpl = template.Must(tmpl.ParseFiles(templatesPaths...))

		tmplVal = &TemplateValue{
			name:     path.Base(templatesPaths[0]),
			template: tmpl,
		}

		if _, ok = tmplCache.Map[tmplPath]; !ok {
			tmplCache.Lock()
			tmplCache.Map[tmplPath] = tmplVal
			tmplCache.Unlock()
		}
	}

	if err := tmplVal.template.ExecuteTemplate(wr, tmplVal.name, model); err != nil {
		return err
	}

	return nil
}

func JSON(w http.ResponseWriter, model interface{}) {

	if ct := w.Header().Get(ContentType); ct == "" {
		w.Header().Set(ContentType, DefaultJSONContentType)
	}

	if err := EncodeJSON(w, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func EncodeJSON(w io.Writer, model interface{}) error {

	return json.NewEncoder(w).Encode(model)
}

func XML(w http.ResponseWriter, model interface{}) {

	if ct := w.Header().Get(ContentType); ct == "" {
		w.Header().Set(ContentType, DefaultXMLContentType)
	}

	if err := EncodeXML(w, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func EncodeXML(w io.Writer, model interface{}) error {

	return xml.NewEncoder(w).Encode(model)
}

