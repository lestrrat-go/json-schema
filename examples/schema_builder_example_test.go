package examples_test

import (
	"encoding/json"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

func Example_schema_builder() {
	s, err := schema.NewBuilder().
		Schema(schema.Version).
		ID(`https://example.com/polygon`).
		Types(schema.ObjectType).
		Property("validProp", schema.New()).
		AdditionalProperties(schema.TrueSchema()).
		Build()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(s.ID())
	buf, err := json.Marshal(s)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%s\n", buf)
	// OUTPUT:
	// https://example.com/polygon
	// {"$id":"https://example.com/polygon","$schema":"https://json-schema.org/draft/2020-12/schema","additionalProperties":true,"properties":{"validProp":{}},"type":"object"}
}
