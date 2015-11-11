package templates

import (
	// "fmt"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Template struct {
	Path string
	html string
	Data map[string]interface{}
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

type TemplateManager struct {
	TemplatePath     string
	Cache            map[string]Template
	LocalizationData map[string]map[string]string //TODO: make local
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

//load localization strings from json file
func (tm *TemplateManager) LoadLocalization(path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &tm.LocalizationData)
}

//render template, return html
func (tm *TemplateManager) Render(t *Template, locale string) (string, error) {
	var rendered string = t.html
	if len(t.Data) > 0 {
		for key, value := range t.Data {
			switch value.(type) {
			case []interface{}:
				log.Println("template array!!")
			case interface{}:
				switch reflect.TypeOf(value).Name() {
				case "int":
					rendered = strings.Replace(rendered, "${{"+key+"}}", strconv.Itoa(value.(int)), -1)
				default:
					rendered = strings.Replace(rendered, "${{"+key+"}}", value.(string), -1)
				}
			default:
				rendered = strings.Replace(rendered, "${{"+key+"}}", value.(string), -1)
			}
		}
	}
	var localizeTag string = "${{localize:"
	for {
		if i := strings.Index(rendered, localizeTag); i > -1 {
			word := rendered[i+12 : i+strings.Index(rendered[i:], "}}")]
			translated := tm.Translate(word, locale)

			break
		} else {
			break
		}
	}
	//TODO: remove unused tags ${}
	return rendered, nil
}

//translate word
func (tm *TemplateManager) Translate(word string, locale string) string {
	if _, ok := tm.LocalizationData[locale]; ok {
		if tr, ok := tm.LocalizationData[locale][word]; ok {
			return tr
		}
	}
	return word
}

//get template from cache or load template
func (tm *TemplateManager) GetTemplate(name string) (Template, error) {
	path := tm.TemplatePath + "/" + name + ".html"
	if tmpl, ok := tm.Cache[path]; ok {
		// log.Println("template from cache")
		if tmpl.Data == nil {
			tmpl.Data = make(map[string]interface{})
		}
		return tmpl, nil
	}
	tmpl := Template{}
	err := tmpl.Load(path)
	if err != nil {
		log.Println("no template:", err)
		return tmpl, err
	}
	tmpl.Data = make(map[string]interface{})
	// log.Println("new template")
	return tmpl, nil
}
