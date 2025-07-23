package schema

// Common Pattern Helpers
//
// This file provides convenient one-liners for frequent JSON Schema validation patterns.
// These functions create pre-configured Builder instances for common use cases.

// Email creates a Builder for email validation using the "email" format
func Email() *Builder {
	return NewBuilder().Types(StringType).Format("email")
}

// URL creates a Builder for URL validation using the "uri" format
func URL() *Builder {
	return NewBuilder().Types(StringType).Format("uri")
}

// NonEmptyString creates a Builder for non-empty string validation
func NonEmptyString() *Builder {
	return NewBuilder().Types(StringType).MinLength(1)
}

// PositiveNumber creates a Builder for positive number validation (>= 0)
func PositiveNumber() *Builder {
	return NewBuilder().Types(NumberType).Minimum(0)
}

// PositiveInteger creates a Builder for positive integer validation (>= 0)
func PositiveInteger() *Builder {
	return NewBuilder().Types(IntegerType).Minimum(0)
}

// Enum creates a Builder for enum validation with the given values
func Enum(values ...any) *Builder {
	return NewBuilder().Enum(values...)
}

// OneOf creates a Builder for oneOf validation with the given schemas
func OneOf(schemas ...*Schema) *Builder {
	schemaOrBools := make([]SchemaOrBool, len(schemas))
	for i, s := range schemas {
		schemaOrBools[i] = s
	}
	return NewBuilder().OneOf(schemaOrBools...)
}

// AnyOf creates a Builder for anyOf validation with the given schemas
func AnyOf(schemas ...*Schema) *Builder {
	schemaOrBools := make([]SchemaOrBool, len(schemas))
	for i, s := range schemas {
		schemaOrBools[i] = s
	}
	return NewBuilder().AnyOf(schemaOrBools...)
}

// AllOf creates a Builder for allOf validation with the given schemas
func AllOf(schemas ...*Schema) *Builder {
	schemaOrBools := make([]SchemaOrBool, len(schemas))
	for i, s := range schemas {
		schemaOrBools[i] = s
	}
	return NewBuilder().AllOf(schemaOrBools...)
}

// Optional creates a Builder that allows either the given schema or null
func Optional(s *Schema) *Builder {
	return NewBuilder().AnyOf(s, NewBuilder().Types(NullType).MustBuild())
}

// UUID creates a Builder for UUID validation using the "uuid" format
func UUID() *Builder {
	return NewBuilder().Types(StringType).Format("uuid")
}

// Date creates a Builder for date validation using the "date" format
func Date() *Builder {
	return NewBuilder().Types(StringType).Format("date")
}

// DateTime creates a Builder for date-time validation using the "date-time" format
func DateTime() *Builder {
	return NewBuilder().Types(StringType).Format("date-time")
}

// AlphanumericString creates a Builder for alphanumeric string validation
func AlphanumericString() *Builder {
	return NewBuilder().Types(StringType).Pattern("^[a-zA-Z0-9]+$")
}
