package render

import (
	"bytes"
	"html/template"
	"path/filepath"
	"reflect"
)

func extends(s string) string {

	return ""
}

func partial(path string, model interface{}) (template.HTML, error) {

	var tmplVal *TemplateValue

	if options.UseCache {
		tmplVal = getTemplateFromCache(path)
	}

	if tmplVal == nil {

		tmplPath, err := filepath.Abs(filepath.Join(options.RootDirectory, path))

		if err != nil {
			return template.HTML(""), err
		}

		if ext := filepath.Ext(tmplPath); ext == "" {
			tmplPath += options.DefaultExtension
		}

		tmpl, err := template.New(tmplPath).Funcs(options.Funcs).ParseFiles(tmplPath)

		if err != nil {
			return template.HTML(""), err
		}

		tmplVal = &TemplateValue{
			name:     filepath.Base(tmplPath),
			template: tmpl,
		}

		if options.UseCache {
			cacheTemplate(tmplPath, tmplVal)
		}
	}

	buf := &bytes.Buffer{}

	if err := tmplVal.template.ExecuteTemplate(buf, tmplVal.name, model); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(string(buf.Bytes())), nil
}

func partials(path string, col interface{}) (template.HTML, error) {

	var html template.HTML

	s := reflect.ValueOf(col)

	switch s.Type().Kind() {

	case reflect.Slice, reflect.Array:

		for i := 0; i < s.Len(); i++ {

			v := s.Index(i).Interface()
			h, err := partial(path, v)

			if err != nil {
				return template.HTML(""), err
			}

			html += h
		}

	case reflect.Map:

		kv := struct{ Key, Value interface{} }{}

		for _, k := range s.MapKeys() {

			kv.Key = k.Interface()
			kv.Value = s.MapIndex(k).Interface()

			h, err := partial(path, kv)

			if err != nil {
				return template.HTML(""), err
			}

			html += h
		}
	}

	return html, nil
}

