package gqltypes

import (
	"strings"
)

// IsScalar checks if a type is a GraphQL scalar
func IsScalar(typeName string) bool {
	switch typeName {
	case "Int", "Float", "String", "Boolean", "ID":
		return true
	default:
		return false
	}
}

// IsInput checks if a type name represents an input type
func IsInput(typeName string) bool {
	return strings.HasSuffix(typeName, "Input")
}

// IsEnum checks if a type name represents an enum type
func IsEnum(typeName string) bool {
	// This is a simple check - in practice you might want to maintain a registry of enum types
	return strings.HasSuffix(typeName, "Enum")
}

// GetGraphQLType converts a protobuf type to its GraphQL equivalent
func GetGraphQLType(protoType string, isRepeated bool, isRequired bool) string {
	var baseType string
	switch protoType {
	case "int32", "int64", "sint32", "sint64", "sfixed32", "sfixed64":
		baseType = "Int"
	case "uint32", "uint64", "fixed32", "fixed64":
		baseType = "Int" // GraphQL doesn't have unsigned integers
	case "float", "double":
		baseType = "Float"
	case "bool":
		baseType = "Boolean"
	case "string":
		baseType = "String"
	case "bytes":
		baseType = "String" // Encode bytes as base64 string
	default:
		baseType = protoType // For custom types, use the type name directly
	}

	// Handle repeated fields (arrays)
	if isRepeated {
		baseType = "[" + baseType + "]"
	}

	// Handle required fields (non-null)
	if isRequired {
		baseType = baseType + "!"
	}

	return baseType
}

// FormatComment formats a comment string for GraphQL schema output
func FormatComment(comment string) string {
	if comment == "" {
		return ""
	}

	// Remove any leading/trailing whitespace
	comment = strings.TrimSpace(comment)

	// Ensure the comment starts with """
	if !strings.HasPrefix(comment, `"""`) {
		comment = `"""` + "\n" + comment
	}

	// Ensure the comment ends with """
	if !strings.HasSuffix(comment, `"""`) {
		comment = comment + "\n" + `"""`
	}

	return comment
}

// SanitizeFieldName ensures the field name follows GraphQL naming conventions
func SanitizeFieldName(name string) string {
	// Convert to camelCase if needed
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		for i := 1; i < len(parts); i++ {
			parts[i] = strings.Title(parts[i])
		}
		name = strings.Join(parts, "")
	}

	// Ensure first character is lowercase
	if len(name) > 0 {
		name = strings.ToLower(name[:1]) + name[1:]
	}

	return name
}

// SanitizeTypeName ensures the type name follows GraphQL naming conventions
func SanitizeTypeName(name string) string {
	// Ensure first character is uppercase
	if len(name) > 0 {
		name = strings.Title(name)
	}

	// Remove any invalid characters
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, name)

	return name
}

// Add helper to determine if a type should be federated
func (t *Type) ShouldFederate() bool {
	// Simple heuristic: Types with ID fields should be federated
	for _, f := range t.Fields {
		if strings.HasSuffix(f.Name, "Id") && f.IsRequired {
			t.IsFederatedEntity = true
			t.KeyFields = f.Name
			return true
		}
	}
	return false
}
