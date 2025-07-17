package validator_test

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/schemas
var metaSchemaFS embed.FS

// setupMetaschemaServer creates a local HTTP server that serves metaschema files using embed.FS
func setupMetaschemaServer(t *testing.T) *httptest.Server {
	server := httptest.NewUnstartedServer(nil)
	mux := http.NewServeMux()
	
	// Map metaschema paths to template files 
	metaschemaFiles := map[string]string{
		"/draft/2020-12/schema":              "meta/schema.json.tpl",
		"/draft/2020-12/meta/core":           "meta/core.json.tpl", 
		"/draft/2020-12/meta/applicator":     "meta/applicator.json.tpl",
		"/draft/2020-12/meta/unevaluated":    "meta/unevaluated.json.tpl",
		"/draft/2020-12/meta/validation":     "meta/validation.json.tpl",
		"/draft/2020-12/meta/meta-data":      "meta/meta-data.json.tpl",
		"/draft/2020-12/meta/format-annotation": "meta/format-annotation.json.tpl",
		"/draft/2020-12/meta/content":        "meta/content.json.tpl",
		"/draft/2020-12/meta/format":         "meta/format.json.tpl",
	}
	
	for path, templateFile := range metaschemaFiles {
		// Capture variables in closure
		path := path
		templateFile := templateFile
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			t.Logf("HTTP Server: Serving %s from template %s", path, templateFile)
			
			// Read template file from embed.FS
			data, err := metaSchemaFS.ReadFile("testdata/schemas/" + templateFile)
			if err != nil {
				t.Logf("Failed to read template file %s: %v", templateFile, err)
				http.Error(w, fmt.Sprintf("Failed to read template: %v", err), http.StatusInternalServerError)
				return
			}
			
			// Parse template and execute with server URL
			tmpl, err := template.New("schema").Parse(string(data))
			if err != nil {
				t.Logf("Failed to parse template %s: %v", templateFile, err)
				http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
				return
			}
			
			templateData := struct {
				ServerURL string
			}{
				ServerURL: server.URL,
			}
			
			var buf strings.Builder
			if err := tmpl.Execute(&buf, templateData); err != nil {
				t.Logf("Failed to execute template %s: %v", templateFile, err)
				http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
				return
			}
			
			result := buf.String()
			t.Logf("HTTP Server: Successfully serving %d bytes for %s", len(result), path)
			// Check for allOf in the content
			if strings.Contains(result, "\"allOf\"") {
				t.Logf("HTTP Server: ✓ allOf field found in served content for %s", path)
			} else {
				t.Logf("HTTP Server: ✗ allOf field NOT found in served content for %s", path)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(result))
		})
	}
	
	// Add catch-all handler to log any unexpected requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("HTTP Server: Unexpected request to %s", r.URL.Path)
		http.NotFound(w, r)
	})
	
	server.Config.Handler = mux
	server.Start()
	t.Logf("Started metaschema server at %s", server.URL)
	return server
}

// TestMetaschemaValidation tests that raw JSON data is properly validated against
// the JSON Schema metaschema to catch semantic errors that should be rejected
// according to the JSON Schema 2020-12 specification.
//
// This test uses map[string]any (raw data) instead of attempting to unmarshal
// into a Schema struct because:
//
//  1. Our Schema.UnmarshalJSON correctly rejects malformed schemas like {"type": 1}
//     at the parsing level with errors like "cannot unmarshal number into Go value of type string"
//
//  2. However, metaschema validation should happen at the SEMANTIC level, where the
//     validator validates raw JSON against metaschema rules to determine if the JSON
//     represents a valid JSON Schema according to the specification
//
//  3. The failing defs.json test expects that when validating raw data like
//     {"$defs": {"foo": {"type": 1}}} against a schema that references the metaschema,
//     the validator should detect that {"type": 1} violates the metaschema rules
//     (type must be string or array of strings, not number)
//
//  4. This separation allows the validator to validate potentially malformed schemas
//     and provide proper metaschema validation errors, rather than just parsing errors
func TestMetaschemaValidation(t *testing.T) {
	// Set up local metaschema server  
	server := setupMetaschemaServer(t)
	defer server.Close()
	
	t.Run("Validate invalid type definition against metaschema", func(t *testing.T) {
		// This schema references the local metaschema server - any data validated
		// against it should conform to the JSON Schema 2020-12 specification
		metaschemaRefJSON := fmt.Sprintf(`{
			"$schema": "%s/draft/2020-12/schema",
			"$ref": "%s/draft/2020-12/schema"
		}`, server.URL, server.URL)

		var metaschemaRef schema.Schema
		require.NoError(t, metaschemaRef.UnmarshalJSON([]byte(metaschemaRefJSON)))

		// Log the schema before compilation to see if allOf is present
		t.Logf("Schema before compilation - HasAllOf: %v, HasReference: %v", metaschemaRef.HasAllOf(), metaschemaRef.HasReference())
		if metaschemaRef.HasReference() {
			t.Logf("Reference: %s", metaschemaRef.Reference())
		}
		if metaschemaRef.HasAllOf() {
			t.Logf("AllOf length: %d", len(metaschemaRef.AllOf()))
		}
		
		v, err := validator.Compile(context.Background(), &metaschemaRef)
		require.NoError(t, err)
		t.Logf("Compiled validator type: %T", v)

		// IMPORTANT: We use map[string]any here instead of trying to unmarshal {"type": 1}
		// into a Schema struct because:
		// - Schema.UnmarshalJSON would reject this at parsing time with a type error
		// - But metaschema validation should happen at validation time against raw data
		// - The validator needs to detect that "type": 1 violates JSON Schema rules
		//   (type field must be string like "string" or array like ["string", "number"])
		invalidTypeData := map[string]any{
			"type": 1, // Invalid according to JSON Schema spec: type must be string or array of strings
		}

		// Expected: validation should fail with a metaschema validation error
		// Current issue: validation passes when it should detect the spec violation
		_, err = v.Validate(context.Background(), invalidTypeData)
		if err == nil {
			t.Errorf("Expected validation to FAIL for invalid type definition, but it PASSED")
			t.Logf("Data that should be invalid according to JSON Schema spec: %+v", invalidTypeData)
		} else {
			t.Logf("Validation correctly FAILED with metaschema error: %v", err)
		}
	})

	t.Run("Validate defs.json test case against metaschema", func(t *testing.T) {
		// This reproduces the exact failing test case from the JSON Schema test suite:
		// tests/draft2020-12/defs.json -> "validate definition against metaschema" -> "invalid definition schema"
		//
		// The test validates that schemas containing $defs are themselves valid according to the metaschema.
		// When a schema definition violates the JSON Schema spec, validation should fail.
		metaschemaRefJSON := fmt.Sprintf(`{
			"$schema": "%s/draft/2020-12/schema",
			"$ref": "%s/draft/2020-12/schema"
		}`, server.URL, server.URL)

		var metaschemaRef schema.Schema
		require.NoError(t, metaschemaRef.UnmarshalJSON([]byte(metaschemaRefJSON)))

		// Log the schema before compilation to see if allOf is present
		t.Logf("Schema before compilation - HasAllOf: %v, HasReference: %v", metaschemaRef.HasAllOf(), metaschemaRef.HasReference())
		if metaschemaRef.HasReference() {
			t.Logf("Reference: %s", metaschemaRef.Reference())
		}
		if metaschemaRef.HasAllOf() {
			t.Logf("AllOf length: %d", len(metaschemaRef.AllOf()))
		}

		v, err := validator.Compile(context.Background(), &metaschemaRef)
		require.NoError(t, err)

		// IMPORTANT: This uses map[string]any to represent the raw JSON data from the test suite.
		// The failing test case has data: {"$defs": {"foo": {"type": 1}}} where "type": 1 is invalid.
		// We can't unmarshal this into a Schema struct because Schema.UnmarshalJSON correctly
		// rejects {"type": 1} at parsing time. But the metaschema validation should happen at
		// the semantic validation level to catch that the nested "foo" definition violates
		// the JSON Schema specification (type must be string or array, not number).
		invalidDefsData := map[string]any{
			"$defs": map[string]any{
				"foo": map[string]any{
					"type": 1, // Invalid according to JSON Schema spec: type must be string or array of strings
				},
			},
		}

		// Expected: validation should fail because the "foo" definition violates metaschema rules
		// Current issue: validation passes when it should detect the nested schema violation
		_, err = v.Validate(context.Background(), invalidDefsData)
		if err == nil {
			t.Errorf("Expected validation to FAIL for invalid definition schema, but it PASSED")
			t.Logf("This reproduces the exact defs.json test suite failure")
			t.Logf("Data that should be invalid according to metaschema: %+v", invalidDefsData)
		} else {
			t.Logf("Validation correctly FAILED with metaschema error: %v", err)
		}
	})

	t.Run("Validate valid schema against metaschema", func(t *testing.T) {
		// Positive control test - this should pass to ensure our metaschema validation
		// works correctly for valid schemas (not just rejecting invalid ones)
		metaschemaRefJSON := fmt.Sprintf(`{
			"$schema": "%s/draft/2020-12/schema",
			"$ref": "%s/draft/2020-12/schema"
		}`, server.URL, server.URL)

		var metaschemaRef schema.Schema
		require.NoError(t, metaschemaRef.UnmarshalJSON([]byte(metaschemaRefJSON)))

		// Log the schema before compilation to see if allOf is present
		t.Logf("Schema before compilation - HasAllOf: %v, HasReference: %v", metaschemaRef.HasAllOf(), metaschemaRef.HasReference())
		if metaschemaRef.HasReference() {
			t.Logf("Reference: %s", metaschemaRef.Reference())
		}
		if metaschemaRef.HasAllOf() {
			t.Logf("AllOf length: %d", len(metaschemaRef.AllOf()))
		}

		v, err := validator.Compile(context.Background(), &metaschemaRef)
		require.NoError(t, err)

		// This uses map[string]any for a valid schema that should pass metaschema validation.
		// This serves as a positive control to ensure that when we fix the metaschema validation
		// to properly reject invalid schemas, we don't break validation of valid schemas.
		validSchemaData := map[string]any{
			"type":      "string", // Valid: type as string (correct according to JSON Schema spec)
			"minLength": 1,        // Valid: minLength constraint with proper type
		}

		_, err = v.Validate(context.Background(), validSchemaData)
		if err != nil {
			t.Errorf("Expected validation to PASS for valid schema, but it FAILED: %v", err)
		} else {
			t.Logf("Validation correctly PASSED for valid schema (positive control)")
		}
	})
}
