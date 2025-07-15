package validator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	schema "github.com/lestrrat-go/json-schema"
)

var _ Interface = (*contentValidator)(nil)

type contentValidator struct {
	contentEncoding   string
	contentMediaType  string
	contentSchema     Interface
}

func compileContentValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	if !s.HasContentEncoding() && !s.HasContentMediaType() && !s.HasContentSchema() {
		return nil, nil // No content validation needed
	}

	cv := &contentValidator{}

	if s.HasContentEncoding() {
		cv.contentEncoding = s.ContentEncoding()
	}

	if s.HasContentMediaType() {
		cv.contentMediaType = s.ContentMediaType()
	}

	if s.HasContentSchema() {
		contentSchemaValidator, err := Compile(ctx, s.ContentSchema())
		if err != nil {
			return nil, fmt.Errorf("failed to compile content schema validator: %w", err)
		}
		cv.contentSchema = contentSchemaValidator
	}

	return cv, nil
}

func (cv *contentValidator) Validate(ctx context.Context, v any) (Result, error) {
	// Content validation only applies to strings
	str, ok := v.(string)
	if !ok {
		// According to JSON Schema spec, content validation is ignored for non-strings
		return nil, nil
	}

	// Apply content encoding (decode the string)
	decodedData := str
	if cv.contentEncoding != "" {
		var err error
		decodedData, err = cv.applyContentDecoding(str, cv.contentEncoding)
		if err != nil {
			// According to JSON Schema spec, encoding errors should be ignored
			// The validation should pass even if decoding fails
			return nil, nil
		}
	}

	// Apply content media type (parse the content)
	var parsedData any = decodedData
	if cv.contentMediaType != "" {
		var err error
		parsedData, err = cv.applyContentMediaType(decodedData, cv.contentMediaType)
		if err != nil {
			// According to JSON Schema spec, media type parsing errors should be ignored
			// The validation should pass even if parsing fails
			return nil, nil
		}
	}

	// Validate against content schema
	// According to JSON Schema 2020-12 spec, content schema validation 
	// is for annotation purposes only and should not affect validation results
	if cv.contentSchema != nil {
		// We could store annotations here in the future, but for now just ignore the result
		_, _ = cv.contentSchema.Validate(ctx, parsedData)
	}

	return nil, nil
}

func (cv *contentValidator) applyContentDecoding(data, encoding string) (string, error) {
	switch strings.ToLower(encoding) {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return "", fmt.Errorf("base64 decoding failed: %w", err)
		}
		return string(decoded), nil
	case "base64url":
		decoded, err := base64.URLEncoding.DecodeString(data)
		if err != nil {
			return "", fmt.Errorf("base64url decoding failed: %w", err)
		}
		return string(decoded), nil
	default:
		// Unknown encoding - just return the original data
		return data, nil
	}
}

func (cv *contentValidator) applyContentMediaType(data, mediaType string) (any, error) {
	switch strings.ToLower(mediaType) {
	case "application/json":
		var result any
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return nil, fmt.Errorf("JSON parsing failed: %w", err)
		}
		return result, nil
	default:
		// Unknown media type - just return the string data
		return data, nil
	}
}