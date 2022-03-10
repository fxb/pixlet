package main

// Generates starlark bindings for the pixlet/render package.
//
// Also produces widget documentation and extracts example snippets
// that can be run with docs/gen.go to produce images for the widget
// docs.

import (
	"bytes"
	"fmt"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"image/color"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"tidbyt.dev/pixlet/render"
	"tidbyt.dev/pixlet/render/animation"
)

const StarlarkHeaderTemplate = "./gen/header.tmpl"
const StarlarkWidgetTemplate = "./gen/widget.tmpl"
const DocumentationTemplate = "./gen/docs.tmpl"
const CodeOut = "./generated.go"
const DocumentationOut = "../docs/widgets.md"
const DocumentationDirectory = "../docs/"

type RenderPackage struct {
	Name       string
	Directory  string
	ImportPath string
}

var RenderPackages = []RenderPackage{
	{"render", "../render", "tidbyt.dev/pixlet/render"},
	{"animation", "../render/animation", "tidbyt.dev/pixlet/render/animation"},
}

var RenderWidgets = []interface{}{
	&render.Text{},
	&render.Image{},
	render.Row{},
	render.Column{},
	render.Stack{},
	render.Padding{},
	render.Box{},
	render.Circle{},
	render.Marquee{},
	render.Animation{},
	render.WrappedText{},

	&animation.Animate{},
	animation.Keyframe{},
	animation.Origin{},
	animation.Translate{},
	animation.Scale{},
	animation.Rotate{},

	animation.Bounce{},
	animation.AnimatedPositioned{},
}

type AttributeDefinition struct {
	Type          string
	StarlarkType  string
	TemplatePath  string
	StarlarkField bool
}

var AttributeDefinitions = map[interface{}]AttributeDefinition{
	// Primitive types
	new(string): {
		Type:         "starlark.String",
		StarlarkType: "str",
		TemplatePath: "./gen/attr/string.tmpl",
	},
	new(int): {
		Type:         "starlark.Int",
		StarlarkType: "int",
		TemplatePath: "./gen/attr/int.tmpl",
	},
	new(float64): {
		Type:         "starlark.Value",
		StarlarkType: "float / int",
		TemplatePath: "./gen/attr/float.tmpl",
	},
	new(bool): {
		Type:         "starlark.Bool",
		StarlarkType: "bool",
		TemplatePath: "./gen/attr/bool.tmpl",
	},

	// Render types
	new(render.Insets): {
		Type:         "starlark.Value",
		StarlarkType: "int / (int, int, int, int)",
		TemplatePath: "./gen/attr/insets.tmpl",
	},
	new(render.Widget): {
		Type:         "starlark.Value",
		StarlarkType: "Widget",
		TemplatePath: "./gen/attr/child.tmpl",
	},
	new([]render.Widget): {
		Type:         "*starlark.List",
		StarlarkType: "[Widget]",
		TemplatePath: "./gen/attr/children.tmpl",
	},
	new(color.Color): {
		Type:          "starlark.String",
		StarlarkType:  `str`,
		TemplatePath:  "./gen/attr/color.tmpl",
		StarlarkField: true,
	},

	// Animation types
	new(animation.Origin): {
		Type:         "starlark.Value",
		StarlarkType: "Origin",
		TemplatePath: "./gen/attr/origin.tmpl",
	},
	new(animation.Curve): {
		Type:         "starlark.Value",
		StarlarkType: `str / function`,
		TemplatePath: "./gen/attr/curve.tmpl",
	},
	new(animation.Rounding): {
		Type:          "starlark.String",
		StarlarkType:  `str`,
		TemplatePath:  "./gen/attr/rounding.tmpl",
		StarlarkField: true,
	},
	new(animation.NumberOrPercentage): {
		Type:         "starlark.Value",
		StarlarkType: `float / int / str`,
		TemplatePath: "./gen/attr/num_pct.tmpl",
	},
	new([]animation.Keyframe): {
		Type:         "*starlark.List",
		StarlarkType: "[Keyframe]",
		TemplatePath: "./gen/attr/keyframes.tmpl",
	},
	new([]animation.Transform): {
		Type:         "*starlark.List",
		StarlarkType: "[Transform]",
		TemplatePath: "./gen/attr/transforms.tmpl",
	},
}

// Defines the starlark version of a render.Widget
type Attribute struct {
	Render        string
	Type          string
	Starlark      string
	StarlarkType  string
	StarlarkField bool
	Required      bool
	ReadOnly      bool
	Template      *template.Template
	Code          string
	Documentation string
}

type StarlarkWidget struct {
	Name          string
	FullName      string
	SuperName     string
	Attr          []*Attribute
	HasSize       bool
	HasPtrRcvr    bool
	RequiresInit  bool
	Documentation string
	Examples      []string
}

type StarlarkHeader struct {
	Widget []StarlarkWidget
}

func nilOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func decay(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		return t.Elem()
	}

	return t
}

func starlarkWidgetFromRenderWidget(w interface{}, docOnly bool) *StarlarkWidget {
	sw := StarlarkWidget{}

	val := reflect.ValueOf(w)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
		sw.HasPtrRcvr = true
	}
	if val.Kind() != reflect.Struct {
		panic("widget is neither struct nor pointer to struct, wtf?")
	}

	typ := val.Type()

	sw.Name = typ.Name()
	sw.FullName = typ.String()

	if _, isWidget := w.(render.Widget); isWidget {
		sw.SuperName = "Widget"
	}

	if _, hasSize := w.(render.WidgetStaticSize); hasSize {
		sw.HasSize = true
	}

	if _, requiresInit := w.(render.WidgetWithInit); requiresInit {
		sw.RequiresInit = true
	}

FieldLoop:
	for _, field := range deepFields(val) {

		if field.PkgPath != "" || field.Anonymous {
			// Field is not an exposed attribute
			continue
		}

		// Widget fields can be tagged `starlark:"<name>[<param>...]"` to
		// control attribute name in starlark.
		//
		// Additional supported flags:
		// "required" - field is required on instantiation
		// "readonly" - field is read-only, and not passed to constructor
		attr := &Attribute{
			Render: field.Name,
		}
		fieldTag, ok := field.Tag.Lookup("starlark")
		if ok {
			tag := strings.Split(fieldTag, ",")
			attr.Starlark = strings.TrimSpace(tag[0])
			for _, t := range tag[1:] {
				t = strings.TrimSpace(t)
				if t == "required" {
					attr.Required = true
				} else if t == "readonly" {
					attr.ReadOnly = true
				} else {
					panic(fmt.Sprintf(
						"%s.%s has unsupported tag: '%s'",
						typ.Name(), field.Name, tag[1],
					))
				}
			}
		}
		if attr.Starlark == "" {
			attr.Starlark = strings.ToLower(field.Name)
		}

		sw.Attr = append(sw.Attr, attr)

		for val, def := range AttributeDefinitions {
			if field.Type == decay(reflect.TypeOf(val)) {
				attr.Type = def.Type
				attr.StarlarkType = def.StarlarkType
				attr.StarlarkField = def.StarlarkField
				attr.Template = loadTemplate("attr", def.TemplatePath)
				continue FieldLoop
			}
		}

		panic(fmt.Sprintf(
			"%s.%s has unsupported type",
			typ.Name(), field.Name,
		))
	}

	// Reorder AttrAll so that required fields appear first
	sort.SliceStable(sw.Attr, func(i, j int) bool {
		return sw.Attr[i].Required && !sw.Attr[j].Required
	})

	return &sw
}

func deepFields(val reflect.Value) []reflect.StructField {
	fields := make([]reflect.StructField, 0)
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		t := typ.Field(i)
		v := val.Field(i)

		if t.Anonymous && t.Type.Kind() == reflect.Struct {
			fields = append(fields, deepFields(v)...)
		} else {
			fields = append(fields, t)
		}
	}

	return fields
}

func attachWidgetDocs(widgets []*StarlarkWidget) {

	// Parse all .go files in pixlet/render packages and extract all type doc comments
	fset := token.NewFileSet()
	docs := map[string]string{}
	for _, p := range RenderPackages {
		astPkgs, err := parser.ParseDir(fset, p.Directory, nil, parser.ParseComments)
		nilOrPanic(err)
		pkg := doc.New(astPkgs[p.Name], p.ImportPath, 0)
		nilOrPanic(err)
		for _, type_ := range pkg.Types {
			docs[type_.Name] = type_.Doc
		}
	}

	// These match our attribute docs and example blocks
	docRe, err := regexp.Compile(`(?m)^DOC\(([^)]+)\): +(.+)$`)
	nilOrPanic(err)
	exampleRe, err := regexp.Compile(`(?s)EXAMPLE BEGIN(.*?)EXAMPLE END`)
	nilOrPanic(err)

	for _, widget := range widgets {
		// Widget doc is full comment sans attribute docs and examples
		widget.Documentation = strings.TrimSpace(string(
			docRe.ReplaceAllString(
				exampleRe.ReplaceAllString(docs[widget.Name], ""),
				"",
			),
		))

		// Attribute docs
		attrDocs := map[string]string{}
		for _, group := range docRe.FindAllStringSubmatch(docs[widget.Name], -1) {
			attrDocs[group[1]] = group[2]
		}
		for _, attr := range widget.Attr {
			attr.Documentation = attrDocs[attr.Render]
		}

		// Examples
		examples := []string{}
		for _, group := range exampleRe.FindAllStringSubmatch(docs[widget.Name], -1) {
			examples = append(examples, strings.TrimSpace(group[1]))
		}
		widget.Examples = examples

	}
}

func loadTemplate(name, path string) *template.Template {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	content, err := ioutil.ReadFile(path)
	nilOrPanic(err)

	template, err := template.New(name).Funcs(funcMap).Parse(string(content))
	nilOrPanic(err)

	return template
}

func generateCode(widgets []*StarlarkWidget) {

	headerTemplate := loadTemplate("header", StarlarkHeaderTemplate)
	widgetTemplate := loadTemplate("widget", StarlarkWidgetTemplate)

	// Execute attribute templates.
	for _, widget := range widgets {
		for _, attr := range widget.Attr {
			var buf bytes.Buffer
			err := attr.Template.Execute(&buf, attr)
			nilOrPanic(err)
			attr.Code = string(buf.Bytes())
		}
	}

	outf, err := os.Create(CodeOut)
	nilOrPanic(err)
	defer outf.Close()

	var buf bytes.Buffer
	err = headerTemplate.Execute(&buf, widgets)
	nilOrPanic(err)

	for _, data := range widgets {
		err = widgetTemplate.Execute(&buf, data)
		nilOrPanic(err)
	}

	formatted, err := format.Source(buf.Bytes())
	nilOrPanic(err)
	outf.Write(formatted)
}

func generateDocs(widgets []*StarlarkWidget) {
	docsTemplateContent, err := ioutil.ReadFile(DocumentationTemplate)
	nilOrPanic(err)

	docsTemplate, err := template.New("docs").Parse(string(docsTemplateContent))
	nilOrPanic(err)

	outf, err := os.Create(DocumentationOut)
	nilOrPanic(err)
	defer outf.Close()

	err = docsTemplate.Execute(outf, widgets)
	nilOrPanic(err)

	for _, widget := range widgets {
		for i, example := range widget.Examples {
			err = ioutil.WriteFile(
				fmt.Sprintf("%s/%s_%d.star", DocumentationDirectory, widget.Name, i),
				[]byte(example),
				0644)
			nilOrPanic(err)
		}
	}
}

func main() {
	widgets := []*StarlarkWidget{}

	for _, w := range RenderWidgets {
		widgets = append(widgets, starlarkWidgetFromRenderWidget(w, false))
	}

	sort.SliceStable(widgets, func(i, j int) bool {
		return widgets[i].Name < widgets[j].Name
	})

	attachWidgetDocs(widgets)
	generateCode(widgets)
	generateDocs(widgets)
}
