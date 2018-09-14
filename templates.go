package templates

import (
	// "fmt"
	// "encoding/json"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	// "horto-meo/model/query"
	"io/ioutil"
	"log"
	"reflect"

	"regexp"
	"strconv"
	"strings"
)

//Template structure
type Template struct {
	Path string
	html string
	Data map[string]interface{}
}

var localizeTag = "${{localize:"
var tagPre = "${{"
var tagPost = "}}"

//Load template html
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

//TemplateManager structure
type TemplateManager struct {
	TemplatePath     string
	Cache            map[string]Template
	LocalizationData []map[string]interface{}
}

//SetTemplatePath set template path, use when not caching
func (tm *TemplateManager) SetTemplatePath(tp string) {
	tm.TemplatePath = tp
}

//Preload templates into cache
func (tm *TemplateManager) Preload(path string) {
	tm.TemplatePath = path
	tm.Cache = make(map[string]Template)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println("ERROR: template:", err)
	} else {
		for _, f := range files {

			if f.IsDir() == false && f.Name()[:1] != "." && filepath.Ext(f.Name()) == ".html" {
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

//AddTemplate adds template to cache
func (tm *TemplateManager) AddTemplate(name, html string) error {
	path := tm.TemplatePath + "/" + name + ".html"
	t := Template{
		Path: path,
		html: html,
		Data: make(map[string]interface{}),
	}
	if t.html == "" {
		t.Load(t.Path)
	}
	if tm.Cache == nil {
		tm.Cache = make(map[string]Template)
	}
	tm.Cache[path] = t
	// log.Println("DEBUG:", tm.Cache)
	return nil
}

//LoadLocalization load localization strings from json file
func (tm *TemplateManager) LoadLocalization() error {
	var err error
	// tm.LocalizationData, err = query.GetLocalizationData()
	if err != nil {
		return err
	}
	return nil
}

//GetLocalizationData gets localization data for Locale
func (tm *TemplateManager) GetLocalizationData(locale string) []map[string]interface{} {
	ret := make([]map[string]interface{}, 0)
	for _, l := range tm.LocalizationData {
		if l["Locale"].(string) == locale {
			ret = append(ret, l)
		}
	}
	return ret
}

//GetTemplates gets templates
func (tm *TemplateManager) GetTemplates() map[string]string {
	if len(tm.Cache) == 0 {
		tm.Preload(tm.TemplatePath)
	}
	ret := make(map[string]string)
	for k, t := range tm.Cache {
		name := strings.Replace(k, tm.TemplatePath, "", -1) //Removes path from name
		name = name[1:]                                     //removes leading /
		name = strings.Replace(name, ".html", "", -1)       //removes .html from name
		ret[name] = t.html
	}
	return ret
}

//GetTemplateJSON gets template GetCartJSON
func (tm *TemplateManager) GetTemplateJSON() ([]byte, error) {
	return json.Marshal(tm.GetTemplates())
}

//ServeTemplateJSON serves all templates as json
func (tm *TemplateManager) ServeTemplateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	bytes, err := tm.GetTemplateJSON()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Write(bytes)
}

//ClearCache clears the cache
func (tm *TemplateManager) ClearCache() {
	tm.Cache = make(map[string]Template)
}

//Render template, return html
func (tm *TemplateManager) Render(t *Template, locale string) (string, error) {
	var rendered = t.html
	var arrHTML string
	var err, err2 error
	var tmpl Template
	if len(t.Data) > 0 {
		//Replace ${{}} tags with data values
		for key, value := range t.Data {
			switch value.(type) {
			case []map[string]string: //handle array, recursive
				tmpl, err = tm.GetTemplate(key)
				if err != nil {
					//try template.key Name
					newname := strings.Split(t.Path, "/")[len(strings.Split(t.Path, "/"))-1]
					tmpl, err2 = tm.GetTemplate(newname)
					if err2 != nil {
						log.Println("Template error:", err)
					}
				}
				arrHTML = ""
				for _, v := range value.([]map[string]string) {
					tmpl.Data = convert(v)
					var res string
					res, err = tm.Render(&tmpl, locale)
					if err == nil {
						arrHTML += res
					}
				}
				rendered = strings.Replace(rendered, tagPre+key+tagPost, arrHTML, -1)
			case []map[string]interface{}: //handle array, recursive
				tmpl, err := tm.GetTemplate(key)
				if err != nil {
					//try template.key Name
					newname := strings.Split(t.Path, "/")[len(strings.Split(t.Path, "/"))-1]
					tmpl, err2 = tm.GetTemplate(newname)
					if err2 != nil {
						log.Println("Template error:", err)
					}
				}
				arrHTML = ""
				for _, v := range value.([]map[string]interface{}) {
					tmpl.Data = v
					var res string
					res, err = tm.Render(&tmpl, locale)
					if err == nil {
						arrHTML += res
					}
				}
				rendered = strings.Replace(rendered, tagPre+key+tagPost, arrHTML, -1)

			case []interface{}: //handle array, recursive
				tmpl, err := tm.GetTemplate(key)
				if err != nil {
					//try template.key Name
					newname := strings.Split(t.Path, "/")[len(strings.Split(t.Path, "/"))-1]
					tmpl, err2 = tm.GetTemplate(newname)
					if err2 != nil {
						log.Println("Template error:", err)
					}
				}
				arrHTML = ""
				for _, v := range value.([]interface{}) {
					tmpl.Data = v.(map[string]interface{})
					var res string
					res, err = tm.Render(&tmpl, locale)
					if err == nil {
						arrHTML += res
					}
				}
				rendered = strings.Replace(rendered, tagPre+key+tagPost, arrHTML, -1)
			case interface{}:
				switch reflect.TypeOf(value).Name() {
				case "int":
					rendered = strings.Replace(rendered, tagPre+key+tagPost, strconv.Itoa(value.(int)), -1)
				case "string":
					rendered = strings.Replace(rendered, tagPre+key+tagPost, value.(string), -1)
				default:
					log.Println("TODO: Handle Template data value:", key, value)
				}
			default:
				if value == nil {
					value = ""
				}
				rendered = strings.Replace(rendered, tagPre+key+tagPost, value.(string), -1)
			}
		}
	}
	//Replace ${{localize:}} tags with localized values
	for {
		if i := strings.Index(rendered, localizeTag); i > -1 {
			word := rendered[i+12 : i+strings.Index(rendered[i:], tagPost)]
			translated := tm.Translate(word, locale)
			//DEBUGlog.Println("template render: word:", word, "translated:", translated, "locale:", locale)
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

func convert(from map[string]string) map[string]interface{} {
	ret := make(map[string]interface{})
	for key, value := range from {
		ret[key] = value
	}
	return ret
}

//Translate translate word
func (tm *TemplateManager) Translate(word string, locale string) string {
	var translated = word
	isUpperCase, _ := regexp.MatchString("[A-Z]", word[:1])
	word = strings.ToLower(word)
	for _, tr := range tm.LocalizationData {
		if tr["Locale"] == locale && strings.ToLower(tr["Word"].(string)) == word {
			translated = tr["Translation"].(string)
			if isUpperCase {
				translated = strings.ToUpper(translated[:1]) + translated[1:]
			}
			return translated
		}
	}
	return translated
}

//GetTemplate get template from cache or load template
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
