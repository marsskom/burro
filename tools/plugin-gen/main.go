package main

import (
	"os"
	"text/template"
)

const outFile = "../../plugins/registry/all_plugins_gen.go"

func main() {
	pluginsDir := "../../plugins"

	entries, _ := os.ReadDir(pluginsDir)

	var imports []string

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		if e.Name() == "registry" {
			continue
		}

		imports = append(imports, `_ "gitlab.com/marsskom/burro/plugins/`+e.Name()+`"`)
	}

	tpl := `package registry

import (
{{range .}}	{{.}}
{{end}}
)
`
	f, _ := os.Create(outFile)
	defer f.Close()

	template.Must(template.New("").Parse(tpl)).Execute(f, imports)
}
