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
		typ == "PrimitiveTypes" ||
		typ == "SchemaOrBool" ||
		field.Name(false) == "schema"
}

func isVariadicSliceType(field codegen.Field) bool {
	typ := field.Type()
	switch typ {
	case "[]*Schema", "[]interface{}", "[]string", "PrimitiveTypes", "[]SchemaOrBool":
		return true
	default:
		return false
	}
}

func getVariadicElementType(field codegen.Field) string {
	typ := field.Type()
	switch typ {
	case "[]*Schema":
		return "*Schema"
	case "[]interface{}":
		return "interface{}"
	case "[]string":
		return "string"
	case "PrimitiveTypes":
		return "PrimitiveType"
	case "[]SchemaOrBool":
		return "SchemaOrBool"
	default:
		return ""
	}
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
		if field.Name(false) != `schema` {
			o.LL("func (s *Schema) Has%s() bool {", field.Name(true))
			o.L("return s.%s != nil", field.Name(false))
			o.L("}")
		}
		o.LL("func (s *Schema) %s() %s {", field.Name(true), field.Type())
		o.L("return ")
		if !isNilZeroType(field) && !isInterfaceField(field) {
			o.R("*(s.%s)", field.Name(false))
		} else {
			o.R("s.%s", field.Name(false))
		}
		o.L("}")
	}

	o.LL("func (s *Schema) ContainsType(typ PrimitiveType) bool {")
	o.L("if s.types == nil {")
	o.L("return false")
	o.L("}")
	o.L("for _, t := range s.types {")
	o.L("if t == typ {")
	o.L("return true")
	o.L("}")
	o.L("}")
	o.L("return false")
	o.L("}")

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
	o.L(`return compareFieldNames(fields[i].Name, fields[j].Name)`)
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
	o.L("return fmt.Errorf(`json-schema: failed to read JSON token: %%w`, err)")
	o.L("}")
	o.L("switch tok := tok.(type) {")
	o.L("case json.Delim:")
	o.L("// Assuming we're doing everything correctly, we should ONLY")
	o.L("// get either '{' or '}' here.")
	o.L("if tok == '}' { // End of object")
	o.L("break LOOP")
	o.L("} else if tok != '{' {")
	o.L("return fmt.Errorf(`json-schema: failed to parse JSON structure: expected '{', but got '%%c'`, tok)")
	o.L("}")
	o.L("case string: // Objects can only have string keys")
	o.L("switch tok {")
	for _, field := range obj.Fields() {
		_ = field
		switch field.Type() {
		default:
			o.L("case %q:", field.JSON())
			if field.Type() == "SchemaOrBool" {
				// Handle single SchemaOrBool fields
				o.L("var rawData json.RawMessage")
				o.L("if err := dec.Decode(&rawData); err != nil {")
				o.L("return fmt.Errorf(`json-schema: failed to decode raw data for field %q: %%w`, err)", field.JSON())
				o.L("}")
				o.L("// Try to decode as boolean first")
				o.L("var b bool")
				o.L("if err := json.Unmarshal(rawData, &b); err == nil {")
				o.L("s.%s = SchemaBool(b)", field.Name(false))
				o.L("} else {")
				o.L("// Try to decode as Schema object")
				o.L("var schema Schema")
				o.L("if err := json.Unmarshal(rawData, &schema); err == nil {")
				o.L("s.%s = &schema", field.Name(false))
				o.L("} else {")
				o.L("return fmt.Errorf(`json-schema: failed to decode value for field %q (attempting to unmarshal as Schema after bool failed): %%w`, err)", field.JSON())
				o.L("}")
				o.L("}")
			} else if field.Type() == "[]SchemaOrBool" {
				// Special handling for []SchemaOrBool fields - use token-based parsing
				o.L("v, err := unmarshalSchemaOrBoolSlice(dec)")
				o.L("if err != nil {")
				o.L("return fmt.Errorf(`json-schema: failed to decode value for field %q (attempting to unmarshal as []SchemaOrBool slice): %%w`, err)", field.JSON())
				o.L("}")
				o.L("s.%s = v", field.Name(false))
			} else if field.Type() == "SchemaOrBool" {
				// Special handling for SchemaOrBool fields - decode as raw JSON values
				o.L("var v %s", field.Type())
				o.L("if err := dec.Decode(&v); err != nil {")
				o.L("return fmt.Errorf(`json-schema: failed to decode value for field %q (attempting to unmarshal as %s): %%w`, err)", field.JSON(), field.Type())
				o.L("}")
				o.L("s.%s = v", field.Name(false))
			} else if field.Type() == "*Schema" {
				// Special handling for *Schema fields - they can be objects or booleans
				o.L("var rawData json.RawMessage")
				o.L("if err := dec.Decode(&rawData); err != nil {")
				o.L("return fmt.Errorf(`json-schema: failed to decode raw data for field %q: %%w`, err)", field.JSON())
				o.L("}")
				o.L("// Try to decode as boolean first")
				o.L("var b bool")
				o.L("if err := json.Unmarshal(rawData, &b); err == nil {")
				o.L("// Convert boolean to Schema object")
				o.L("if b {")
				o.L("s.%s = &Schema{} // true schema - allow everything", field.Name(false))
				o.L("} else {")
				o.L("s.%s = &Schema{not: &Schema{}} // false schema - deny everything", field.Name(false))
				o.L("}")
				o.L("} else {")
				o.L("// Try to decode as Schema object")
				o.L("var schema Schema")
				o.L("if err := json.Unmarshal(rawData, &schema); err == nil {")
				o.L("s.%s = &schema", field.Name(false))
				o.L("} else {")
				o.L("return fmt.Errorf(`json-schema: failed to decode value for field %q (attempting to unmarshal as Schema after bool failed): %%w`, err)", field.JSON())
				o.L("}")
				o.L("}")
			} else if field.Type() == "map[string]*Schema" {
				// Special handling for map[string]*Schema fields - values can be objects or booleans
				o.L("var rawData json.RawMessage")
				o.L("if err := dec.Decode(&rawData); err != nil {")
				o.L("return fmt.Errorf(`json-schema: failed to decode raw data for field %q: %%w`, err)", field.JSON())
				o.L("}")
				o.L("// First unmarshal as map[string]json.RawMessage")
				o.L("var rawMap map[string]json.RawMessage")
				o.L("if err := json.Unmarshal(rawData, &rawMap); err != nil {")
				o.L("return fmt.Errorf(`json-schema: failed to decode value for field %q (attempting to unmarshal as map): %%w`, err)", field.JSON())
				o.L("}")
				o.L("// Convert each value to *Schema")
				o.L("v := make(map[string]*Schema)")
				o.L("for key, rawValue := range rawMap {")
				o.L("// Try to decode as boolean first")
				o.L("var b bool")
				o.L("if err := json.Unmarshal(rawValue, &b); err == nil {")
				o.L("// Convert boolean to Schema object")
				o.L("if b {")
				o.L("v[key] = &Schema{} // true schema - allow everything")
				o.L("} else {")
				o.L("v[key] = &Schema{not: &Schema{}} // false schema - deny everything")
				o.L("}")
				o.L("} else {")
				o.L("// Try to decode as Schema object")
				o.L("var schema Schema")
				o.L("if err := json.Unmarshal(rawValue, &schema); err == nil {")
				o.L("v[key] = &schema")
				o.L("} else {")
				o.L("return fmt.Errorf(`json-schema: failed to decode value for field %q key %%q (attempting to unmarshal as Schema after bool failed): %%w`, key, err)", field.JSON())
				o.L("}")
				o.L("}")
				o.L("}")
				o.L("s.%s = v", field.Name(false))
			} else {
				o.L("var v %s", field.Type())
				o.L("if err := dec.Decode(&v); err != nil {")
				o.L("return fmt.Errorf(`json-schema: failed to decode value for field %q (attempting to unmarshal as %s): %%w`, err)", field.JSON(), field.Type())
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
		if field.Type() == "SchemaOrBool" {
			o.LL("func (b *Builder) %s(v SchemaOrBool) *Builder {", field.Name(true))
			o.L("if b.err != nil {")
			o.L("return b")
			o.L("}")
			o.L("b.%s = v", field.Name(false))
			o.L("return b")
			o.L("}")
			continue
		}
		switch field.Type() {
		case `map[string]*Schema`:
			name := strings.Replace(field.Name(true), `Properties`, `Property`, 1)
			o.LL("func (b *Builder) %s(n string, v %s) *Builder {", name, `*Schema`)
			o.L("if b.err != nil {")
			o.L("return b")
			o.L("}")

			o.LL(`b.%[1]s = append(b.%[1]s, &propPair{Name: n, Schema: v})`, field.Name(false))
			o.L("return b")
			o.L("}")
		default:
			if isVariadicSliceType(field) {
				elementType := getVariadicElementType(field)
				o.LL("func (b *Builder) %s(v ...%s) *Builder {", field.Name(true), elementType)
				o.L("if b.err != nil {")
				o.L("return b")
				o.L("}")
				if field.Type() == "PrimitiveTypes" {
					o.LL("b.%s = PrimitiveTypes(v)", field.Name(false))
				} else if field.Type() == "[]SchemaOrBool" {
					o.LL("for _, item := range v {")
					o.L("if err := validateSchemaOrBool(item); err != nil {")
					o.L("b.err = fmt.Errorf(`invalid value in %s: %%w`, err)", field.Name(true))
					o.L("return b")
					o.L("}")
					o.L("}")
					o.LL("b.%s = v", field.Name(false))
				} else {
					o.LL("b.%s = v", field.Name(false))
				}
				o.L("return b")
				o.L("}")
			} else {
				paramType := field.Type()
				o.LL("func (b *Builder) %s(v %s) *Builder {", field.Name(true), paramType)
				o.L("if b.err != nil {")
				o.L("return b")
				o.L("}")

				if field.Type() == "SchemaOrBool" {
					o.LL("if err := validateSchemaOrBool(v); err != nil {")
					o.L("b.err = fmt.Errorf(`invalid value for %s: %%w`, err)", field.Name(true))
					o.L("return b")
					o.L("}")
				}

				if !isNilZeroType(field) && !isInterfaceField(field) {
					o.LL("b.%s = &v", field.Name(false))
				} else {
					o.LL("b.%s = v", field.Name(false))
				}
				o.L("return b")
				o.L("}")
			}
		}
	}

	// Clone method creates a new Builder pre-initialized with values from an existing Schema
	o.LL("func (b *Builder) Clone(original *Schema) *Builder {")
	o.L("if b.err != nil {")
	o.L("return b")
	o.L("}")
	o.L("if original == nil {")
	o.L("return b")
	o.L("}")
	
	// Copy all fields from original schema to builder
	for _, field := range obj.Fields() {
		switch field.Type() {
		case `map[string]*Schema`:
			// For map fields, we need to copy the map to propPair slices
			o.LL("if original.%s != nil {", field.Name(false))
			o.L("for name, schema := range original.%s {", field.Name(false))
			o.L("b.%s = append(b.%s, &propPair{Name: name, Schema: schema})", field.Name(false), field.Name(false))
			o.L("}")
			o.L("}")
		default:
			if field.Name(false) == "schema" {
				// Special handling for schema field (it's a string, not a pointer)
				o.LL("b.%s = original.%s", field.Name(false), field.Name(false))
			} else {
				o.LL("if original.%s != nil {", field.Name(false))
				o.L("b.%s = original.%s", field.Name(false), field.Name(false))
				o.L("}")
			}
		}
	}
	
	o.L("return b")
	o.L("}")

	// Reset methods for clearing individual fields
	for _, field := range obj.Fields() {
		methodName := "Reset" + field.Name(true)
		o.LL("func (b *Builder) %s() *Builder {", methodName)
		o.L("if b.err != nil {")
		o.L("return b")
		o.L("}")
		
		switch field.Type() {
		case `map[string]*Schema`:
			// For map fields, clear the propPair slice
			o.L("b.%s = nil", field.Name(false))
		default:
			if field.Name(false) == "schema" {
				// Special handling for schema field (it's a string, reset to default)
				o.L("b.%s = Version", field.Name(false))
			} else {
				// For all other fields, set to nil/zero value
				o.L("b.%s = nil", field.Name(false))
			}
		}
		
		o.L("return b")
		o.L("}")
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

	o.LL("func (b *Builder) MustBuild() *Schema {")
	o.L("s, err := b.Build()")
	o.L("if err != nil {")
	o.L("panic(fmt.Errorf(`failed to build schema: %%w`, err))")
	o.L("}")
	o.L("return s")
	o.L("}")

	if err := o.WriteFile(fn, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return err
	}
	return nil
}
