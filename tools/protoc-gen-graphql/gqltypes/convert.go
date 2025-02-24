package gqltypes

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtoKindToGraphQL converts a protobuf kind to a GraphQL type
func ProtoKindToGraphQL(kind protoreflect.Kind, isRepeated bool, isRequired bool) string {
	var baseType string

	switch kind {
	case protoreflect.BoolKind:
		baseType = "Boolean"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		baseType = "Int"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		baseType = "Float"
	case protoreflect.StringKind:
		baseType = "ID" // Use ID for fields ending in _id
	case protoreflect.BytesKind:
		baseType = "String"
	case protoreflect.EnumKind:
		baseType = "String"
	case protoreflect.MessageKind:
		baseType = "Object"
	default:
		baseType = "String"
	}

	if isRepeated {
		baseType = fmt.Sprintf("[%s]", baseType)
	}

	return baseType
}

// ProtoMessageToGraphQL converts a protobuf message to a GraphQL type
func ProtoMessageToGraphQL(msg protoreflect.MessageDescriptor) *Type {
	fields := make([]*Field, 0)
	fieldsDescriptors := msg.Fields()
	for i := 0; i < fieldsDescriptors.Len(); i++ {
		fd := fieldsDescriptors.Get(i)
		field := &Field{
			Name:       string(fd.Name()),
			Type:       ProtoKindToGraphQL(fd.Kind(), fd.IsList(), fd.HasPresence()),
			Comment:    string(fd.Name()),
			IsRequired: fd.HasPresence(),
		}
		fields = append(fields, field)
	}

	return &Type{
		Name:    string(msg.Name()),
		Fields:  fields,
		Comment: string(msg.FullName()),
	}
}

// ProtoEnumToGraphQL converts a protobuf enum to a GraphQL enum
func ProtoEnumToGraphQL(enum protoreflect.EnumDescriptor) *Enum {
	options := make([]*EnumOption, 0)
	for i := 0; i < enum.Values().Len(); i++ {
		ed := enum.Values().Get(i)
		option := &EnumOption{
			Name:    string(ed.Name()),
			Comment: fmt.Sprintf("Value: %d", ed.Number()),
		}
		options = append(options, option)
	}

	return &Enum{
		Name:    string(enum.Name()),
		Options: options,
		Comment: string(enum.FullName()),
	}
}

// ValidateType checks if a type is valid in GraphQL
func ValidateType(typeName string) error {
	if typeName == "" {
		return fmt.Errorf("type name cannot be empty")
	}

	// Check if it's a scalar
	if IsScalar(typeName) {
		return nil
	}

	// Check if it's a list type
	if len(typeName) > 2 && typeName[0] == '[' && typeName[len(typeName)-1] == ']' {
		return ValidateType(typeName[1 : len(typeName)-1])
	}

	// Check if it's a non-null type
	if len(typeName) > 0 && typeName[len(typeName)-1] == '!' {
		return ValidateType(typeName[:len(typeName)-1])
	}

	// For custom types, we assume they are valid if they follow naming conventions
	return nil
}
