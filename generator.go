package safebsoncodecgen

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	_ "embed"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

//go:embed bsonRegisterType.go.tpl
var bsonRegisterType string

type BSONGenerator struct {
	handledPackages []string
	config          *Config
}

type Config struct {
	PackageName string
	PackagePath string
	Direction   string
	FileName    string
}

func (c *Config) GetFileName() string {
	if c.FileName == "" {
		return "safe_option_bson_register.go"
	}
	return c.FileName
}

func (c *Config) GetPackageName() string {
	if c.PackageName == "" {
		panic("Packagename is required")
	}
	return c.PackageName
}

func (c *Config) GetDirection() string {
	if c.Direction == "" {
		panic("Destination is required")
	}
	return c.Direction
}

func (c *Config) GetPackagePath() string {
	if c.PackagePath == "" {
		panic("PackagePath for target is required")
	}
	return c.PackagePath
}

func NewBsonGenerator(c *Config, packages ...string) BSONGenerator {
	return BSONGenerator{
		handledPackages: packages,
		config:          c,
	}
}

type Data struct {
	PackageName string
	PackagePath string
	OptionTypes RegisterCodecTypes
}

func (b *BSONGenerator) Run() error {

	registeredTypes := make(RegisterCodecTypes)

	packs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo}, b.handledPackages...)
	if err != nil {
		return err
	}
	for _, p := range packs {
		scope := p.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if obj == nil {
				continue
			}
			if typ, ok := obj.Type().(*types.Named); ok {
				// struct, interface, etc
				if structType, ok := typ.Underlying().(*types.Struct); ok {
					// structs
					count := structType.NumFields()
					for i := 0; i < count; i++ {
						field := structType.Field(i)
						t := field.Type()
						if strings.Contains(t.String(), "github.com/fasibio/safe.Option") {
							if s, ok := field.Type().Underlying().(*types.Struct); ok {
								f0 := s.Field(0)
								if f0p, ok := f0.Type().Underlying().(*types.Pointer); ok {
									tmp := fillNeededTypes(f0p.Elem(), b.config.GetPackagePath())
									for k, v := range tmp {
										registeredTypes[k] = v
									}

								}
							}
						}
					}
				}

			}
		}
	}

	data := Data{
		PackageName: b.config.GetPackageName(),
		PackagePath: b.config.GetPackagePath(),
		OptionTypes: registeredTypes,
	}

	tmpl, err := template.New("bsonRegisterType").Parse(bsonRegisterType)
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	err = tmpl.Execute(&buff, data)
	if err != nil {
		return err
	}
	formated, err := formatAndFixImports(buff.Bytes())

	if err != nil {
		return err
	}
	filename := filepath.Join(b.config.GetDirection(), b.config.GetFileName())
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.Write(formated)
	if err != nil {
		return err
	}
	return checkSyntax(filename)
}

type RegisterCodecTypes map[string]RegisterCodecType

func (r RegisterCodecTypes) GetPackagePaths(ownPath string) []string {
	l := make(map[string]struct{})
	for _, v := range r {
		if v.PackagePath != "" && v.PackagePath != ownPath {
			l[v.PackagePath] = struct{}{}
		}
	}

	res := make([]string, 0, len(l))
	for k := range l {
		res = append(res, k)
	}
	return res
}

type RegisterCodecType struct {
	PackagePath     string
	PackageName     string
	TypeName        string
	GetCodecSnippet func() string
}

func fillNeededTypes(f0p types.Type, ownPath string) RegisterCodecTypes {
	registeredTypes := make(RegisterCodecTypes)

	switch e := f0p.(type) {
	case *types.Slice:
		// fmt.Println("SLICE", e.Elem().String())
		res := fillNeededTypes(e.Elem(), ownPath)
		for k, v := range res {
			oldGetCodecSnippet := v.GetCodecSnippet
			v.GetCodecSnippet = func() string {
				return fmt.Sprintf("[]%s", oldGetCodecSnippet())
			}
			registeredTypes[k] = v
		}
	case *types.Array:
		{
			// fmt.Println("ARRAY", e.Elem().String())
			res := fillNeededTypes(e.Elem(), ownPath)
			e.Len()
			for k, v := range res {
				oldGetCodecSnippet := v.GetCodecSnippet

				v.GetCodecSnippet = func() string {
					return fmt.Sprintf("[%d]%s", e.Len(), oldGetCodecSnippet())
				}
				registeredTypes[k] = v
			}
		}
	case *types.Basic:
		{
			// fmt.Println("BASIC", e.Name())
			registeredTypes[e.Name()] = RegisterCodecType{
				PackagePath: "",
				PackageName: "",
				TypeName:    e.Name(),
				GetCodecSnippet: func() string {
					return e.Name()
				},
			}
		}
	case *types.Named:
		{
			// fmt.Println("NAMED", e.Obj().Name())
			registeredTypes[fmt.Sprintf("%s.%s", e.Obj().Pkg().Path(), e.Obj().Name())] = RegisterCodecType{
				PackagePath: e.Obj().Pkg().Path(),
				PackageName: e.Obj().Pkg().Name(),
				TypeName:    e.Obj().Name(),
				GetCodecSnippet: func() string {
					if ownPath == e.Obj().Pkg().Path() {
						return e.Obj().Name()
					}
					return fmt.Sprintf("%s.%s", e.Obj().Pkg().Name(), e.Obj().Name())
				},
			}
		}
	default:
		{
			fmt.Println("Not handeled type:", reflect.TypeOf(e), e.String())
		}
	}

	return registeredTypes
}

func formatAndFixImports(code []byte) ([]byte, error) {
	formattedCode, err := imports.Process("", code, &imports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to format and fix imports: %w", err)
	}
	return formatCode(formattedCode)
}

func formatCode(code []byte) ([]byte, error) {
	return format.Source(code)
}

func checkSyntax(filename string) error {
	fset := token.NewFileSet()

	_, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}
	return nil
}
