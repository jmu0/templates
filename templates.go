package templates

import (
	// "fmt"
	"errors"
	"io/ioutil"
	"log"
	"strings"
)

type Template struct {
	Path string
	html string
	Data map[string]string
}

//load template html
func (t *Template) Load(path string) error {
	var err error
	var bytes []byte
	if len(path) == 0 {
		path = t.Path
	}
	if len(path) == 0 {
		return errors.New("No path for template given")
	}
	bytes, err = ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	t.html = string(bytes)
	return nil
}

//render template, return html
func (t *Template) Render() (string, error) {
	//TODO: remove unused tags ${{}}
	var rendered string = t.html
	if len(t.Data) > 0 {
		for key, value := range t.Data {
			rendered = strings.Replace(rendered, "${{"+key+"}}", value, -1)
		}
	}
	return rendered, nil
}

type TemplateManager struct {
	TemplatePath string
	Cache        map[string]Template
}

//preload templates into cache
func (tm *TemplateManager) Preload(path string) {
	tm.TemplatePath = path
	tm.Cache = make(map[string]Template)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err)
	} else {
		for _, f := range files {
			if f.IsDir() == false {
				tPath := path + "/" + f.Name()
				templ := Template{
					Path: tPath,
				}
				if templ.Load("") == nil {
					tm.Cache[tPath] = templ
					log.Println("Preloading:", tPath)
				}
			}
		}
	}
}

//get template from cache or load template
func (tm *TemplateManager) GetTemplate(name string) (Template, error) {
	path := tm.TemplatePath + "/" + name + ".html"
	if tmpl, ok := tm.Cache[path]; ok {
		log.Println("template from cache")
		return tmpl, nil
	}
	tmpl := Template{}
	err := tmpl.Load(path)
	if err != nil {
		log.Println("no template:", err)
		return tmpl, err
	}
	log.Println("new template")
	return tmpl, nil
}
