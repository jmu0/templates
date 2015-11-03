package templates

import (
	// "fmt"
	"errors"
)

type Template struct {
	Path string
	html string
	Data map[string]string
}

//load template html
func (t *Template) Load(path string) error {
	if len(path) == 0 {
		path = t.Path
	}
	if len(path) == 0 {
		return errors.New("No path for template given")
	}
	//TODO: load file into t.html
	return nil
}

//render template, return html
func (t *Template) Render() (string, error) {
	//TODO: render template
	return "", nil
}
