package plugin

import (
	"errors"
	"fmt"
	"github.com/catalystsquad/protoc-gen-go-ent/config"
	ent "github.com/catalystsquad/protoc-gen-go-ent/options"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/golang/glog"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"strings"
)

type EntFieldType int

const (
	EntFieldTypeUnknown EntFieldType = iota
	EntFieldTypeEnum
	EntFieldTypeString
	EntFieldTypeUUID
	EntFieldTypeFloat32
	EntFieldTypeBool
	EntFieldTypeFloat
	EntFieldTypeTime
	EntFieldTypeAny
	EntFieldTypeBytes
	EntFieldTypeFloats
	EntFieldTypeInt
	EntFieldTypeInt8
	EntFieldTypeInt16
	EntFieldTypeInt32
	EntFieldTypeInt64
	EntFieldTypeInts
	EntFieldTypeJson
	EntFieldTypeOther
	EntFieldTypeStrings
	EntFieldTypeText
	EntFieldTypeUint
	EntFieldTypeUint8
	EntFieldTypeUint16
	EntFieldTypeUint32
	EntFieldTypeUint64
	EntFieldTypeId
)

type SchemaObject struct {
	GoType                 string
	SchemaFileName         string
	SchemaFileGoImportPath protogen.GoImportPath
	SchemaFileImports      []string
	GraphqlSchemaFileName  string
	ResolverFileName       string
	EntFields              []EntField
	EntEdges               []EntEdge
	EntAnnotations         []string
	PluralName             string
	GraphqlTypes           *hashset.Set
}

type EntField struct {
	Type            EntFieldType
	TypeString      string
	FullGraphqlType string
	GraphqlTypeName string
	Name            string
	IsIdField       bool
	GoType          string
	Default         string
	DefaultFunc     string
	Optional        bool
	Nillable        bool
	Immutable       bool
	Unique          bool
	Comment         string
	StorageKey      string
	StructTag       string
	Sensitive       bool
	EnumValues      []string
	Annotations     []string
}

type EntEdge struct {
	To          string
	ToType      string
	From        string
	FromType    string
	Ref         string
	Unique      bool
	Field       string
	Required    bool
	Immutable   bool
	StorageKey  *ent.EdgeStorageKey
	StructTag   string
	Comment     string
	Annotations []string
}

var request *pluginpb.CodeGeneratorRequest

func parseFiles(gen *protogen.Plugin) ([]SchemaObject, error) {
	request = gen.Request
	objects := []SchemaObject{}
	for _, f := range gen.Files {
		if !getFileOptions(f).Generate {
			continue
		}
		fileObjects, err := ParseFile(gen, f)
		if err != nil {
			return nil, err
		}
		objects = append(objects, fileObjects...)
	}

	return objects, nil
}

func ParseFile(gen *protogen.Plugin, file *protogen.File) ([]SchemaObject, error) {
	objects := []SchemaObject{}
	for _, message := range file.Messages {
		if !shouldHandleMessage(message) {
			continue
		}
		object, err := ParseMessage(message)
		if err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	return objects, nil
}

func ParseMessage(message *protogen.Message) (SchemaObject, error) {
	schemaObject := SchemaObject{
		GoType:                 getEntSchemaGoType(message),
		SchemaFileName:         getEntSchemaFileName(message),
		GraphqlSchemaFileName:  getGraphqlSchemaFileName(message),
		ResolverFileName:       getObjectResolverFileName(message),
		SchemaFileGoImportPath: getEntSchemaFileGoImportPath(message),
		SchemaFileImports:      getEntSchemaFileImports(message),
		EntAnnotations:         getEntSchemaAnnotations(message),
		PluralName:             getPluralMessageProtoName(message),
	}
	fields, err := getEntSchemaFields(message)
	if err != nil {
		return schemaObject, err
	}
	edges, err := getEntSchemaEdges(message)
	if err != nil {
		return schemaObject, err
	}

	schemaObject.EntEdges = edges
	schemaObject.EntFields = fields
	schemaObject.GraphqlTypes = getSchemaGraphqlTypes(schemaObject)
	return schemaObject, nil
}

func getSchemaGraphqlTypes(schema SchemaObject) *hashset.Set {
	graphqlTypes := hashset.New()
	for _, field := range schema.EntFields {
		graphqlTypes.Add(field.GraphqlTypeName)
	}

	return graphqlTypes
}

func getEntSchemaGoType(message *protogen.Message) string {
	return getMessageGoName(message)
}

func getEntSchemaFileName(message *protogen.Message) string {
	return fmt.Sprintf("%s/%s.pb.ent.go", *config.EntSchemaDir, strcase.ToSnake(getMessageProtoName(message)))
}

func getGraphqlSchemaFileName(message *protogen.Message) string {
	return fmt.Sprintf("%s/%s.graphql", *config.GraphqlSchemaDir, strcase.ToLowerCamel(getMessageProtoName(message)))
}

func getObjectResolverFileName(message *protogen.Message) string {
	res := fmt.Sprintf("%s/%s.resolvers.go", *config.GraphqlResolverDir, strcase.ToLowerCamel(getMessageProtoName(message)))
	return res
}

func getEntSchemaFileGoImportPath(message *protogen.Message) protogen.GoImportPath {
	return "."
}

func getEntSchemaFileImports(message *protogen.Message) []string {
	schemaFileImports := []string{
		"entgo.io/ent",
		"entgo.io/contrib/entgql",
		"entgo.io/ent/schema",
	}
	if shouldImportTime(message) {
		schemaFileImports = append(schemaFileImports, "time")
	}
	return schemaFileImports
}

func shouldImportTime(message *protogen.Message) bool {
	for _, field := range message.Fields {
		if strings.Contains(getFieldOptions(field).Default, "time.") {
			return true
		}
	}
	return false
}

func getEntSchemaFields(message *protogen.Message) ([]EntField, error) {
	fields := []EntField{}
	nonMessageFields := getNonMessageFields(message)
	for _, field := range nonMessageFields {
		entField, err := getEntFieldFromProtoField(field)
		if err != nil {
			return nil, err
		}
		fields = append(fields, entField)
	}

	return fields, nil
}

func getEntSchemaEdges(message *protogen.Message) ([]EntEdge, error) {
	edges := []EntEdge{}
	messageFields := getMessageFields(message)
	for _, field := range messageFields {
		edge, err := getEntEdgeFromProtoField(field)
		if err != nil {
			return nil, err
		}
		edges = append(edges, edge)
	}

	return edges, nil
}

func getEntSchemaAnnotations(message *protogen.Message) []string {
	annotations := []string{
		"entgql.QueryField()",
		"entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate())",
		"entgql.MultiOrder()",
		"entgql.RelayConnection()",
	}

	return annotations
}

func getEntFieldFromProtoField(field *protogen.Field) (EntField, error) {
	entField := EntField{
		Name:        getEntFieldName(field),
		IsIdField:   entFieldIsId(field),
		GoType:      getEntFieldGoType(field),
		Default:     getEntFieldDefault(field),
		DefaultFunc: getEntFieldDefaultFunc(field),
		Optional:    getEntFieldOptional(field),
		Nillable:    getEntFieldNillable(field),
		Immutable:   getEntFieldImmutable(field),
		Unique:      getEntFieldUnique(field),
		Comment:     getEntFieldComment(field),
		StorageKey:  getEntFieldStorageKey(field),
		StructTag:   getEntFieldStructTag(field),
		Sensitive:   getEntFieldSensitive(field),
		EnumValues:  getEntFieldEnumValues(field),
	}

	fullGraphqlType, err := getFullGraphqlType(field)
	if err != nil {
		return entField, err
	}
	entField.FullGraphqlType = fullGraphqlType

	graphqlTypeName, err := getGraphlTypeName(field)
	if err != nil {
		return entField, err
	}
	entField.GraphqlTypeName = graphqlTypeName

	entType, err := getEntFieldType(field)
	if err != nil {
		return entField, err
	}
	entField.Type = entType

	entTypeString, err := getEntTypeString(field)
	if err != nil {
		return entField, err
	}

	entField.TypeString = entTypeString

	annotations, err := getEntFieldAnnotations(field)
	if err != nil {
		return entField, err
	}

	entField.Annotations = annotations

	return entField, nil
}

func getGraphlTypeName(field *protogen.Field) (string, error) {
	fullGraphqlType, err := getFullGraphqlType(field)
	if err != nil {
		return "", err
	}
	return strings.Trim(fullGraphqlType, "[]!"), nil
}

func getEntEdgeFromProtoField(field *protogen.Field) (EntEdge, error) {
	edge := EntEdge{
		To:         getEntEdgeTo(field),
		ToType:     getEntEdgeToType(field),
		From:       getEntEdgeFrom(field),
		FromType:   getEntEdgeFromType(field),
		Ref:        getEntEdgeRef(field),
		Field:      getEntEdgeField(field),
		Required:   getEntEdgeRequired(field),
		Immutable:  getEntEdgeImmutable(field),
		StorageKey: getEntEdgeStorageKey(field),
		StructTag:  getEntEdgeStructTag(field),
		Comment:    getEntEdgeComment(field),
	}

	unique, err := getEntEdgeUnique(field)
	if err != nil {
		return edge, err
	}
	edge.Unique = unique
	annotations, err := getEntEdgeAnnotations(field)
	if err != nil {
		return edge, err
	}
	edge.Annotations = annotations

	return edge, nil
}

func getEntEdgeTo(field *protogen.Field) string {
	override := getEdgeOptions(field).To
	if override != "" {
		return override
	}
	bidirectional := edgeIsBidirectional(field)
	if bidirectional || getEdgeOptions(field).Ref == "" {
		return getFieldProtoName(field)
	}
	return ""
}

func getEntEdgeAnnotations(field *protogen.Field) ([]string, error) {
	annotations := []string{}
	unique, err := getEntEdgeUnique(field)
	if err != nil {
		return nil, err
	}
	if !unique {
		// non unique edges get the relay connection annotation
		annotations = append(annotations, "entgql.RelayConnection()")
		// non unique edges get the order by count annotation
		annotations = append(annotations, fmt.Sprintf("entgql.OrderField(\"%s_COUNT\")", getOrderFieldName(field)))
	} else {
		// unique edges can be ordered by the other edge fields
		fieldMessage := getFieldMessage(field)
		orderFieldPrefix := getOrderFieldName(field)
		for _, field := range fieldMessage.Fields {
			if !isIdField(field) && !fieldIsMessage(field) {
				orderFieldName := fmt.Sprintf("%s_%s", orderFieldPrefix, getOrderFieldName(field))
				annotations = append(annotations, getOrderFieldDefinition(orderFieldName))
			}
		}
	}
	return annotations, nil
}

func getOrderFieldDefinition(name string) string {
	return fmt.Sprintf("entgql.OrderField(\"%s\")", strcase.ToScreamingSnake(name))
}

func getEntEdgeComment(field *protogen.Field) string {
	return getEdgeOptions(field).Comment
}

func getEntEdgeField(field *protogen.Field) string {
	opts := getEdgeOptions(field)
	if opts.Bind {
		override := opts.BindField
		if override != "" {
			return override
		}
		return fmt.Sprintf("%s_id", getEntFieldName(field))
	}

	return ""
}

func getEntEdgeFrom(field *protogen.Field) string {
	override := getEdgeOptions(field).From
	if override != "" {
		return override
	}
	if !edgeIsBidirectional(field) && getEntEdgeRef(field) != "" {
		return getFieldProtoName(field)
	}

	return ""
}

func getEntEdgeFromType(field *protogen.Field) string {
	if getEntEdgeRef(field) != "" {
		return getMessageProtoName(getFieldMessage(field))
	}

	return ""
}

func getEntEdgeImmutable(field *protogen.Field) bool {
	return getEdgeOptions(field).Immutable
}

func getEntEdgeRef(field *protogen.Field) string {
	if !edgeIsBidirectional(field) {
		return getEdgeOptions(field).Ref
	}

	return ""
}

func getEntEdgeRequired(field *protogen.Field) bool {
	return getEdgeOptions(field).Required
}

func getEntEdgeStorageKey(field *protogen.Field) *ent.EdgeStorageKey {
	return getEdgeOptions(field).StorageKey
}

func getEntEdgeStructTag(field *protogen.Field) string {
	return getEdgeOptions(field).StructTag
}

func getEntEdgeToType(field *protogen.Field) string {
	return getMessageProtoName(getFieldMessage(field))
}

func getEntEdgeUnique(field *protogen.Field) (bool, error) {
	isUnique, err := edgeIsUnique(field)
	if err != nil {
		return false, err
	}

	return isUnique, nil
}

func getEntFieldNillable(field *protogen.Field) bool {
	return getFieldOptions(field).Nillable
}

func entFieldIsId(field *protogen.Field) bool {
	return isIdField(field)
}

func getEntFieldName(field *protogen.Field) string {
	return strcase.ToSnake(getFieldProtoName(field))
}

func getEntFieldType(field *protogen.Field) (entFieldType EntFieldType, err error) {
	if isIdField(field) {
		return EntFieldTypeUUID, nil
	}
	kind := getFieldKind(field)
	var ok bool
	if fieldIsRepeated(field) {
		entFieldType, ok = repeatedProtoTypeToEntTypeMap[kind]
	} else {
		entFieldType, ok = protoTypeToEntTypeMap[kind]
	}
	if !ok {
		if fieldIsTimestamp(field) {
			return EntFieldTypeTime, nil
		}
		err = errors.New(fmt.Sprintf("unsupported field type: %s", kind))
	}

	return
}

func getEntTypeString(field *protogen.Field) (string, error) {
	entType, err := getEntFieldType(field)
	if err != nil {
		return "", err
	}
	typeString, ok := entTypeToStringTypeMap[entType]
	if !ok {
		return "", errors.New(fmt.Sprintf("unsupported field type: %s", getFieldKind(field)))
	}

	return typeString, nil
}

func getEntFieldAnnotations(field *protogen.Field) ([]string, error) {
	annotations := []string{}
	opts := getFieldOptions(field)
	if opts.GraphqlType != "" {
		annotations = append(annotations, getFieldGraphqlTypeAnnotationDefinition(opts.GraphqlType))
	} else {
		// check to see if we need to override the graphql type
		gqlTypeOverride, ok := graphqlTypeOverrideMap[getFieldKind(field)]
		if ok && !fieldIsRepeated(field) {
			annotations = append(annotations, getFieldGraphqlTypeAnnotationDefinition(gqlTypeOverride))
		}
	}
	if !fieldIsRepeated(field) {
		annotations = append(annotations, fmt.Sprintf("entgql.OrderField(\"%s\")", strcase.ToScreamingSnake(getFieldProtoName(field))))
	}
	return annotations, nil
}

func getFieldGraphqlTypeAnnotationDefinition(gqlType string) string {
	return fmt.Sprintf("entgql.Type(\"%s\")", gqlType)
}

var graphqlTypeOverrideMap = map[protoreflect.Kind]string{
	protoreflect.Uint32Kind:  "Uint32",
	protoreflect.Fixed32Kind: "Uint32",
	protoreflect.Uint64Kind:  "Uint64",
	protoreflect.Fixed64Kind: "Uint64",
}

func getEntFieldComment(field *protogen.Field) string {
	return getFieldOptions(field).Comment
}

func getEntFieldDefault(field *protogen.Field) string {
	return getFieldOptions(field).Default
}

func getEntFieldDefaultFunc(field *protogen.Field) string {
	return getFieldOptions(field).DefaultFunc
}

func getEntFieldEnumValues(field *protogen.Field) []string {
	values := []string{}
	if fieldIsEnum(field) {
		for _, enumValue := range field.Enum.Values {
			value := getEnumValue(enumValue)
			values = append(values, value)
		}
		if !getFieldOptions(field).IncludeUnspecifiedEnum {
			values = lo.Filter(values, func(item string, index int) bool {
				return !enumValueIsUnspecified(item)
			})
		}
	}
	return values
}

func getEntFieldGoType(field *protogen.Field) string {
	return getFieldOptions(field).GoType
}

func getEntFieldImmutable(field *protogen.Field) bool {
	return getFieldOptions(field).Immutable
}

func getEntFieldOptional(field *protogen.Field) bool {
	return fieldIsOptional(field)
}

func getEntFieldSensitive(field *protogen.Field) bool {
	return getFieldOptions(field).Sensitive
}

func getEntFieldStorageKey(field *protogen.Field) string {
	return getFieldOptions(field).StorageKey
}

func getEntFieldUnique(field *protogen.Field) bool {
	return getFieldOptions(field).Unique
}

func getEntFieldStructTag(field *protogen.Field) string {
	return getFieldOptions(field).StructTag
}

var repeatedProtoTypeToEntTypeMap = map[protoreflect.Kind]EntFieldType{
	protoreflect.StringKind:   EntFieldTypeStrings,
	protoreflect.Int32Kind:    EntFieldTypeInts,
	protoreflect.Sint32Kind:   EntFieldTypeInts,
	protoreflect.Uint32Kind:   EntFieldTypeInts,
	protoreflect.Sfixed32Kind: EntFieldTypeInts,
	protoreflect.Fixed32Kind:  EntFieldTypeInts,
	protoreflect.Int64Kind:    EntFieldTypeInts,
	protoreflect.Sint64Kind:   EntFieldTypeInts,
	protoreflect.Uint64Kind:   EntFieldTypeInts,
	protoreflect.Sfixed64Kind: EntFieldTypeInts,
	protoreflect.Fixed64Kind:  EntFieldTypeInts,
	protoreflect.FloatKind:    EntFieldTypeFloats,
	protoreflect.DoubleKind:   EntFieldTypeFloats,
}

var protoTypeToEntTypeMap = map[protoreflect.Kind]EntFieldType{
	protoreflect.StringKind:   EntFieldTypeString,
	protoreflect.BoolKind:     EntFieldTypeBool,
	protoreflect.EnumKind:     EntFieldTypeEnum,
	protoreflect.Int32Kind:    EntFieldTypeInt32,
	protoreflect.Sint32Kind:   EntFieldTypeInt32,
	protoreflect.Uint32Kind:   EntFieldTypeUint32,
	protoreflect.Sfixed32Kind: EntFieldTypeInt32,
	protoreflect.Fixed32Kind:  EntFieldTypeUint32,
	protoreflect.Int64Kind:    EntFieldTypeInt64,
	protoreflect.Sint64Kind:   EntFieldTypeInt64,
	protoreflect.Uint64Kind:   EntFieldTypeUint64,
	protoreflect.Sfixed64Kind: EntFieldTypeInt64,
	protoreflect.Fixed64Kind:  EntFieldTypeUint64,
	protoreflect.FloatKind:    EntFieldTypeFloat,
	protoreflect.DoubleKind:   EntFieldTypeFloat,
	protoreflect.BytesKind:    EntFieldTypeBytes,
}

var entTypeToStringTypeMap = map[EntFieldType]string{
	EntFieldTypeEnum:    "Enum",
	EntFieldTypeString:  "String",
	EntFieldTypeUUID:    "UUID",
	EntFieldTypeFloat32: "Float32",
	EntFieldTypeBool:    "Bool",
	EntFieldTypeFloat:   "Float",
	EntFieldTypeTime:    "Time",
	EntFieldTypeAny:     "Any",
	EntFieldTypeBytes:   "Bytes",
	EntFieldTypeFloats:  "Floats",
	EntFieldTypeInt:     "Int",
	EntFieldTypeInt8:    "Int8",
	EntFieldTypeInt16:   "Int16",
	EntFieldTypeInt32:   "Int32",
	EntFieldTypeInt64:   "Int64",
	EntFieldTypeInts:    "Ints",
	EntFieldTypeJson:    "JSON",
	EntFieldTypeOther:   "Other",
	EntFieldTypeStrings: "Strings",
	EntFieldTypeText:    "Text",
	EntFieldTypeUint:    "Uint",
	EntFieldTypeUint8:   "Uint8",
	EntFieldTypeUint16:  "Uint16",
	EntFieldTypeUint32:  "Uint32",
	EntFieldTypeUint64:  "Uint64",
	EntFieldTypeId:      "ID",
}

func shouldHandleMessage(message *protogen.Message) bool {
	messageOptions := getMessageOptions(message)
	return messageOptions.Gen
}

func logImportant(text string, args ...any) {
	val := fmt.Sprintf(text, args...)
	glog.Infof("***********************************************************%s***********************************************************", val)
}

func getFileOptions(file *protogen.File) ent.EntFileOptions {
	emptyOptions := ent.EntFileOptions{}
	if file.Desc.Options() == nil {
		// return empty options
		return emptyOptions
	}
	options, ok := file.Desc.Options().(*descriptorpb.FileOptions)
	if !ok {
		// return empty options
		return emptyOptions
	}

	v := proto.GetExtension(options, ent.E_EntFileOpts)
	if v == nil {
		// return empty options
		return emptyOptions
	}

	opts, ok := v.(*ent.EntFileOptions)
	if !ok || opts == nil {
		// return empty options
		return emptyOptions
	}
	return *opts
}
