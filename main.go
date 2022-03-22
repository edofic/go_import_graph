package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/goccy/go-graphviz"
)

func main() {
	rootFolder := flag.String("root", ".", "directory root to parse")
	baseModule := flag.String("module", "", "base module to filter for")
	targetFlag := flag.String("target", "", "output file, new is created for blank")
	open := flag.Bool("open", true, "open the output file once rendered")
	flag.Parse()

	if *rootFolder == "" {
		panic("root is required")
	}

	if *baseModule == "" {
		panic("module is required")
	}

	target := *targetFlag
	if target == "" {
		var err error
		target, err = createNewFile()
		if err != nil {
			panic(err)
		}
	}

	log.Println("rendering to", target)

	err := run(*rootFolder, *baseModule, target, *open)
	if err != nil {
		log.Fatal(err)
	}
}

func run(rootFolder, baseModule, target string, open bool) error {
	byPkg, err := searchTree(rootFolder, baseModule)
	if err != nil {
		return err
	}
	g := graphviz.New()
	graph, err := graphviz.ParseBytes(prepareDot(byPkg))
	if err != nil {
		return err
	}

	err = g.RenderFilename(graph, graphviz.PNG, target)
	if err != nil {
		log.Fatal(err)
	}

	if open {
		systemOpen(target)
	}
	return nil
}

func createNewFile() (string, error) {
	file, err := ioutil.TempFile("", "graph*.png")
	if err != nil {
		return "", err
	}
	file.Close()
	return file.Name(), nil
}

func prepareDot(byPkg map[string][]string) []byte {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "digraph G {\n")
	for pkg, imports := range byPkg {
		for _, import_ := range imports {
			fmt.Fprintf(buf, " \"%s\" -> \"%s\"\n", pkg, import_)
		}
	}
	fmt.Fprintf(buf, "}")
	return buf.Bytes()
}

func searchTree(root string, baseModule string) (map[string][]string, error) {
	byPkg := make(map[string][]string)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		imports, err := searchFile(path, baseModule)
		if err != nil {
			return err
		}
		byPkg[strings.TrimPrefix(filepath.Dir(path), root+"/")] = imports
		return nil
	})
	if err != nil {
		return nil, err
	}
	return byPkg, nil
}

func searchFile(filename string, baseModule string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ast, err := decorator.Parse(file)
	if err != nil {
		return nil, err
	}
	visitor := &visitor{base: baseModule}
	dst.Walk(visitor, ast)
	return visitor.imports, nil
}

type visitor struct {
	base    string
	imports []string
}

func (v *visitor) Visit(n dst.Node) dst.Visitor {
	spec, ok := n.(*dst.ImportSpec)
	if !ok {
		return v
	}
	target := spec.Path.Value
	if strings.Contains(target, v.base) {
		target = strings.Trim(target, `"`)
		target = strings.TrimPrefix(target, v.base+"/")
		v.imports = append(v.imports, target)
	}
	return v
}

func systemOpen(url string) {
	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	default:
	}
}
