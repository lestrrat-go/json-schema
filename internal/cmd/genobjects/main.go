package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/lestrrat-go/codegen"
)

func yaml2json(fn string) ([]byte, error) {
	in, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf(`failed to open %q: %w`, fn, err)
	}
	defer in.Close()

	var v interface{}
	if err := yaml.NewDecoder(in).Decode(&v); err != nil {
		return nil, fmt.Errorf(`failed to decode %q: %w`, fn, err)
	}

	return json.Marshal(v)
}

func isNilZeroType(field codegen.Field) bool {
	typ := field.Type()
	return strings.HasPrefix(typ, "*") ||
		strings.HasPrefix(typ, "[]") ||
		strings.HasPrefix(typ, "map[") ||
		field.Name(false) == "schema"
}

func hasAccept(field codegen.Field) bool {
	switch field.Type() {
	case "*Schema":
		return true
	default:
		return false
	}
}

func isInterfaceField(field codegen.Field) bool {
	v, ok := field.Extra(`is_interface`)
	if !ok {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func main() {
	if err := _main(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func _main() error {
	var objectsFile = flag.String("objects", "objects.yml", "")
	flag.Parse()
	jsonSrc, err := yaml2json(*objectsFile)
	if err != nil {
		return err
	}

	var def struct {
		Objects []*codegen.Object `json:"objects"`
	}
	if err := json.NewDecoder(bytes.NewReader(jsonSrc)).Decode(&def); err != nil {
		return fmt.Errorf(`failed to decode %q: %w`, *objectsFile, err)
	}

	for _, object := range def.Objects {
		object.Organize()
		if err := genObject(object); err != nil {
			return fmt.Errorf(`failed to generate object %q: %w`, object.Name(true), err)
		}

		if err := genBuilder(object); err != nil {
			return fmt.Errorf(`failed to generate builder %q: %w`, object.Name(true), err)
		}
	}
	return nil
}

func genObject(obj *codegen.Object) error {
	fn := obj.Name(false) + `_gen.go`

	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("package schema")
	o.LL("type Schema struct {")
	o.L(`isRoot bool`)
	for _, field := range obj.Fields() {
		typ := field.Type()
		if !isNilZeroType(field) && !isInterfaceField(field) {
			typ = "*" + typ
		}
		o.L("%s %s", field.Name(false), typ)
	}
	o.L("}")

	o.LL(`func New() *Schema {`)
	o.L(`return &Schema{`)
	o.L(`schema: Version,`)
	o.L(`}`)
	o.L(`}`)

	for _, field := range obj.Fields() {
		o.LL("func (s *Schema) %s() %s {", field.Name(true), field.Type())
		o.L("return ")
		if !isNilZeroType(field) && !isInterfaceField(field) {
			o.R("*(s.%s)", field.Name(false))
		} else {
			o.R("s.%s", field.Name(false))
		}
		o.L("}")
	}

	o.L(`type pair struct {`)
	o.L(`Name string`)
	o.L(`Value interface{}`)
	o.L(`}`)
	o.LL(`func (s *Schema) MarshalJSON() ([]byte, error) {`)
	o.L(`s.isRoot = true`)
	o.L(`defer func() { s.isRoot = false }()`)
	o.L(`fields := make([]pair, 0, %d)`, len(obj.Fields()))
	for _, field := range obj.Fields() {
		if field.Name(false) == `schema` {
			o.L(`if v := s.%s; s.isRoot && v != "" {`, field.Name(false))
			o.L(`fields = append(fields, pair{Name: %q, Value: v})`, field.JSON())
			o.L(`}`)
		} else {
			o.L(`if v := s.%s; v != nil {`, field.Name(false))
			if !isNilZeroType(field) && !isInterfaceField(field) {
				o.L(`fields = append(fields, pair{Name: %q, Value: *v})`, field.JSON())
			} else {
				o.L(`fields = append(fields, pair{Name: %q, Value: v})`, field.JSON())
			}
			o.L(`}`)
		}
	}
	o.L(`sort.Slice(fields, func(i, j int) bool {`)
	o.L(`return fields[i].Name < fields[j].Name`)
	o.L(`})`)
	o.L(`var buf bytes.Buffer`)
	o.L(`enc := json.NewEncoder(&buf)`)
	o.L(`buf.WriteByte('{')`)
	o.L(`for i, field := range fields {`)
	o.L(`if i > 0 {`)
	o.L(`buf.WriteByte(',')`)
	o.L(`}`)
	o.L(`enc.Encode(field.Name)`)
	o.L(`buf.WriteByte(':')`)
	o.L(`enc.Encode(field.Value)`)
	o.L(`}`)
	o.L(`buf.WriteByte('}')`)
	o.L(`return buf.Bytes(), nil`)
	o.L(`}`)
	o.LL(`func (s *Schema) UnmarshalJSON(buf []byte) error {`)
	o.L("dec := json.NewDecoder(bytes.NewReader(buf))")
	o.L("LOOP:")
	o.L("for {")
	o.L("tok, err := dec.Token()")
	o.L("if err != nil {")
	o.L("return fmt.Errorf(`error reading token: %%w`, err)")
	o.L("}")
	o.L("switch tok := tok.(type) {")
	o.L("case json.Delim:")
	o.L("// Assuming we're doing everything correctly, we should ONLY")
	o.L("// get either '{' or '}' here.")
	o.L("if tok == '}' { // End of object")
	o.L("break LOOP")
	o.L("} else if tok != '{' {")
	o.L("return fmt.Errorf(`expected '{', but got '%%c'`, tok)")
	o.L("}")
	o.L("case string: // Objects can only have string keys")
	o.L("switch tok {")
	for _, field := range obj.Fields() {
		_ = field
		switch field.Type() {
		default:
			o.L("case %q:", field.JSON())
			if field.Name(false) == "additionalProperties" {
				// Attempt to decode as *Schema first, then as a boolean
				o.L("var v *Schema")
				o.L("var tmp Schema")
				o.L("if err := dec.Decode(&tmp); err == nil {")
				o.L("v = &tmp")
				o.L("} else {")
				o.L("var b bool")
				o.L("if err = dec.Decode(&b); err != nil {")
				o.L("return fmt.Errorf(`failed to decode value for field %q: %%w`, err)", field.JSON())
				o.L("}")
				o.L("if b {")
				o.L("v = &Schema{}")
				o.L("}")
				o.L("}")
				o.L("s.%s = v", field.Name(false))
			} else {
				o.L("var v %s", field.Type())
				o.L("if err := dec.Decode(&v); err != nil {")
				o.L("return fmt.Errorf(`failed to decode value for field %q: %%w`, err)", field.JSON())
				o.L("}")
				if !isNilZeroType(field) {
					o.L("s.%s = &v", field.Name(false))
				} else {
					o.L("s.%s = v", field.Name(false))
				}
			}
		}
	}
	o.L("}")
	o.L("}")
	o.L("}")
	o.L("return nil")
	o.L(`}`)

	if err := o.WriteFile(fn, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return err
	}
	return nil
}

func genBuilder(obj *codegen.Object) error {
	fn := `builder_gen.go`

	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("package schema")

	o.LL("type Builder struct {")
	o.L("err error")
	for _, field := range obj.Fields() {
		switch field.Type() {
		case `map[string]*Schema`:
			o.L("%s []*propPair", field.Name(false))
		default:
			if !isNilZeroType(field) && !isInterfaceField(field) {
				o.L("%s *%s", field.Name(false), field.Type())
			} else {
				o.L("%s %s", field.Name(false), field.Type())
			}
		}
	}
	o.L("}")

	o.LL("func NewBuilder() *Builder {")
	o.L(`return &Builder{`)
	o.L(`schema: Version,`)
	o.L(`}`)
	o.L("}")

	for _, field := range obj.Fields() {
		if field.Name(true) == "AdditionalProperties" {
			o.LL("func (b *Builder) AdditionalProperties(v SchemaOrBool) *Builder {")
			o.L("if b.err != nil {")
			o.L("return b")
			o.L("}")
			o.L("var tmp Schema")
			o.L("if err := tmp.Accept(v); err != nil {")
			o.L("b.err = fmt.Errorf(`failed to accept value for %q: %%w`, err)", field.JSON())
			o.L("return b")
			o.L("}")
			o.L("b.additionalProperties = &tmp")
			o.L("return b")
			o.L("}")
			continue
		}
		switch field.Type() {
		case `[]PrimitiveType`:
			o.LL("func (b *Builder) Type(v PrimitiveType) *Builder {")
			o.L("if b.err != nil {")
			o.L("return b")
			o.L("}")
			o.L("b.types = append(b.types, v)")
			o.L("return b")
			o.L("}")
		case `map[string]*Schema`:
			name := strings.Replace(field.Name(true), `Properties`, `Property`, 1)
			if name == "PropertyNames" {
				name = "PropertyName"
			}
			o.LL("func (b *Builder) %s(n string, v %s) *Builder {", name, `*Schema`)
			o.L("if b.err != nil {")
			o.L("return b")
			o.L("}")

			o.LL(`b.%[1]s = append(b.%[1]s, &propPair{Name: n, Schema: v})`, field.Name(false))
			o.L("return b")
			o.L("}")
		default:
			o.LL("func (b *Builder) %s(v %s) *Builder {", field.Name(true), field.Type())
			o.L("if b.err != nil {")
			o.L("return b")
			o.L("}")

			if !isNilZeroType(field) && !isInterfaceField(field) {
				o.LL("b.%s = &v", field.Name(false))
			} else {
				o.LL("b.%s = v", field.Name(false))
			}
			o.L("return b")
			o.L("}")
		}
	}

	o.LL("func (b *Builder) Build() (*Schema, error) {")
	o.L("s := New()")
	for _, field := range obj.Fields() {
		switch field.Type() {
		case `map[string]*Schema`:
			o.LL(`if b.%s != nil {`, field.Name(false))
			o.L("s.%s = make(map[string]*Schema)", field.Name(false))
			o.L("for _, pair := range b.%s {", field.Name(false))
			o.L("if _, ok := s.%s[pair.Name]; ok {", field.Name(false))
			o.L("return nil, fmt.Errorf(`duplicate key %%q in %q`, pair.Name)", field.JSON())
			o.L("}")
			o.L("s.%s[pair.Name] = pair.Schema", field.Name(false))
			o.L("}")
			o.L("}")
		default:
			if field.Name(false) == `schema` {
				o.L("s.%[1]s = b.%[1]s", field.Name(false))
			} else {
				o.L(`if b.%s != nil {`, field.Name(false))
				o.L("s.%[1]s = b.%[1]s", field.Name(false))
				o.L(`}`)
			}
		}
	}
	o.L("return s, nil")
	o.L("}")

	if err := o.WriteFile(fn, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return err
	}
	return nil
}
