package templates

import (
	// "fmt"
	"errors"
	"io/ioutil"
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
	var rendered string = t.html
	if len(t.Data) > 0 {
		for key, value := range t.Data {
			rendered = strings.Replace(rendered, "${{"+key+"}}", value, -1)
		}
	}
	return rendered, nil
}
