package render

import (
	"bytes"
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
	JSONIndent             string
	XMLIndent              string
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
)

var (
	layoutPattern  = regexp.MustCompile(`\{\{\s*extends\s+"((?:\.{1,2}/)*[^"]+)"\s*\}\}`)
	partialPattern = regexp.MustCompile(`\{\{\s*template\s+"((?:\.{1,2}/)*_[^"]+)"(?:\s+\.)?\s*\}\}`)

	options = &Options{
		RootDirectory:          "views",
		DefaultLayout:          "_layout",
		DefaultExtension:       ".html",
		DefaultHTMLContentType: ContentTypeHTML,
		JSONIndent:             "",
		XMLIndent:              "",
		DefaultCharset:         CharsetUTF8,

		Funcs: template.FuncMap{
			"extends": func(s string) string {
				return ""
			},
		},
	}

	tmplCache = TemplateCache{Map: TemplateMap{}, RWMutex: &sync.RWMutex{}}
)

func Init(o *Options) {

	if o != nil {
		trySetOption(o.RootDirectory, &options.RootDirectory)
		trySetOption(o.DefaultLayout, &options.DefaultLayout)
		trySetOption(o.DefaultExtension, &options.DefaultExtension)
		trySetOption(o.DefaultCharset, &options.DefaultCharset)
		trySetOption(o.JSONIndent, &options.JSONIndent)
		trySetOption(o.XMLIndent, &options.XMLIndent)

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

func getTemplatesPaths(templatePath string) ([]string, map[string]string) {

	if ext := path.Ext(templatePath); ext == "" {
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

	partialsPaths := getPartialsPaths(absPath, content)

	layoutFile := getLayoutFile(content)

	if layoutFile != nil {

		layoutFilePath := filepath.Join(filepath.Dir(absPath), string(layoutFile))

		if ext := path.Ext(layoutFilePath); ext == "" {
			layoutFilePath += options.DefaultExtension
		}

		lps, pps := getTemplatesPaths(layoutFilePath)

		return append(lps, absPath), appendMaps(pps, partialsPaths)
	}

	return []string{absPath}, partialsPaths
}

func getLayoutFile(bs []byte) []byte {

	match := layoutPattern.FindSubmatch(bs)

	if len(match) > 0 {
		return match[1]
	}

	return nil
}

func getPartialsPaths(templatePath string, bs []byte) map[string]string {

	partialPaths := make(map[string]string)

	matches := partialPattern.FindAllSubmatch(bs, -1)

	for _, bss := range matches {

		partialName := string(bss[1])
		partialPath, err := filepath.Abs(filepath.Join(filepath.Dir(templatePath), partialName))

		if err != nil {
			log.Fatal(err)
		}

		if ext := path.Ext(partialPath); ext == "" {
			partialPath += options.DefaultExtension
		}

		partialPaths[partialName] = partialPath
	}

	return partialPaths
}

func appendMaps(m1 map[string]string, m2 map[string]string) map[string]string {

	m := make(map[string]string)

	mergeFunc := func(_m map[string]string) {

		for k, v := range _m {
			m[k] = v
		}
	}

	mergeFunc(m1)
	mergeFunc(m2)

	return m
}

func HTML(w http.ResponseWriter, filepath string, model interface{}) {

	if ct := w.Header().Get(ContentType); ct == "" {
		ct = options.DefaultHTMLContentType + "; charset=" + options.DefaultCharset
		w.Header().Set(ContentType, ct)
	}

	if err := RenderHTML(w, filepath, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func RenderHTML(wr io.Writer, filepath string, model interface{}) error {

	tmplPath := path.Join(options.RootDirectory, filepath)

	var tmplVal *TemplateValue
	var ok bool

	tmplCache.RLock()
	tmplVal, ok = tmplCache.Map[tmplPath]
	tmplCache.RUnlock()

	if !ok {

		templatesPaths, partialsPaths := getTemplatesPaths(tmplPath)

		tmpl := template.New("root")

		if options.Funcs != nil {
			tmpl = tmpl.Funcs(options.Funcs)
		}

		tmpl = template.Must(tmpl.ParseFiles(templatesPaths...))

		for k, v := range partialsPaths {

			bs, err := ioutil.ReadFile(v)

			if err != nil {
				return err
			}

			_, err = tmpl.New(k).Parse(string(bs))

			if err != nil {
				return err
			}
		}

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

	buf := &bytes.Buffer{}

	if err := tmplVal.template.ExecuteTemplate(buf, tmplVal.name, model); err != nil {
		return err
	}

	io.Copy(wr, buf)

	return nil
}

func JSON(w http.ResponseWriter, model interface{}) {

	if ct := w.Header().Get(ContentType); ct == "" {
		ct = ContentTypeJSON + "; charset=" + options.DefaultCharset
		w.Header().Set(ContentType, ct)
	}

	buf, err := MarshalJSON(model)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buf)
}

func MarshalJSON(model interface{}) ([]byte, error) {

	if options.JSONIndent != "" {
		return json.MarshalIndent(model, "", options.JSONIndent)
	}

	return json.Marshal(model)
}

func XML(w http.ResponseWriter, model interface{}) {

	if ct := w.Header().Get(ContentType); ct == "" {
		ct = ContentTypeXML + "; charset=" + options.DefaultCharset
		w.Header().Set(ContentType, ct)
	}

	buf, err := MarshalXML(model)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buf)
}

func MarshalXML(model interface{}) ([]byte, error) {

	if options.XMLIndent != "" {
		return xml.MarshalIndent(model, "", options.XMLIndent)
	}

	return xml.Marshal(model)
}

