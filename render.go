package render

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Options struct {
	RootDirectory    string
	DefaultLayout    string
	DefaultExtension string
	Funcs            template.FuncMap
}

type TemplateValue struct {
	Name     string
	Template *template.Template
}

var (
	layoutPattern  = regexp.MustCompile(`^\s*\{\{/\*\s*Layout\s*=\s*"([^"]+)"\s*\*/\}\}`)
	partialPattern = regexp.MustCompile(`\{\{\s*template\s+"((?:\.{1,2}/)*_[^"]+)"(?:\s+\.)?\s*\}\}`)

	options = &Options{
		RootDirectory:    "views",
		DefaultLayout:    "_layout",
		DefaultExtension: ".html",
		Funcs:            template.FuncMap{},
	}

	tmplCache = make(map[string]*TemplateValue)
)

func Init(o *Options) {

	if o != nil {
		trySetOption(o.RootDirectory, &options.RootDirectory)
		trySetOption(o.DefaultLayout, &options.DefaultLayout)
		trySetOption(o.DefaultExtension, &options.DefaultExtension)

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

func GetTemplatesPaths(templatePath string) ([]string, map[string]string) {

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

	partialsPaths := GetPartialsPaths(absPath, content)

	layoutFile := GetLayoutFile(content)

	if layoutFile != nil {

		layoutFilePath := filepath.Join(filepath.Dir(absPath), string(layoutFile))

		if ext := path.Ext(layoutFilePath); ext == "" {
			layoutFilePath += options.DefaultExtension
		}

		lps, pps := GetTemplatesPaths(layoutFilePath)

		return append(lps, absPath), AppendMaps(pps, partialsPaths)
	}

	return []string{absPath}, partialsPaths
}

func GetLayoutFile(bs []byte) []byte {

	match := layoutPattern.FindSubmatch(bs)

	if len(match) > 0 {
		return match[1]
	}

	return nil
}

func GetPartialsPaths(templatePath string, bs []byte) map[string]string {

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

func AppendMaps(m1 map[string]string, m2 map[string]string) map[string]string {

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

func Render(w io.Writer, filepath string, model interface{}) {

	tmplPath := path.Join(options.RootDirectory, filepath)

	tmplVal, ok := tmplCache[tmplPath]

	if !ok {

		templatesPaths, partialsPaths := GetTemplatesPaths(tmplPath)

		tmpl := template.New("root")

		if options.Funcs != nil {

			tmpl = tmpl.Funcs(options.Funcs)
		}

		tmpl = template.Must(tmpl.ParseFiles(templatesPaths...))

		for k, v := range partialsPaths {

			bs, _ := ioutil.ReadFile(v)
			_ = template.Must(tmpl.New(k).Parse(string(bs)))
		}

		tmplVal = &TemplateValue{
			Name:     path.Base(templatesPaths[0]),
			Template: tmpl,
		}

		tmplCache[tmplPath] = tmplVal
	}

	tmplVal.Template.ExecuteTemplate(w, tmplVal.Name, model)
}

