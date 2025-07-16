package field

// Field bit flags for tracking populated fields
type Flag uint64

const (
	AdditionalProperties Flag = 1 << iota
	AllOf
	Anchor
	AnyOf
	Comment
	Const
	Contains
	ContentEncoding
	ContentMediaType
	ContentSchema
	Default
	Definitions
	DependentRequired
	DependentSchemas
	DynamicAnchor
	DynamicReference
	ElseSchema
	Enum
	ExclusiveMaximum
	ExclusiveMinimum
	Format
	ID
	IfSchema
	Items
	MaxContains
	MaxItems
	MaxLength
	MaxProperties
	Maximum
	MinContains
	MinItems
	MinLength
	MinProperties
	Minimum
	MultipleOf
	Not
	OneOf
	Pattern
	PatternProperties
	Properties
	PropertyNames
	Reference
	Required
	ThenSchema
	Types
	UnevaluatedItems
	UnevaluatedProperties
	UniqueItems
	Vocabulary
)
