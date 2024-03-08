package plugin

//
//import (
//	"errors"
//	"fmt"
//	ent "github.com/catalystsquad/protoc-gen-go-ent/options"
//	"github.com/golang/glog"
//	"github.com/iancoleman/strcase"
//	"google.golang.org/protobuf/compiler/protogen"
//	"google.golang.org/protobuf/proto"
//	"google.golang.org/protobuf/reflect/protoreflect"
//	"google.golang.org/protobuf/types/descriptorpb"
//)
//
//type EdgeType int
//
//func (e EdgeType) String() string {
//	return []string{"Unknown", "OneToOne", "OneToMany", "ManyToOne", "ManyToMany"}[e]
//}
//
//const (
//	Unknown EdgeType = iota
//	OneToOneTwoTypes
//	OneToOneSameType
//	OneToOneBidirectional
//	OneToManyTwoTypes
//	OneToManySameType
//	ManyToOneTwoTypes
//	ManyToOneSameType
//	ManyToManyTwoTypes
//	ManyToManySameType
//	ManyToManyBidirectional
//)
//
//func ParseProtogenMessage(file *protogen.File, protogenMessage *protogen.Message) (*ParsedMessage, error) {
//	fields, edges, err := parseProtogenFields(protogenMessage.Fields)
//	if err != nil {
//		return nil, err
//	}
//	return &ParsedMessage{
//		FileName:    getProtogenMessageFileName(file, protogenMessage),
//		PackageName: getProtogenMessagePackageName(file),
//		ImportPath:  getMessageImportPath(file),
//		StructName:  getProtogenMessageStructName(protogenMessage),
//		Options:     getMessageOptions(protogenMessage),
//		Fields:      fields,
//		Edges:       edges,
//	}, nil
//}
//
//func getProtogenMessageFileName(file *protogen.File, protogenMessage *protogen.Message) string {
//	return fmt.Sprintf("%s_%s_ent.pb.go", file.GeneratedFilenamePrefix, strcase.ToSnake(protogenMessage.GoIdent.GoName))
//}
//
//func getProtogenMessagePackageName(file *protogen.File) string {
//	return string(file.GoPackageName)
//}
//
//func getMessageImportPath(file *protogen.File) protogen.GoImportPath {
//	return file.GoImportPath
//}
//
//func getProtogenMessageStructName(protogenMessage *protogen.Message) string {
//	return protogenMessage.GoIdent.GoName
//}
//
//func parseProtogenFields(protogenFields []*protogen.Field) ([]*Field, []*Edge, error) {
//	fields := buildFieldsFromProtogenFields(protogenFields)
//	edges, err := buildEdgesFromProtogenFields(protogenFields)
//	if err != nil {
//		return nil, nil, err
//	}
//	return fields, edges, nil
//}
//
//func buildFieldsFromProtogenFields(protogenFields []*protogen.Field) []*Field {
//	fields := []*Field{}
//	for _, protogenField := range protogenFields {
//		field := buildFieldFromProtogenField(protogenField)
//		if shouldGenerateField(field) {
//			fields = append(fields, field)
//		}
//	}
//
//	return fields
//}
//
//func buildFieldFromProtogenField(protogenField *protogen.Field) *Field {
//	fieldOptions := getFieldOptions(protogenField)
//	field := Field{
//		Name:      getFieldNameFromProtogenField(protogenField),
//		EntType:   getEntTypeFromProtogenField(protogenField),
//		Optional:  fieldIsOptional(protogenField),
//		Options:   fieldOptions,
//		ProtoKind: protogenField.Desc.Kind(),
//		IsEnum:    fieldIsEnum(protogenField),
//	}
//
//	if field.IsEnum {
//		field.EnumValues = getFieldEnumValues(protogenField)
//	}
//
//	return &field
//}
//
//func getFieldNameFromProtogenField(protogenField *protogen.Field) string {
//	fieldName := strcase.ToSnake(protogenField.GoName)
//	//glog.Infof("getFieldNameFromProtogenField(): message field name: %s, parsed field name: %s", protogenField.GoName, fieldName)
//	return fieldName
//}
//
//func getEntTypeFromProtogenField(protogenField *protogen.Field) string {
//	return protoFieldTypeToEntFieldTypeMap[protogenField.Desc.Kind()]
//}
//
//func fieldIsEnum(protogenField *protogen.Field) bool {
//	return protogenField.Desc.Kind() == protoreflect.EnumKind
//}
//
//func getFieldEnumValues(protogenField *protogen.Field) []string {
//	values := []string{}
//	for _, value := range protogenField.Enum.Values {
//		values = append(values, string(value.Desc.Name()))
//	}
//	return values
//}
//
//func buildEdgesFromProtogenFields(protogenFields []*protogen.Field) ([]*Edge, error) {
//	edges := []*Edge{}
//	for _, protogenField := range protogenFields {
//		field := buildFieldFromProtogenField(protogenField)
//		if shouldIncludeFieldInEdges(field) {
//			edge, err := buildEdgeFromProtogenField(protogenField)
//			if err != nil {
//				return nil, err
//			}
//			edges = append(edges, edge)
//		}
//	}
//
//	return edges, nil
//}
//
//func buildEdgeFromProtogenField(protogenField *protogen.Field) (*Edge, error) {
//	edgeOptions := getEdgeOptions(protogenField)
//	edgeType, err := getEdgeType(protogenField)
//	if err != nil {
//		return nil, err
//	}
//	ref, err := getEdgeRef(protogenField, edgeOptions)
//	if err != nil {
//		return nil, err
//	}
//	return &Edge{
//		To:         getEdgeTo(protogenField, edgeOptions),
//		ToType:     getEdgeToType(protogenField),
//		From:       getEdgeFrom(protogenField, edgeOptions),
//		FromType:   getEdgeFromType(protogenField),
//		Ref:        ref,
//		TypeName:   getEdgeTypeName(protogenField),
//		IsSameType: edgeIsSameType(protogenField),
//		IsSingular: edgeIsSingular(protogenField),
//		Options:    getEdgeOptions(protogenField),
//		EdgeType:   edgeType,
//	}, nil
//}
//
//func getEdgeTo(protogenField *protogen.Field, options *ent.EntEdgeOptions) string {
//	if options.To != "" {
//		return options.To
//	}
//	return strcase.ToSnake(protogenField.GoName)
//}
//
//func getEdgeToType(protogenField *protogen.Field) string {
//	return protogenField.Message.GoIdent.GoName
//}
//
//func getEdgeFrom(protogenField *protogen.Field, options *ent.EntEdgeOptions) string {
//	if options.From != "" {
//		return options.From
//	}
//	return strcase.ToSnake(string(protogenField.Desc.Name()))
//}
//
//func getEdgeFromType(protogenField *protogen.Field) string {
//	return protogenField.Message.GoIdent.GoName
//}
//
//func getEdgeRef(protogenField *protogen.Field, options *ent.EntEdgeOptions) (string, error) {
//	if options.Ref != "" {
//		return options.Ref, nil
//	}
//	otherEdgeField, err := getOtherEdgeField(protogenField)
//	if err != nil {
//		return "", err
//	}
//	return strcase.ToSnake(getProtoFieldName(otherEdgeField)), nil
//}
//
//func getEdgeTypeName(protogenField *protogen.Field) string {
//	return protogenField.Parent.GoIdent.GoName
//}
//
//func edgeIsSameType(protogenField *protogen.Field) bool {
//	toType := getEdgeToType(protogenField)
//	fromType := getEdgeFromType(protogenField)
//	isSameType := toType == fromType
//	glog.Infof("edge is same type: toType: %s, fromType: %s, result: %t", toType, fromType, isSameType)
//	return isSameType
//}
//
//func edgeIsSingular(protogenField *protogen.Field) bool {
//	isSinguler := !protogenField.Desc.IsList()
//	glog.Infof("edge: %s is singular: %t", protogenField.GoName, isSinguler)
//	return isSinguler
//}
//
//type ParsedMessage struct {
//	FileName    string
//	ImportPath  protogen.GoImportPath
//	PackageName string
//	StructName  string
//	Fields      []*Field
//	Edges       []*Edge
//	Options     ent.EntMessageOptions
//}
//
//type Field struct {
//	Name       string
//	EntType    string
//	Optional   bool
//	Options    ent.EntFieldOptions
//	ProtoKind  protoreflect.Kind
//	IsEnum     bool
//	EnumValues []string
//}
//
//type Edge struct {
//	To         string
//	ToType     string
//	From       string
//	FromType   string
//	Ref        string
//	TypeName   string
//	IsSameType bool
//	IsSingular bool
//	EdgeType   EdgeType
//	Options    *ent.EntEdgeOptions
//}
//
//func shouldIncludeFieldInEdges(field *Field) bool {
//	return field.ProtoKind == protoreflect.MessageKind
//}
//
//func getEdgeOptions(field *protogen.Field) *ent.EntEdgeOptions {
//	fieldOptions := getFieldOptions(field)
//	if fieldOptions.EdgeOptions == nil {
//		return &ent.EntEdgeOptions{}
//	}
//	return fieldOptions.EdgeOptions
//}
//
//func getEdgeType(field *protogen.Field) (EdgeType, error) {
//	otherEdgeField, err := getOtherEdgeField(field)
//	if err != nil {
//		return 0, err
//	}
//	thisFieldIsRepeated := protoFieldIsRepeated(field)
//	otherEdgeIsRepeated := protoFieldIsRepeated(otherEdgeField)
//	edgeType := Unknown
//	sameType := fieldsAreSameType(field, otherEdgeField)
//	hasRef := oneFieldHasRef(field, otherEdgeField)
//	if !thisFieldIsRepeated && !otherEdgeIsRepeated {
//		if sameType {
//			if hasRef {
//				edgeType = OneToOneSameType
//			} else {
//				edgeType = OneToOneBidirectional
//			}
//		} else {
//			edgeType = OneToOneTwoTypes
//		}
//	} else if !thisFieldIsRepeated && otherEdgeIsRepeated {
//		if sameType {
//			edgeType = OneToManySameType
//		} else {
//			edgeType = OneToManyTwoTypes
//		}
//	} else if thisFieldIsRepeated && !otherEdgeIsRepeated {
//		if sameType {
//			edgeType = ManyToOneSameType
//		} else {
//			edgeType = ManyToOneTwoTypes
//		}
//	} else {
//		if sameType {
//			if hasRef {
//				edgeType = ManyToManySameType
//			} else {
//				edgeType = ManyToManyBidirectional
//			}
//		} else {
//			edgeType = ManyToManyTwoTypes
//		}
//	}
//
//	glog.Infof(fmt.Sprintf("%s to %s is edge type %s", field.GoIdent.GoName, otherEdgeField.GoIdent.GoName, edgeType))
//	if edgeType == Unknown {
//		return Unknown, errors.New(fmt.Sprintf("unable to determine type of edge between %s and %s", field.GoIdent.GoName, otherEdgeField.GoIdent.GoName))
//	}
//	return edgeType, nil
//}
//
//func oneFieldHasRef(a, b *protogen.Field) bool {
//	if fieldHasRef(a) {
//		return true
//	}
//	if fieldHasRef(b) {
//		return true
//	}
//	return false
//}
//
//func fieldHasRef(field *protogen.Field) bool {
//	return getFieldRef(field) != ""
//}
//
//func getFieldRef(field *protogen.Field) string {
//	return getFieldOptions(field).EdgeOptions.Ref
//}
//
//func fieldsAreSameType(a, b *protogen.Field) bool {
//	return getProtoFieldType(a) == getProtoFieldType(b)
//}
//
//func getOtherEdgeMessage(field *protogen.Field) *protogen.Message {
//	return field.Message
//}
//
//func getOtherEdgeField(field *protogen.Field) (*protogen.Field, error) {
//	otherEdgeMessage := getOtherEdgeMessage(field)
//	thisFieldName := getProtoFieldName(field)
//	thisFieldType := getProtoFieldType(field)
//	thisFieldParentMessageType := getProtoFieldMessageType(field)
//	otherType := getProtoMessageType(otherEdgeMessage)
//	glog.Infof("getOtherEdgeField(): this field name: %s, this field type: %s, thisFieldParentMessageType: %s, other edge type: %s", thisFieldName, thisFieldType, thisFieldParentMessageType, otherType)
//	otherEdgeFields := getOtherEdgeFieldsOfSameType(field)
//	var otherEdgeField *protogen.Field
//	// see if we need the option specified
//	if len(otherEdgeFields) == 0 {
//		return nil, errors.New(fmt.Sprintf("field %s on type %s references type %s, but type %s has no fields of type %s", thisFieldName, thisFieldParentMessageType, otherType, otherType, thisFieldParentMessageType))
//	} else if len(otherEdgeFields) == 1 {
//		otherEdgeField = otherEdgeFields[0]
//	} else {
//		// multiple fields of the same type, that means a ref is required either on this field referencing  a field on
//		// the other edge, or on the other edge referencing this field
//		otherEdgeField = getOtherEdgeFieldReferencedByThisField(field)
//		// if the other edge is nil, then this field doesn't have a ref to the other edge so we need to check the other
//		// edge to see if it has a ref to this field
//		if otherEdgeField == nil {
//			otherEdgeField = getOtherEdgeReferencingField(field)
//		}
//	}
//	if otherEdgeField == nil {
//		return nil, errors.New(fmt.Sprintf(fmt.Sprintf("unable to determine edge, field %s on type %s references type %s, but type %s has more than one field of type %s and no ref is specified, when a message has more than one field of the same message type, a ref must be specified on one of the edges because it can't be inferred", thisFieldName, thisFieldParentMessageType, otherType, otherType, thisFieldParentMessageType)))
//	}
//	return otherEdgeField, nil
//}
//
//func getOtherEdgeFieldReferencedByThisField(field *protogen.Field) *protogen.Field {
//	options := getFieldOptions(field)
//	if options.EdgeOptions.Ref == "" {
//		return nil
//	}
//	ref := options.EdgeOptions.Ref
//	otherEdgeFields := getOtherEdgeFieldsOfSameType(field)
//	for _, otherEdgeField := range otherEdgeFields {
//		otherFieldName := getProtoFieldName(otherEdgeField)
//		if ref == otherFieldName {
//			return otherEdgeField
//		}
//	}
//	return nil
//}
//
//func getOtherEdgeReferencingField(field *protogen.Field) *protogen.Field {
//	thisFieldName := getProtoFieldName(field)
//	otherEdgeFields := getOtherEdgeFieldsOfSameType(field)
//	for _, otherEdgeField := range otherEdgeFields {
//		otherEdgeOptions := getFieldOptions(otherEdgeField)
//		glog.Infof("getOtherEdgeReferencingField() field name: %s, other edge field name %s ref %s", thisFieldName, getProtoFieldName(otherEdgeField), otherEdgeOptions.EdgeOptions.Ref)
//		if otherEdgeOptions.EdgeOptions.Ref == thisFieldName {
//			return otherEdgeField
//		}
//	}
//	return nil
//}
//
//func getOtherEdgeFieldsOfSameType(field *protogen.Field) []*protogen.Field {
//	otherEdgeMessage := getOtherEdgeMessage(field)
//	thisFieldParentMessageType := getProtoFieldMessageType(field)
//	fieldsOfThisType := []*protogen.Field{}
//	for _, otherEdgeMessageField := range otherEdgeMessage.Fields {
//		otherEdgeFieldIsMessage := protoFieldIsMessage(otherEdgeMessageField)
//		//otherEdgeFieldName := getProtoFieldName(otherEdgeMessageField)
//		if !otherEdgeFieldIsMessage {
//			//glog.Infof("%s field %s is not message type, skipping", otherType, otherEdgeFieldName)
//			continue
//		}
//		otherEdgeMessageType := getProtoFieldType(otherEdgeMessageField)
//		//glog.Infof("%s field %s is of type %s, thisField parent is of type %s", otherType, otherEdgeFieldName, otherEdgeMessageType, thisFieldParentMessageType)
//		if otherEdgeMessageType == thisFieldParentMessageType {
//			fieldsOfThisType = append(fieldsOfThisType, otherEdgeMessageField)
//		}
//	}
//
//	return fieldsOfThisType
//}
//
////func getOtherEdgeField(field *protogen.Field) *protogen.Field {
////	otherEdgeMessage := getOtherEdgeMessage(field)
////	thisFieldName := getProtoFieldName(field)
////	thisFieldType := getProtoFieldType(field)
////	thisFieldParentMessageType := getProtoFieldMessageType(field)
////	otherType := getProtoMessageType(otherEdgeMessage)
////	glog.Infof("getOtherEdgeField(): this field name: %s, this field type: %s, thisFieldParentMessageType: %s, other edge type: %s", thisFieldName, thisFieldType, thisFieldParentMessageType, otherType)
////	fieldsOfThisType := []*protogen.Field{}
////	for _, otherEdgeMessageField := range otherEdgeMessage.Fields {
////		otherEdgeFieldIsMessage := protoFieldIsMessage(otherEdgeMessageField)
////		//otherEdgeFieldName := getProtoFieldName(otherEdgeMessageField)
////		if !otherEdgeFieldIsMessage {
////			//glog.Infof("%s field %s is not message type, skipping", otherType, otherEdgeFieldName)
////			continue
////		}
////		otherEdgeMessageType := getProtoFieldType(otherEdgeMessageField)
////		//glog.Infof("%s field %s is of type %s, thisField parent is of type %s", otherType, otherEdgeFieldName, otherEdgeMessageType, thisFieldParentMessageType)
////		if otherEdgeMessageType == thisFieldParentMessageType {
////			fieldsOfThisType = append(fieldsOfThisType, otherEdgeMessageField)
////		}
////	}
////	// see if we need the option specified
////	if len(fieldsOfThisType) == 0 {
////		panic(fmt.Sprintf("field %s on type %s references type %s, but type %s has no fields of type %s", thisFieldName, thisFieldParentMessageType, otherType, otherType, thisFieldParentMessageType))
////	} else if len(fieldsOfThisType) == 1 {
////		return fieldsOfThisType[0]
////	} else {
////		// multiple fields of the same type, that means a ref is required
////		options := getFieldOptions(field)
////		ref := options.EdgeOptions.Ref
////		if options.EdgeOptions.Ref == "" {
////			panic(fmt.Sprintf("field %s on type %s references type %s, but type %s has two fields of type %s and no ref specified, when a message has more than one field of the same message type, a ref must be specified because it can't be inferred", thisFieldName, thisFieldParentMessageType, otherType, otherType, thisFieldParentMessageType))
////		}
////		for _, field := range fieldsOfThisType {
////			fieldName := getProtoFieldName(field)
////			if ref == fieldName {
////				return field
////			}
////		}
////		panic(fmt.Sprintf("field %s on type %s references type %s but type %s does not have a field matching the given ref name %s", thisFieldName, thisFieldParentMessageType, otherType, otherType, ref))
////	}
////
////}
//
//func protoFieldIsMessage(field *protogen.Field) bool {
//	return field.Message != nil
//}
//
//func protoFieldIsRepeated(field *protogen.Field) bool {
//	return field.Desc.IsList()
//}
//
//func getProtoMessageType(message *protogen.Message) string {
//	return string(message.Desc.Name())
//}
//
//func getProtoFieldMessageType(field *protogen.Field) string {
//	return getProtoMessageType(field.Parent)
//}
//
//func getProtoFieldName(field *protogen.Field) string {
//	return string(field.Desc.Name())
//}
//
//func getProtoFieldType(field *protogen.Field) string {
//	return string(field.Message.Desc.Name())
//}
