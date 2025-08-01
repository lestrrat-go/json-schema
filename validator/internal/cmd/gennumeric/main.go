package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"unicode"

	"github.com/lestrrat-go/codegen"
	"github.com/lestrrat-go/xstrings"
)

type definition struct {
	typ      string
	class    string
	filename string
}

// Generate type NumberValidator and type IntegerValidator
func main() {
	var outputDir = flag.String("output", ".", "output directory for generated files")
	flag.Parse()

	if err := _main(*outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func _main(outputDir string) error {
	defs := []definition{
		{
			typ:      "int",
			class:    "Integer",
			filename: "int_gen.go",
		},
		{
			typ:      "float64",
			class:    "Number",
			filename: "number_gen.go",
		},
	}

	for _, def := range defs {
		if err := generateValidator(def, outputDir); err != nil {
			return fmt.Errorf(`failed to generate validator for %s: %w`, def.typ, err)
		}
	}
	return nil
}

func generateValidator(def definition, outputDir string) error {
	props := []string{
		"multipleOf",
		"maximum",
		"exclusiveMaximum",
		"minimum",
		"exclusiveMinimum",
		"constantValue",
		"enum",
	}

	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("package validator")
	o.L("")
	o.L("import (")
	o.L("\t\"context\"")
	o.L("\t\"fmt\"")
	o.L("\t\"math\"")
	o.L("\t\"reflect\"")
	o.L("")
	o.L("\tschema \"github.com/lestrrat-go/json-schema\"")
	o.L("\t\"github.com/lestrrat-go/json-schema/vocabulary\"")
	o.L(")")
	o.L("")
	o.L("var _ Builder = (*%sValidatorBuilder)(nil)", def.class)
	o.L("var _ Interface = (*%sValidator)(nil)", xstrings.Snake(def.class))

	o.LL("func compile%sValidator(ctx context.Context, s *schema.Schema) (Interface, error) {", def.class)
	o.L("b := %s()", def.class)
	for _, prop := range props {
		var methodName string
		if prop == "constantValue" {
			methodName = "Const"
		} else {
			methodName = xstrings.Camel(prop)
		}

		if prop == "enum" {
			o.LL("if s.HasEnum() && vocabulary.IsKeywordEnabledInContext(ctx, \"enum\") {")
			o.L("enums := s.Enum()")
			o.L("l := make([]%s, 0, len(enums))", def.typ)
			o.L("for i, e := range s.Enum() {")
			o.L("rv := reflect.ValueOf(e)")
			o.L("var tmp %s", def.typ)
			o.L("switch rv.Kind() {")
			o.L("case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:")
			if def.typ == "int" {
				o.L("tmp = int(rv.Int())")
			} else {
				o.L("tmp = float64(rv.Int())")
			}
			o.L("case reflect.Float32, reflect.Float64:")
			if def.typ == "int" {
				o.L("tmp = int(rv.Float())")
			} else {
				o.L("tmp = rv.Float()")
			}
			o.L("default:")
			o.L("return nil, fmt.Errorf(`invalid element in enum: expected numeric element, got %%T for element %%d`, e, i)")
			o.L("}") // switch
			o.L("l = append(l, tmp)")
			o.L("}") // for
			o.L("b.Enum(l...)")
			o.L("}") // if s.HasEnum
		} else {
			runes := []rune(methodName)
			first := runes[0]
			lower := string(append(append([]rune(nil), unicode.ToLower(first)), runes[1:]...))
			o.LL("if s.Has%s() && vocabulary.IsKeywordEnabledInContext(ctx, %q) {", methodName, lower)
			o.L("rv := reflect.ValueOf(s.%s())", methodName)
			o.L("var tmp %s", def.typ)
			o.L("switch rv.Kind() {")
			o.L("case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:")
			if def.typ == "int" {
				o.L("tmp = int(rv.Int())")
			} else {
				o.L("tmp = float64(rv.Int())")
			}
			if methodName == "MultipleOf" {
				o.L("b.%s(tmp)", methodName)
			}
			o.L("case reflect.Float32, reflect.Float64:")
			if def.typ == "int" {
				if prop == "multipleOf" {
					o.L("f := rv.Float()")
					o.L("// Skip multipleOf constraint for very small values with integer type")
					o.L("// Any integer is a multiple of very small numbers like 1e-8")
					o.L("if f <= 0 || f >= 1 {")
					o.L("tmp = int(f)")
					o.L("b.%s(tmp)", methodName)
					o.L("}")
				} else {
					o.L("tmp = int(rv.Float())")
				}
			} else {
				o.L("tmp = rv.Float()")
			}
			if def.typ != "int" || prop != "multipleOf" {
				o.L("default:")
				o.L("return nil, fmt.Errorf(`invalid type for %s field: expected numeric type, got %%T`, rv.Interface())", prop)
				o.L("}") // switch
				if def.typ != "int" || prop != "multipleOf" {
					o.L("b.%s(tmp)", methodName)
				}
			} else {
				o.L("default:")
				o.L("return nil, fmt.Errorf(`invalid type for %s field: expected numeric type, got %%T`, rv.Interface())", prop)
				o.L("}") // switch
			}
			o.L("}") // if s.Has
		}
	}
	o.L("return b.Build()")
	o.L("}")

	o.LL("type %sValidator struct {", xstrings.Snake(def.class))
	for _, prop := range props {
		if prop == "enum" {
			o.L("%s []%s", prop, def.typ)
		} else {
			o.L("%s *%s", prop, def.typ)
		}
	}
	o.L("}")

	o.LL("type %sValidatorBuilder struct {", def.class)
	o.L("err error")
	o.L("c *%sValidator", xstrings.Snake(def.class))
	o.L("}")

	o.LL("func %[1]s() *%[1]sValidatorBuilder {", def.class)
	o.L("return (&%[1]sValidatorBuilder{}).Reset()", def.class)
	o.L("}")

	for _, prop := range props {
		var methodName string
		if prop == "constantValue" {
			methodName = "Const"
		} else {
			methodName = xstrings.Camel(prop)
		}

		if prop == "enum" {
			o.LL("func (b *%[1]sValidatorBuilder) %[2]s(v ...%[3]s) *%[1]sValidatorBuilder {", def.class, methodName, def.typ)
			o.L("if b.err != nil {")
			o.L("return b")
			o.L("}")
			o.L("b.c.%s = make([]%s, len(v))", prop, def.typ)
			o.L("copy(b.c.%s, v)", prop)
			o.L("return b")
			o.L("}")
		} else {
			o.LL("func (b *%[1]sValidatorBuilder) %[2]s(v %[3]s) *%[1]sValidatorBuilder {", def.class, methodName, def.typ)
			o.L("if b.err != nil {")
			o.L("return b")
			o.L("}")
			o.L("b.c.%s = &v", prop)
			o.L("return b")
			o.L("}")
		}
	}

	o.LL("func (b *%[1]sValidatorBuilder) Build() (Interface, error) {", def.class)
	o.L("if b.err != nil {")
	o.L("return nil, b.err")
	o.L("}")
	o.L("return b.c, nil")
	o.L("}")
	o.L("")
	o.L("func (b *%[1]sValidatorBuilder) MustBuild() Interface {", def.class)
	o.L("if b.err != nil {")
	o.L("panic(b.err)")
	o.L("}")
	o.L("return b.c")
	o.L("}")
	o.L("")
	o.L("func (b *%[1]sValidatorBuilder) Reset() *%[1]sValidatorBuilder {", def.class)
	o.L("b.err = nil")
	o.L("b.c = &%sValidator{}", xstrings.Snake(def.class))
	o.L("return b")
	o.L("}")

	var template string
	if def.typ == "int" {
		template = "d"
	} else {
		template = "f"
	}
	o.LL("func (v *%sValidator) Validate(_ context.Context, in any) (Result, error) {", xstrings.Snake(def.class))
	o.L("rv := reflect.ValueOf(in)")
	if def.typ == "int" {
		o.LL("var n int")
		o.L("switch rv.Kind() {")
		o.L("case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:")
		o.L("n = int(rv.Int())")
		o.L("case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:")
		o.L("n = int(rv.Uint())")
		o.L("case reflect.Float32, reflect.Float64:")
		o.L("f := rv.Float()")
		o.L("if f != float64(int(f)) {")
		o.L("return nil, fmt.Errorf(`expected integer, got float value %%g`, f)")
		o.L("}")
		o.L("n = int(f)")
		o.L("default:")
		o.L("return nil, fmt.Errorf(`invalid value passed to IntegerValidator: expected integer, got %%T`, in)")
		o.L("}")
	} else {
		o.LL("var n float64")
		o.L("switch rv.Kind() {")
		o.L("case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:")
		o.L("n = float64(rv.Int())")
		o.L("case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:")
		o.L("n = float64(rv.Uint())")
		o.L("case reflect.Float32, reflect.Float64:")
		o.L("n = rv.Float()")
		o.L("default:")
		o.L("return nil, fmt.Errorf(`invalid value passed to NumberValidator: expected number, got %%T`, in)")
		o.L("}")
		o.L("")
		o.L("// Reject NaN but allow infinity")
		o.L("if math.IsNaN(n) {")
		o.L("return nil, fmt.Errorf(`invalid value passed to NumberValidator: value is not a valid number (NaN)`)")
		o.L("}")
	}
	o.LL("if m := v.maximum; m != nil {")
	o.L("if n > *m {")
	o.L("return nil, fmt.Errorf(`invalid value passed to %sValidator: value is greater than maximum %%%s`, *m)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if em := v.exclusiveMaximum; em != nil {")
	o.L("if n >= *em {")
	o.L("return nil, fmt.Errorf(`invalid value passed to %sValidator: value is greater than or equal to exclusiveMaximum %%%s`, *em)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if m := v.minimum; m != nil {")
	o.L("if n < *m {")
	o.L("return nil, fmt.Errorf(`invalid value passed to %sValidator: value is less than minimum %%%s`, *m)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if em := v.exclusiveMinimum; em != nil {")
	o.L("if n <= *em {")
	o.L("return nil, fmt.Errorf(`invalid value passed to %sValidator: value is less than or equal to exclusiveMinimum %%%s`, *em)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if mo := v.multipleOf; mo != nil {")
	if def.typ == "int" {
		o.L("if *mo == 0 {")
		o.L("return nil, fmt.Errorf(`invalid value passed to IntegerValidator: multipleOf cannot be zero`)")
		o.L("}")
		o.L("if math.Mod(float64(n), float64(*mo)) != 0 {")
	} else {
		o.L("remainder := math.Mod(n, *mo)")
		o.L("if math.Abs(remainder) > 1e-9 && math.Abs(remainder - *mo) > 1e-9 {")
	}
	o.L("return nil, fmt.Errorf(`invalid value passed to %sValidator: value is not multiple of %%%s`, *mo)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if c := v.constantValue; c != nil {")
	o.L("if *c != n {")
	o.L("return nil, fmt.Errorf(`invalid value passed to %sValidator: value must be const value %%%s`, *c)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if enums := v.enum; len(enums) > 0 {")
	o.L("var found bool")
	o.L("for _, e := range enums {")
	o.L("if e == n {")
	o.L("found = true")
	o.L("break")
	o.L("}")
	o.L("}")
	o.L("if !found {")
	o.L("return nil, fmt.Errorf(`invalid value passed to %sValidator: value not found in enum`)", def.class)
	o.L("}")
	o.L("}")
	o.L("return nil, nil")
	o.L("}")

	fn := filepath.Join(outputDir, def.filename)
	if err := o.WriteFile(fn, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return err
	}
	return nil
}
