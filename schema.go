//go:generate ./gen.sh

package schema

import (
	"unicode"
)

// The schema that this implementation supports. We use the name
// `Version` here because `Schema` is confusin with other types
const Version = `https://json-schema.org/draft/2020-12/schema`

// schemaOrBool implements the SchemaOrBool interface for Schema
func (s *Schema) schemaOrBool() {}

// Predefined field groups for common bit flag checks

// StringConstraintFields groups all string-related validation fields
const StringConstraintFields = MinLengthField | MaxLengthField | PatternField | FormatField

// NumericConstraintFields groups all numeric validation fields
const NumericConstraintFields = MinimumField | MaximumField | MultipleOfField | ExclusiveMinimumField | ExclusiveMaximumField

// ObjectConstraintFields groups all object validation fields
const ObjectConstraintFields = PropertiesField | AdditionalPropertiesField | RequiredField | MinPropertiesField | MaxPropertiesField | PatternPropertiesField | PropertyNamesField | UnevaluatedPropertiesField | DependentSchemasField | DependentRequiredField

// ArrayConstraintFields groups all array validation fields
const ArrayConstraintFields = ItemsField | MinItemsField | MaxItemsField | UniqueItemsField | PrefixItemsField | ContainsField | MinContainsField | MaxContainsField | UnevaluatedItemsField

// CompositionFields groups all composition/logical fields
const CompositionFields = AllOfField | AnyOfField | OneOfField | NotField

// ConditionalFields groups all conditional validation fields
const ConditionalFields = IfSchemaField | ThenSchemaField | ElseSchemaField

// ContentFields groups all content-related validation fields
const ContentFields = ContentEncodingField | ContentMediaTypeField | ContentSchemaField

// BasicPropertiesFields groups basic property-related fields (excluding constraints like minProperties)
const BasicPropertiesFields = PropertiesField | PatternPropertiesField | AdditionalPropertiesField

// UnevaluatedFields groups unevaluated items and properties fields
const UnevaluatedFields = UnevaluatedPropertiesField | UnevaluatedItemsField

// ValueConstraintFields groups enum and const constraint fields
const ValueConstraintFields = EnumField | ConstField

// compareFieldNames compares two field names with custom sorting logic:
// Character-by-character comparison where at each position:
// 1. Non-alphanumeric characters sort before alphanumeric characters
// 2. Within each category (non-alphanumeric or alphanumeric), sort lexicographically
func compareFieldNames(a, b string) bool {
	runesA := []rune(a)
	runesB := []rune(b)

	// Compare character by character up to the length of the longer string
	maxLen := len(runesA)
	if len(runesB) > maxLen {
		maxLen = len(runesB)
	}

	for i := range maxLen {
		var charA, charB rune
		var hasCharA, hasCharB bool

		if i < len(runesA) {
			charA = runesA[i]
			hasCharA = true
		}
		if i < len(runesB) {
			charB = runesB[i]
			hasCharB = true
		}

		// If one string is shorter, the longer string comes later
		if !hasCharA && hasCharB {
			return true // A is shorter, comes first
		}
		if hasCharA && !hasCharB {
			return false // B is shorter, comes first
		}
		if !hasCharA && !hasCharB {
			break // Both strings end here
		}

		isAlphaNumA := unicode.IsLetter(charA) || unicode.IsDigit(charA)
		isAlphaNumB := unicode.IsLetter(charB) || unicode.IsDigit(charB)

		// If one is non-alphanumeric and the other is alphanumeric,
		// non-alphanumeric comes first
		if !isAlphaNumA && isAlphaNumB {
			return true
		}
		if isAlphaNumA && !isAlphaNumB {
			return false
		}

		// Both are in same category (both alphanumeric or both non-alphanumeric)
		// Compare lexicographically (case-insensitive for letters)
		lowerA := unicode.ToLower(charA)
		lowerB := unicode.ToLower(charB)

		if lowerA != lowerB {
			return lowerA < lowerB
		}

		// If lowercase versions are equal, compare original case
		if charA != charB {
			return charA < charB
		}
	}

	// All characters are equal
	return false
}
