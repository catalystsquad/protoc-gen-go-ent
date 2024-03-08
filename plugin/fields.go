package plugin

import (
	"fmt"
	ent "github.com/catalystsquad/protoc-gen-go-ent/options"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"strings"
)

func WriteFields(g *protogen.GeneratedFile, message *protogen.Message) {
	messageGoName := getMessageGoName(message)
	g.P(fmt.Sprintf("func (%s) Fields() []ent.Field {", messageGoName))
	fields := getStructFields(message)
	if len(fields) == 0 {
		g.P("return nil")
	} else {
		g.P("return []ent.Field{")
		for _, field := range fields {
			writeField(g, field)
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

	return fields
}

func writeField(g *protogen.GeneratedFile, field *protogen.Field) {
	builder := &strings.Builder{}
	builder.WriteString("field.")
	writeFieldType(builder, field)
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
	builder.WriteString(",")
	g.P(builder.String())
}

func writeFieldType(builder *strings.Builder, field *protogen.Field) {
	entType := getFieldEntType(field)
	fieldName := getFieldName(field)
	builder.WriteString(entType)
	builder.WriteString("(\"")
	builder.WriteString(fieldName)
	builder.WriteString("\")")
}

func getFieldEntType(field *protogen.Field) string {
	kind := getFieldKind(field)
	return protoFieldTypeToEntFieldTypeMap[kind]
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

func getFieldName(field *protogen.Field) string {
	return strcase.ToSnake(getFieldGoName(field))
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
		options := getFieldOptions(field)
		if !options.IncludeUnspecifiedEnum {
			values = lo.Filter(values, func(item string, index int) bool {
				return !enumValueIsUnspecified(item)
			})
		}
		values = lo.Map(values, func(item string, index int) string {
			return fmt.Sprintf("\"%s\"", item)
		})
		builder.WriteString(".Values(")
		builder.WriteString(strings.Join(values, ","))
		builder.WriteString(")")
	}
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
	protoreflect.FloatKind:    "Float32",
	protoreflect.DoubleKind:   "Float",
	protoreflect.BytesKind:    "Bytes",
}

func getFieldOptions(field *protogen.Field) ent.EntFieldOptions {
	emptyOptions := ent.EntFieldOptions{EdgeOptions: &ent.EntEdgeOptions{}}
	if field.Desc.Options() == nil {
		//glog.Infof("field %s doesn't have options", getFieldName(field))
		// return empty options
		return emptyOptions
	}
	options, ok := field.Desc.Options().(*descriptorpb.FieldOptions)
	if !ok {
		//glog.Infof("field %s doesn't have field options", getFieldName(field))
		// return empty options
		return emptyOptions
	}

	v := proto.GetExtension(options, ent.E_Field)
	if v == nil {
		// return empty options
		//glog.Infof("field %s doesn't have the right extension", getFieldName(field))
		return emptyOptions
	}

	opts, ok := v.(*ent.EntFieldOptions)
	if !ok || opts == nil {
		//glog.Infof("field %s doesn't have ent field options", getFieldName(field))
		// return empty options
		return emptyOptions
	}
	//glog.Infof("field %s has field options", getFieldName(field))
	return *opts
}

func shouldGenerateField(field *protogen.Field) bool {
	return !fieldIsIgnored(field) && !fieldTypeIsMessage(field) && !fieldTypeIsGroup(field)
}

func fieldTypeIsMessage(field *protogen.Field) bool {
	return getFieldKind(field) == protoreflect.MessageKind
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
