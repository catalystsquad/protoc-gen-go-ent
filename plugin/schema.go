package plugin

import (
	"fmt"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

func generateEntSchemas(gen *protogen.Plugin, objects []SchemaObject) {
	for _, object := range objects {
		generateEntSchema(gen, object)
	}
}

func generateEntSchema(gen *protogen.Plugin, object SchemaObject) {
	g := createSchemaFile(gen, object)
	writeImports(g, object)
	g.P("package schema")
	writeSchema(g, object)
}

func writeImports(g *protogen.GeneratedFile, object SchemaObject) {
	for _, importName := range object.SchemaFileImports {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: protogen.GoImportPath(importName)})
	}
}

func writeSchema(g *protogen.GeneratedFile, object SchemaObject) {
	writeStruct(g, object)
	writeFields(g, object)
	writeEdges(g, object)
	writeAnnotations(g, object)
}

func writeFields(g *protogen.GeneratedFile, object SchemaObject) {
	fieldsDefinition := getFieldsDefinition(g, object)
	g.P(fmt.Sprintf(fieldsTemplate, object.GoType, fieldsDefinition))
}

func getFieldsDefinition(g *protogen.GeneratedFile, object SchemaObject) string {
	if len(object.EntFields) == 0 {
		return ""
	}
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "entgo.io/ent/schema/field"})
	fieldDefinitions := []string{}
	for _, field := range object.EntFields {
		fieldDefinitions = append(fieldDefinitions, getFieldDefinition(g, field))
	}

	return strings.Join(fieldDefinitions, ",\n") + "," // trailing comma
}

func getFieldDefinition(g *protogen.GeneratedFile, field EntField) string {
	parts := []string{}
	parts = append(parts, getFieldTypeDefinition(g, field))
	parts = append(parts, getFieldEnumValuesDefinition(field))
	parts = append(parts, getFieldDefaultDefinition(field))
	parts = append(parts, getFieldDefaultFuncDefinition(field))
	parts = append(parts, getFieldOptionalDefinition(field))
	parts = append(parts, getFieldNillableDefinition(field))
	parts = append(parts, getFieldImmutableDefinition(field))
	parts = append(parts, getFieldUniqueDefinition(field))
	parts = append(parts, getFieldCommentDefinition(field))
	parts = append(parts, getFieldStorageKeyDefinition(field))
	parts = append(parts, getFieldStructTagDefinition(field))
	parts = append(parts, getFieldSensitiveDefinition(field))
	parts = append(parts, getFieldAnnotationsDefinition(field))

	return joinNonEmptyStrings(parts, ".")
}

func getFieldTypeDefinition(g *protogen.GeneratedFile, field EntField) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("field.%s(\"%s\"", field.TypeString, field.Name))
	if field.IsIdField {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/google/uuid"})
		builder.WriteString(", uuid.UUID{}")
	}
	builder.WriteString(")")
	return builder.String()
}

func getFieldEnumValuesDefinition(field EntField) string {
	if len(field.EnumValues) > 0 {
		values := field.EnumValues
		// enclose values in quotes
		values = lo.Map(values, func(item string, index int) string {
			return fmt.Sprintf("\"%s\"", item)
		})
		return fmt.Sprintf("Values(%s)", strings.Join(encloseStringsInQuotes(field.EnumValues), ","))
	}

	return ""
}

func getFieldDefaultDefinition(field EntField) string {
	if field.Default != "" {
		return fmt.Sprintf("Default(%s)", field.Default)
	}
	if field.IsIdField {
		return "Default(uuid.New)"
	}

	return ""
}

func getFieldDefaultFuncDefinition(field EntField) string {
	if field.DefaultFunc != "" {
		return fmt.Sprintf("DefaultFunc(%s)", field.DefaultFunc)
	}

	return ""
}

func getFieldOptionalDefinition(field EntField) string {
	if field.Optional {
		return "Optional()"
	}

	return ""
}

func getFieldNillableDefinition(field EntField) string {
	if field.Nillable {
		return "Nillable()"
	}

	return ""
}

func getFieldImmutableDefinition(field EntField) string {
	if field.Immutable {
		return "Immutable()"
	}

	return ""
}

func getFieldUniqueDefinition(field EntField) string {
	if field.Unique {
		return "Unique()"
	}

	return ""
}

func getFieldCommentDefinition(field EntField) string {
	if field.Comment != "" {
		return fmt.Sprintf("Comment(\"%s\")", field.Comment)
	}

	return ""
}

func getFieldStorageKeyDefinition(field EntField) string {
	if field.StorageKey != "" {
		return fmt.Sprintf("StorageKey(\"%s\")", field.Comment)
	}

	return ""
}

func getFieldStructTagDefinition(field EntField) string {
	if field.StructTag != "" {
		return fmt.Sprintf("StructTag(%s)", field.Comment)
	}

	return ""
}

func getFieldSensitiveDefinition(field EntField) string {
	if field.Sensitive {
		return "Sensitive()"
	}

	return ""
}

func getFieldAnnotationsDefinition(field EntField) string {
	if len(field.Annotations) > 0 {
		return fmt.Sprintf("Annotations(%s)", strings.Join(field.Annotations, ","))
	}

	return ""
}

func writeEdges(g *protogen.GeneratedFile, object SchemaObject) {
	edgesDefinition := getEdgesDefinition(g, object)
	g.P(fmt.Sprintf(edgesTemplate, object.GoType, edgesDefinition))
}

func getEdgesDefinition(g *protogen.GeneratedFile, object SchemaObject) string {
	if len(object.EntEdges) == 0 {
		return ""
	}
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "entgo.io/ent/schema/edge"})
	edgeDefinitions := []string{}
	for _, edge := range object.EntEdges {
		edgeDefinitions = append(edgeDefinitions, getEdgeDefinition(edge))
	}

	return strings.Join(edgeDefinitions, ",\n") + "," // trailing comma
}

func getEdgeDefinition(edge EntEdge) string {
	parts := []string{}
	parts = append(parts, getEdgeTypeDefinition(edge))
	parts = append(parts, getEdgeRefDefinition(edge))
	parts = append(parts, getEdgeUniqueDefinition(edge))
	parts = append(parts, getEdgeFieldDefinition(edge))
	parts = append(parts, getEdgeRequiredDefinition(edge))
	parts = append(parts, getEdgeImmutableDefinition(edge))
	parts = append(parts, getEdgeStorageKeyDefinition(edge))
	parts = append(parts, getEdgeStructTagDefinition(edge))
	parts = append(parts, getEdgeCommentDefinition(edge))
	parts = append(parts, getEdgeAnnotationsDefinition(edge))

	return joinNonEmptyStrings(parts, ".")
}

func getEdgeTypeDefinition(edge EntEdge) string {
	if edge.Ref != "" {
		return fmt.Sprintf("edge.From(\"%s\", %s.Type)", edge.From, edge.FromType)
	} else {
		return fmt.Sprintf("edge.To(\"%s\", %s.Type)", edge.To, edge.ToType)
	}
}

func getEdgeRefDefinition(edge EntEdge) string {
	if edge.Ref != "" {
		return fmt.Sprintf("Ref(%s)", encloseStringInQuotes(edge.Ref))
	}

	return ""
}

func getEdgeUniqueDefinition(edge EntEdge) string {
	if edge.Unique {
		return "Unique()"
	}

	return ""
}

func getEdgeFieldDefinition(edge EntEdge) string {
	if edge.Field != "" {
		return fmt.Sprintf("Field(%s)", encloseStringInQuotes(edge.Field))
	}

	return ""
}

func getEdgeRequiredDefinition(edge EntEdge) string {
	if edge.Required {
		return "Required()"
	}

	return ""
}

func getEdgeImmutableDefinition(edge EntEdge) string {
	if edge.Immutable {
		return "Immutable()"
	}

	return ""
}

func getEdgeStructTagDefinition(edge EntEdge) string {
	if edge.StructTag != "" {
		return fmt.Sprintf("StructTag(%s)", encloseStringInBacktick(edge.StructTag))
	}

	return ""
}

func getEdgeCommentDefinition(edge EntEdge) string {
	if edge.Comment != "" {
		return fmt.Sprintf("Comment(%s)", encloseStringInQuotes(edge.Comment))
	}

	return ""
}

func getEdgeStorageKeyDefinition(edge EntEdge) string {
	sk := edge.StorageKey
	if sk != nil {
		parts := []string{}
		if len(sk.Symbols) > 0 {
			parts = append(parts, fmt.Sprintf("edge.Symbols(%s)", strings.Join(encloseStringsInQuotes(sk.Symbols), ",")))
		}
		if len(sk.Columns) > 0 {
			parts = append(parts, fmt.Sprintf("edge.Columns(%s)", strings.Join(encloseStringsInQuotes(sk.Columns), ",")))
		}
		if sk.Column != "" {
			parts = append(parts, fmt.Sprintf("edge.Column(%s)", encloseStringInQuotes(sk.Column)))
		}
		if sk.Table != "" {
			parts = append(parts, fmt.Sprintf("edge.Table(%s)", encloseStringInQuotes(sk.Table)))
		}
		if sk.Symbol != "" {
			parts = append(parts, fmt.Sprintf("edge.Symbol(%s)", encloseStringInQuotes(sk.Symbol)))
		}
		return fmt.Sprintf("StorageKey(%s)", joinNonEmptyStrings(parts, ","))
	}

	return ""
}

func getEdgeAnnotationsDefinition(edge EntEdge) string {
	if len(edge.Annotations) > 0 {
		return fmt.Sprintf("Annotations(%s)", strings.Join(edge.Annotations, ","))
	}

	return ""
}

func writeAnnotations(g *protogen.GeneratedFile, object SchemaObject) {
	annotationsDefinition := getAnnotationsDefinition(object)
	g.P(fmt.Sprintf(annotationsTemplate, object.GoType, annotationsDefinition))
}

func getAnnotationsDefinition(object SchemaObject) string {
	return strings.Join(object.EntAnnotations, ",\n") + "," // trailing comma
}

func writeStruct(g *protogen.GeneratedFile, object SchemaObject) {
	g.P(fmt.Sprintf(structTemplate, object.GoType))
}

var structTemplate = `
type %s struct {
	ent.Schema
}`

var fieldsTemplate = `
func (%s) Fields() []ent.Field {
	return []ent.Field{
		%s
	}
}`

var edgesTemplate = `
func (%s) Edges() []ent.Edge {
	return []ent.Edge{
		%s
	}
}`

var annotationsTemplate = `
func (%s) Annotations() []schema.Annotation {
	return []schema.Annotation{
		%s
	}
}`

func createSchemaFile(gen *protogen.Plugin, object SchemaObject) *protogen.GeneratedFile {
	g := gen.NewGeneratedFile(object.SchemaFileName, object.SchemaFileGoImportPath)
	return g
}

func encloseStringsInQuotes(values []string) []string {
	return lo.Map(values, func(item string, index int) string {
		return encloseStringInQuotes(item)
	})
}

func encloseStringInQuotes(value string) string {
	if strings.HasPrefix(value, "\"") {
		// if the value is already in quotes, return it as is
		return value
	}
	return fmt.Sprintf("\"%s\"", value)
}

func encloseStringInBacktick(value string) string {
	if strings.HasPrefix(value, "`") {
		// if the value is already in quotes, return it as is
		return value
	}
	return fmt.Sprintf("`%s`", value)
}

func joinNonEmptyStrings(parts []string, separator string) string {
	// filter empty strings
	parts = lo.Filter(parts, func(item string, index int) bool {
		return item != ""
	})

	return strings.Join(parts, separator)
}
