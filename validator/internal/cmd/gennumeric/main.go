package main

import (
	"bytes"
	"fmt"
	"os"

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
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func _main() error {
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
		if err := generateValidator(def); err != nil {
			return fmt.Errorf(`failed to generate validator for %s: %w`, def.typ, err)
		}
	}
	return nil
}

func generateValidator(def definition) error {
	props := []string{
		"multipleOf",
		"maximum",
		"exclusiveMaximum",
		"minimum",
		"exclusiveMinimum",
		"constantValue",
	}

	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("package validator")

	o.LL("func compile%sValidator(s *schema.Schema) (Validator, error) {", def.class)
	o.L("b := %s()", def.class)
	for _, prop := range props {
		var methodName string
		if prop == "constantValue" {
			methodName = "Const"
		} else {
			methodName = xstrings.Camel(prop)
		}
		o.LL("if s.Has%s() {", methodName)
		o.L("rv := reflect.ValueOf(s.%s())", methodName)
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
		o.L("panic(`poop`)")
		o.L("}") // switch
		o.L("b.%s(tmp)", methodName)
		o.L("}") // if s.Has
	}
	o.L("return b.Build()")
	o.L("}")

	o.LL("type %sValidator struct {", def.class)
	for _, prop := range props {
		o.L("%s *%s", prop, def.typ)
	}
	o.L("}")

	o.LL("type %sValidatorBuilder struct {", def.class)
	o.L("err error")
	o.L("c *%sValidator", def.class)
	o.L("}")

	o.LL("func %[1]s() *%[1]sValidatorBuilder {", def.class)
	o.L("return &%[1]sValidatorBuilder{c: &%[1]sValidator{}}", def.class)
	o.L("}")

	for _, prop := range props {
		var methodName string
		if prop == "constantValue" {
			methodName = "Const"
		} else {
			methodName = xstrings.Camel(prop)
		}
		o.LL("func (b *%[1]sValidatorBuilder) %[2]s(v %[3]s) *%[1]sValidatorBuilder {", def.class, methodName, def.typ)
		o.L("if b.err != nil {")
		o.L("return b")
		o.L("}")
		o.L("b.c.%s = &v", prop)
		o.L("return b")
		o.L("}")
	}

	o.LL("func (b *%[1]sValidatorBuilder) Build() (*%[1]sValidator, error) {", def.class)
	o.L("if b.err != nil {")
	o.L("return nil, b.err")
	o.L("}")
	o.L("return b.c, nil")
	o.L("}")

	var template string
	if def.typ == "int" {
		template = "d"
	} else {
		template = "f"
	}
	o.LL("func (v *%sValidator) Validate(in interface{}) error {", def.class)
	o.L("rv := reflect.ValueOf(in)")
	if def.typ == "int" {
		o.LL("var n int")
		o.L("switch rv.Kind() {")
		o.L("case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:")
		o.L("n = int(rv.Int())")
		o.L("case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:")
		o.L("n = int(rv.Uint())")
		o.L("default:")
		o.L("return fmt.Errorf(`invalid value passed to IntegerValidator: value is not an integer type (%%T)`, in)")
		o.L("}")
	} else {
		o.L("n := rv.Float()")
	}
	o.LL("if m := v.maximum; m != nil {")
	o.L("if *m <= n {")
	o.L("return fmt.Errorf(`invalid value passed to %sValidator: value is greater than %%%s`, *m)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if em := v.exclusiveMaximum; em != nil {")
	o.L("if *em < n {")
	o.L("return fmt.Errorf(`invalid value passed to %sValidator: value is greater than or equal to %%%s`, *em)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if m := v.minimum; m != nil {")
	o.L("if *m >= n {")
	o.L("return fmt.Errorf(`invalid value passed to %sValidator: value is less than %%%s`, *m)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if em := v.exclusiveMinimum; em != nil {")
	o.L("if *em > n {")
	o.L("return fmt.Errorf(`invalid value passed to %sValidator: value is less than or equal to %%%s`, *em)", def.class, template)
	o.L("}")
	o.L("}")
	o.LL("if mo := v.multipleOf; mo != nil {")
	if def.typ == "int" {
		o.L("if math.Mod(float64(n), float64(*mo)) != 0 {")
	} else {
		o.L("if math.Mod(n, *mo) != 0 {")
	}
	o.L("return fmt.Errorf(`invalid value passed to %sValidator: value is not multiple of %%%s`, *mo)", def.class, template)
	o.L("}")
	o.L("}")
	o.L("return nil")
	o.L("}")

	fn := def.filename
	if err := o.WriteFile(fn, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return err
	}
	return nil
}
