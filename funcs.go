package render

import (
	"bytes"
	"html/template"
	"path/filepath"
)

func extends(s string) string {

	return ""
}

func partial(path string, model interface{}) (template.HTML, error) {

	var tmplVal *TemplateValue
	var ok bool

	tmplCache.RLock()
	tmplVal, ok = tmplCache.Map[path]
	tmplCache.RUnlock()

	if !ok {

		tmplPath, err := filepath.Abs(filepath.Join(options.RootDirectory, path))

		if err != nil {
			return template.HTML(""), err
		}

		if ext := filepath.Ext(tmplPath); ext == "" {
			tmplPath += options.DefaultExtension
		}

		tmpl := template.Must(template.New(tmplPath).Funcs(options.Funcs).ParseFiles(tmplPath))

		tmplVal = &TemplateValue{
			name:     filepath.Base(tmplPath),
			template: tmpl,
		}

		if _, ok = tmplCache.Map[tmplPath]; !ok {
			tmplCache.Lock()
			tmplCache.Map[path] = tmplVal
			tmplCache.Unlock()
		}
	}

	buf := &bytes.Buffer{}

	if err := tmplVal.template.ExecuteTemplate(buf, tmplVal.name, model); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(string(buf.Bytes())), nil
}

