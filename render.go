package render

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"io/ioutil"
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

	UseCache bool
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

	ContentTypeText   = "text/plain"
	ContentTypeHTML   = "text/html"
	ContentTypeXHTML  = "application/xhtml+xml"
	ContentTypeXML    = "text/xml"
	ContentTypeJSON   = "application/json"
	ContentTypeBinary = "application/octet-stream"

	CharsetUTF8 = "UTF-8"

	DefaultTextContentType = ContentTypeText + "; charset=" + CharsetUTF8
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

	if o != nil {
		trySetOption(o.RootDirectory, &options.RootDirectory)
		trySetOption(o.DefaultLayout, &options.DefaultLayout)
		trySetOption(o.DefaultExtension, &options.DefaultExtension)
		trySetOption(o.DefaultCharset, &options.DefaultCharset)

		options.UseCache = o.UseCache

		for k, v := range o.Funcs {
			options.Funcs[k] = v
		}
	}

	options.Funcs["extends"] = extends
	options.Funcs["render"] = render
	options.Funcs["renderEach"] = renderEach
	options.Funcs["formatTime"] = formatTime
	options.Funcs["formatFloat"] = formatFloat
	options.Funcs["formatInt"] = formatInt
}

func trySetOption(value string, option *string) {

	if s := strings.TrimSpace(value); s > "" {
		*option = s
	}
}

func getTemplateFromCache(tmplPath string) *TemplateValue {

	var tmplVal *TemplateValue

	tmplCache.RLock()
	defer tmplCache.RUnlock()

	tmplVal, _ = tmplCache.Map[tmplPath]

	return tmplVal
}

func cacheTemplate(tmplPath string, tmplVal *TemplateValue) {

	tmplCache.Lock()
	defer tmplCache.Unlock()

	tmplCache.Map[tmplPath] = tmplVal
}

func getTemplatesPaths(templatePath string) ([]string, error) {

	if ext := filepath.Ext(templatePath); ext == "" {
		templatePath += options.DefaultExtension
	}

	absPath, err := filepath.Abs(templatePath)

	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(absPath)

	if err != nil {
		return nil, err
	}

	layoutFile := getLayoutFile(content)

	if layoutFile != nil {

		layoutFilePath := filepath.Join(options.RootDirectory, string(layoutFile))

		if ext := filepath.Ext(layoutFilePath); ext == "" {
			layoutFilePath += options.DefaultExtension
		}

		lps, err := getTemplatesPaths(layoutFilePath)

		if err != nil {
			return nil, err
		}

		return append(lps, absPath), nil
	}

	return []string{absPath}, nil
}

func getLayoutFile(bs []byte) []byte {

	match := layoutPattern.FindSubmatch(bs)

	if len(match) > 0 {
		return match[1]
	}

	return nil
}

func Nothing(w http.ResponseWriter, code int) {

	w.WriteHeader(code)

	w.Write(nil)
}

func Plain(w http.ResponseWriter, s string) {

	w.Header().Set(ContentType, DefaultTextContentType)

	w.Write([]byte(s))
}

func HTML(w http.ResponseWriter, fpath string, model interface{}) {

	ct := options.DefaultHTMLContentType + "; charset=" + options.DefaultCharset
	w.Header().Set(ContentType, ct)

	if err := ExecuteHTML(w, fpath, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ExecuteHTML(wr io.Writer, fpath string, model interface{}) error {

	tmplPath := filepath.Join(options.RootDirectory, fpath)

	var tmplVal *TemplateValue

	if options.UseCache {
		tmplVal = getTemplateFromCache(tmplPath)
	}

	if tmplVal == nil {

		templatesPaths, err := getTemplatesPaths(tmplPath)
		
		if err != nil {
			return err
		}

		tmpl := template.New("root")

		if options.Funcs != nil {
			tmpl = tmpl.Funcs(options.Funcs)
		}

		tmpl, err = tmpl.ParseFiles(templatesPaths...)

		if err != nil {
			return err
		}

		tmplVal = &TemplateValue{
			name:     path.Base(templatesPaths[0]),
			template: tmpl,
		}

		if options.UseCache {
			cacheTemplate(tmplPath, tmplVal)
		}
	}

	if err := tmplVal.template.ExecuteTemplate(wr, tmplVal.name, model); err != nil {
		return err
	}

	return nil
}

func JSON(w http.ResponseWriter, model interface{}) {

	w.Header().Set(ContentType, DefaultJSONContentType)

	if err := EncodeJSON(w, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func EncodeJSON(w io.Writer, model interface{}) error {

	return json.NewEncoder(w).Encode(model)
}

func XML(w http.ResponseWriter, model interface{}) {

	w.Header().Set(ContentType, DefaultXMLContentType)

	if err := EncodeXML(w, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func EncodeXML(w io.Writer, model interface{}) error {

	return xml.NewEncoder(w).Encode(model)
}

func File(w http.ResponseWriter, r *http.Request, fpath string) {

	w.Header().Set(ContentType, ContentTypeBinary)

	http.ServeFile(w, r, fpath)
}

