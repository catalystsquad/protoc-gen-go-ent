package plugin

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

func writeFields(g *protogen.GeneratedFile, message *protogen.Message) {
	messageGoName := getMessageGoName(message)
	g.P(fmt.Sprintf("func (%s) Fields() []ent.Field {", messageGoName))
	if len(message.Fields) == 0 {
		g.P("return nil")
	} else {
		g.P("return []ent.Field{")
		for _, field := range message.Fields {
			fieldDefinition := getFieldDefinition(field)
			g.P(fieldDefinition, ",")
		}
		g.P("}")
	}
	g.P(fmt.Sprintf("}"))
}

func getStructFields(message *protogen.Message) []*protogen.Field {
	fields := []*protogen.Field{}
	for _, field := range message.Fields {
		if shouldGenerateField(field) {
			fields = append(fields, field)
		}
	}
}

func getFieldDefinition(field *Field) string {
	builder := strings.Builder{}
	builder.WriteString("field.")
	builder.WriteString(fmt.Sprintf("%s(\"%s\")", field.EntType, field.Name))
	if field.Options.Nillable {
		builder.WriteString(".Nillable()")
	}
	if field.Optional {
		builder.WriteString(".Optional()")
	}
	if field.Options.Default != "" {
		// not using fmt.Sprintf so that we don't get weird formatting
		builder.WriteString(".Default(")
		builder.WriteString(field.Options.Default)
		builder.WriteString(")")
	}
	if field.Options.DefaultFunc != "" {
		// not using fmt.Sprintf so that we don't get weird formatting
		builder.WriteString(".DefaultFunc(")
		builder.WriteString(field.Options.DefaultFunc)
		builder.WriteString(")")
	}
	if field.Options.Unique {
		builder.WriteString(".Unique()")
	}
	if field.Options.Immutable {
		builder.WriteString(".Immutable()")
	}
	if field.Options.Comment != "" {
		// not using fmt.Sprintf so that we don't get weird formatting
		builder.WriteString(".Comment(\"")
		builder.WriteString(field.Options.Comment)
		builder.WriteString("\")")
	}
	if field.Options.StorageKey != "" {
		// not using fmt.Sprintf so that we don't get weird formatting
		builder.WriteString(".StorageKey(\"")
		builder.WriteString(field.Options.StorageKey)
		builder.WriteString("\")")
	}
	if field.Options.StructTag != "" {
		// not using fmt.Sprintf so that we don't get weird formatting
		builder.WriteString(".StructTag(")
		builder.WriteString(field.Options.StructTag)
		builder.WriteString(")")
	}
	if field.IsEnum && len(field.EnumValues) > 0 {
		builder.WriteString(".Values(")
		for i, value := range field.EnumValues {
			if enumValueIsUnspecified(value) && !field.Options.IncludeUnspecifiedEnum {
				continue
			}
			builder.WriteString(fmt.Sprintf("\"%s\"", value))
			if i != len(field.EnumValues)-1 {
				builder.WriteString(", ")
			}
		}
		builder.WriteString(")")
	}
	return builder.String()
}

func enumValueIsUnspecified(value string) bool {
	return strings.HasSuffix(value, "_UNSPECIFIED")
}
