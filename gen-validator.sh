#!/bin/bash

# gen-validator.sh - Generate pre-compiled meta-schema validator
#
# This script runs the meta-schema compilation tool to generate the 
# meta/meta.go file with a pre-compiled validator for JSON Schema 2020-12.
#
# The generated validator can be used to validate JSON Schema documents
# themselves without needing to build and compile schemas at runtime.

set -e

echo "Generating pre-compiled meta-schema validator..."

# Run the meta-schema generator tool
cd internal/cmd/genmeta
go run main.go

echo "✓ Meta-schema validator generated successfully!"
echo "✓ Generated file: meta/meta.go"
echo ""
echo "Usage example:"
echo "  import \"github.com/lestrrat-go/json-schema/meta\""
echo "  validator := meta.Validator()"
echo "  err := validator.Validate(ctx, jsonSchemaDocument)"