#Templates

Simple html template package for Go\n
Replace ${{key}} with value in html file


#Godoc output:
PACKAGE DOCUMENTATION

package templates\n
    import "github.com/jmu0/templates"


TYPES

type Template struct {\n
    Path string/n
    Data map[string]string\n
    // contains filtered or unexported fields
}

func (t *Template) Load(path string) error\n
    load template html

func (t *Template) Render() (string, error)\n
    render template, return html


