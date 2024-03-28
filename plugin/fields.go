package plugin

import (
	"errors"
	"fmt"
	ent "github.com/catalystsquad/protoc-gen-go-ent/options"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"strconv"
	"strings"
)

func WriteSchemaFileFields(g *protogen.GeneratedFile, message *protogen.Message) error {
	messageGoName := getMessageGoName(message)
	g.P(fmt.Sprintf("func (%s) Fields() []ent.Field {", messageGoName))
	fields := getStructFields(message)
	if len(fields) == 0 {
		g.P("return nil")
	} else {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "entgo.io/ent/schema/field"})
		g.P("return []ent.Field{")
		for _, field := range fields {
			err := writeField(g, field)
			if err != nil {
				return err
			}
		}
		g.P("}")
	}
	g.P(fmt.Sprintf("}"))
	return nil
}

func getStructFields(message *protogen.Message) []*protogen.Field {
	fields := []*protogen.Field{}
	for _, field := range message.Fields {
		if shouldGenerateField(field) {
			fields = append(fields, field)
		}
	}

	return fields
}

func writeField(g *protogen.GeneratedFile, field *protogen.Field) error {
	if isIdField(field) {
		writeIdField(g)
	} else {
		builder := &strings.Builder{}
		builder.WriteString("field.")
		err := writeFieldType(builder, field)
		if err != nil {
			return err
		}
		writeNillable(builder, field)
		writeOptional(builder, field)
		writeDefault(builder, field)
		writeDefaultFunc(builder, field)
		writeFieldUnique(builder, field)
		writeImmutable(builder, field)
		writeComment(builder, field)
		writeStorageKey(builder, field)
		writeStructTag(builder, field)
		writeEnum(builder, field)
		//writeAnnotations(builder, field)
		builder.WriteString(",")
		g.P(builder.String())
	}
	return nil
}

func writeIdField(g *protogen.GeneratedFile) {
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/google/uuid"})
	g.P("field.UUID(\"id\", uuid.UUID{}).Default(uuid.New),")
}

func writeFieldType(builder *strings.Builder, field *protogen.Field) error {
	entType, err := getFieldEntType(field)
	if err != nil {
		return err
	}
	fieldName := getFieldName(field)
	builder.WriteString(entType)
	builder.WriteString("(\"")
	builder.WriteString(fieldName)
	builder.WriteString("\")")
	return nil
}

func isIdField(field *protogen.Field) bool {
	return getFieldProtoName(field) == "id"
}

func getFieldEntType(field *protogen.Field) (entType string, err error) {
	kind := getFieldKind(field)
	var ok bool
	if fieldIsRepeated(field) {
		entType, ok = repeatedProtoFieldTypeToEntFieldTypeMap[kind]
	} else {
		entType, ok = protoFieldTypeToEntFieldTypeMap[kind]
	}
	if !ok {
		err = errors.New(fmt.Sprintf("unsupported type: %s", kind))
	}

	return
}

func getFieldKind(field *protogen.Field) protoreflect.Kind {
	return field.Desc.Kind()
}

func getFieldGoName(field *protogen.Field) string {
	return field.GoName
}

func getFieldProtoName(field *protogen.Field) string {
	return string(field.Desc.Name())
}

func getFieldMessage(field *protogen.Field) *protogen.Message {
	return field.Message
}

func getFieldMessageType(field *protogen.Field) string {
	return getMessageProtoName(getFieldMessage(field))
}

func getFieldEnumType(field *protogen.Field) string {
	return fmt.Sprintf("%s%s", getFieldParentMessageType(field), strcase.ToCamel(getFieldProtoName(field)))
}

func getFieldName(field *protogen.Field) string {
	return strcase.ToSnake(getFieldProtoName(field))
}

func getGraphqlFieldName(field *protogen.Field) string {
	graphqlFieldName := strcase.ToLowerCamel(strcase.ToSnake(getFieldProtoName(field)))
	return graphqlFieldName
}

func shouldUpperCaseLastCharacterOfFieldName(fieldName string) bool {
	return charIsNumber(string(fieldName[len(fieldName)-2])) && !charIsNumber(string(fieldName[len(fieldName)-1]))
}

func charIsNumber(char string) bool {
	_, err := strconv.Atoi(char)
	return err == nil
}

func writeNillable(builder *strings.Builder, field *protogen.Field) {
	options := getFieldOptions(field)
	if options.Nillable {
		builder.WriteString(".Nillable()")
	}
}

func writeOptional(builder *strings.Builder, field *protogen.Field) {
	optional := fieldIsOptional(field)
	if optional {
		builder.WriteString(".Optional()")
	}
}

func writeDefault(builder *strings.Builder, field *protogen.Field) {
	options := getFieldOptions(field)
	if options.Default != "" {
		builder.WriteString(".Default(")
		builder.WriteString(options.Default)
		builder.WriteString(")")
	}
}

func writeDefaultFunc(builder *strings.Builder, field *protogen.Field) {
	options := getFieldOptions(field)
	if options.DefaultFunc != "" {
		builder.WriteString(".DefaultFunc(")
		builder.WriteString(options.DefaultFunc)
		builder.WriteString(")")
	}
}

func writeFieldUnique(builder *strings.Builder, field *protogen.Field) {
	options := getFieldOptions(field)
	if options.Unique {
		builder.WriteString(".Unique()")
	}
}

func writeImmutable(builder *strings.Builder, field *protogen.Field) {
	options := getFieldOptions(field)
	if options.Immutable {
		builder.WriteString(".Immutable()")
	}
}

func writeComment(builder *strings.Builder, field *protogen.Field) {
	options := getFieldOptions(field)
	if options.Comment != "" {
		builder.WriteString(".Comment(\"")
		builder.WriteString(options.Comment)
		builder.WriteString("\")")
	}
}

func writeStorageKey(builder *strings.Builder, field *protogen.Field) {
	options := getFieldOptions(field)
	if options.StorageKey != "" {
		builder.WriteString(".StorageKey(\"")
		builder.WriteString(options.StorageKey)
		builder.WriteString("\")")
	}
}

func writeStructTag(builder *strings.Builder, field *protogen.Field) {
	options := getFieldOptions(field)
	if options.StructTag != "" {
		builder.WriteString(".StructTag(")
		builder.WriteString(options.StructTag)
		builder.WriteString(")")
	}
}

func writeEnum(builder *strings.Builder, field *protogen.Field) {
	if fieldIsEnum(field) {
		values := getFieldEnumValues(field)
		values = lo.Map(values, func(item string, index int) string {
			return fmt.Sprintf("\"%s\"", item)
		})
		builder.WriteString(".Values(")
		builder.WriteString(strings.Join(values, ","))
		builder.WriteString(")")
	}
}

//func writeAnnotations(builder *strings.Builder, field *protogen.Field) {
//	if fieldIsRepeated(field) {
//		return
//	}
//	annotations := []string{}
//	annotations = append(annotations, getOrderFieldDefinition(getOrderFieldName(field)))
//	graphqlType, ok := protoToGraphqlTypeMap[getFieldKind(field)]
//	if ok {
//		annotations = append(annotations, fmt.Sprintf("entgql.Type(\"%s\")", graphqlType))
//	}
//	builder.WriteString(fmt.Sprintf(".Annotations(%s)", strings.Join(annotations, ",")))
//}

// entgql doesn't support all types out of the box, these types must be paired with their respective gqlgen type
// which is specified in gqlgen.yml, and the entgql.Type() must be specified, or the code is not generated correctly
var protoToGraphqlTypeMap = map[protoreflect.Kind]string{
	protoreflect.Uint32Kind:  "Uint32",
	protoreflect.Fixed32Kind: "Uint32",
	protoreflect.Uint64Kind:  "Uint64",
	protoreflect.Fixed64Kind: "Uint64",
}

func getOrderFieldName(field *protogen.Field) string {
	return strings.ToUpper(strcase.ToSnake(getFieldProtoName(field)))
}

func fieldIsEnum(field *protogen.Field) bool {
	kind := getFieldKind(field)
	return kind == protoreflect.EnumKind
}

func getFieldEnumValues(field *protogen.Field) []string {
	values := []string{}
	for _, enumValue := range field.Enum.Values {
		value := getEnumValue(enumValue)
		values = append(values, value)
	}
	if !getFieldOptions(field).IncludeUnspecifiedEnum {
		values = lo.Filter(values, func(item string, index int) bool {
			return !enumValueIsUnspecified(item)
		})
	}
	return values
}

func getEnumValue(enumValue *protogen.EnumValue) string {
	return string(enumValue.Desc.Name())
}

func enumValueIsUnspecified(value string) bool {
	return strings.HasSuffix(value, "_UNSPECIFIED")
}

var protoFieldTypeToEntFieldTypeMap = map[protoreflect.Kind]string{
	protoreflect.StringKind:   "String",
	protoreflect.BoolKind:     "Bool",
	protoreflect.EnumKind:     "Enum",
	protoreflect.Int32Kind:    "Int32",
	protoreflect.Sint32Kind:   "Int32",
	protoreflect.Uint32Kind:   "Uint32",
	protoreflect.Sfixed32Kind: "Int32",
	protoreflect.Fixed32Kind:  "Uint32",
	protoreflect.Int64Kind:    "Int64",
	protoreflect.Sint64Kind:   "Int64",
	protoreflect.Uint64Kind:   "Uint64",
	protoreflect.Sfixed64Kind: "Int64",
	protoreflect.Fixed64Kind:  "Uint64",
	protoreflect.FloatKind:    "Float",
	protoreflect.DoubleKind:   "Float",
	protoreflect.BytesKind:    "Bytes",
}

var repeatedProtoFieldTypeToEntFieldTypeMap = map[protoreflect.Kind]string{
	protoreflect.StringKind:   "Strings",
	protoreflect.Int32Kind:    "Ints",
	protoreflect.Sint32Kind:   "Ints",
	protoreflect.Uint32Kind:   "Ints",
	protoreflect.Sfixed32Kind: "Ints",
	protoreflect.Fixed32Kind:  "Ints",
	protoreflect.Int64Kind:    "Ints",
	protoreflect.Sint64Kind:   "Ints",
	protoreflect.Uint64Kind:   "Ints",
	protoreflect.Sfixed64Kind: "Ints",
	protoreflect.Fixed64Kind:  "Ints",
	protoreflect.FloatKind:    "Floats",
	protoreflect.DoubleKind:   "Floats",
}

func getFieldOptions(field *protogen.Field) ent.EntFieldOptions {
	emptyOptions := ent.EntFieldOptions{EdgeOptions: &ent.EntEdgeOptions{}}
	if field.Desc.Options() == nil {
		// return empty options
		return emptyOptions
	}
	options, ok := field.Desc.Options().(*descriptorpb.FieldOptions)
	if !ok {
		// return empty options
		return emptyOptions
	}

	v := proto.GetExtension(options, ent.E_EntFieldOpts)
	if v == nil {
		// return empty options
		return emptyOptions
	}

	opts, ok := v.(*ent.EntFieldOptions)
	if !ok || opts == nil {
		// return empty options
		return emptyOptions
	}
	return *opts
}

func shouldGenerateField(field *protogen.Field) bool {
	return !fieldIsIgnored(field) && !fieldIsMessage(field) && !fieldTypeIsGroup(field)
}

func fieldIsMessage(field *protogen.Field) bool {
	return getFieldKind(field) == protoreflect.MessageKind
}

func fieldIsIncludedMessage(field *protogen.Field) bool {
	return fieldIsMessage(field) && fieldIsIncludedInSource(field)
}

func fieldIsTimestamp(field *protogen.Field) bool {
	return fieldIsMessage(field) && string(field.Message.Desc.FullName()) == "google.protobuf.Timestamp"
}

func fieldIsIncludedInSource(field *protogen.Field) bool {
	return lo.Contains(request.FileToGenerate, field.Message.Location.SourceFile)
}

func fieldTypeIsEnum(field *protogen.Field) bool {
	return getFieldKind(field) == protoreflect.EnumKind
}

func fieldTypeIsGroup(field *protogen.Field) bool {
	return getFieldKind(field) == protoreflect.GroupKind
}

func fieldIsIgnored(field *protogen.Field) bool {
	fieldOptions := getFieldOptions(field)
	return fieldOptions.Ignore
}

func fieldIsOptional(protogenField *protogen.Field) bool {
	return protogenField.Desc.HasOptionalKeyword()
}

func fieldIsRepeated(field *protogen.Field) bool {
	return field.Desc.IsList()
}

func getFieldParentMessage(field *protogen.Field) *protogen.Message {
	return field.Parent
}

func getFieldParentMessageType(field *protogen.Field) string {
	return getMessageProtoName(getFieldParentMessage(field))
}

func getQualifiedProtoFieldName(field *protogen.Field) string {
	return fmt.Sprintf("%s.%s", getFieldParentMessageType(field), getFieldProtoName(field))
}
