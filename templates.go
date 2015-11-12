package templates

import (
	// "fmt"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Template struct {
	Path string
	html string
	Data map[string]interface{}
}

var localizeTag string = "${{localize:"
var tagPre string = "${{"
var tagPost string = "}}"

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
	LocalizationData map[string]map[string]string //TODO: make this variable local only
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
		//Replace ${{}} tags with data values
		for key, value := range t.Data {
			switch value.(type) {
			case []interface{}:
				log.Println("Template array:", key)
			case interface{}:
				switch reflect.TypeOf(value).Name() {
				case "int":
					rendered = strings.Replace(rendered, tagPre+key+tagPost, strconv.Itoa(value.(int)), -1)
				default:
					rendered = strings.Replace(rendered, tagPre+key+tagPost, value.(string), -1)
				}
			default:
				rendered = strings.Replace(rendered, tagPre+key+tagPost, value.(string), -1)
			}
		}
	}
	//Replace ${{localize:}} tags with localized values
	for {
		if i := strings.Index(rendered, localizeTag); i > -1 {
			word := rendered[i+12 : i+strings.Index(rendered[i:], tagPost)]
			isUpperCase, _ := regexp.MatchString("[A-Z]", word[:1])
			translated := tm.Translate(strings.ToLower(word), locale)
			if isUpperCase {
				translated = strings.ToUpper(translated[:1]) + translated[1:]
			}
			rendered = strings.Replace(rendered, localizeTag+word+tagPost, translated, -1)
		} else {
			break
		}
	}
	//Remove unused tags ${}
	for {
		if i := strings.Index(rendered, tagPre); i > -1 {
			tag := rendered[i : i+strings.Index(rendered[i:], tagPost)+2]
			if len(tag) < 5 {
				tag = tagPre
			}
			//DEBUG: log.Println("Unused tag:", tag)
			rendered = strings.Replace(rendered, tag, "", -1)
		} else {
			break
		}
	}
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
		//DEBUG log.Println("template from cache")
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
	//DEBUG log.Println("new template")
	return tmpl, nil
}
