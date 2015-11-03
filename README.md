#Templates

Simple html template package for Go
Replace ${{key}} with value in html file


#Godoc output:
PACKAGE DOCUMENTATION

package templates
    import "github.com/jmu0/templates"


TYPES

type Template struct {

    Path string
    Data map[string]string
    // contains filtered or unexported fields
}

func (t *Template) Load(path string) error

    load template html

func (t *Template) Render() (string, error)

    render template, return html


