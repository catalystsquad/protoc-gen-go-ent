package plugin

import (
	"errors"
	"fmt"
	ent "github.com/catalystsquad/protoc-gen-go-ent/options"
	"github.com/golang/glog"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

func WriteEdges(g *protogen.GeneratedFile, message *protogen.Message) error {
	messageGoName := getMessageGoName(message)
	g.P(fmt.Sprintf("func (%s) Edges() []ent.Edge {", messageGoName))
	edgeFields := getEdgeFields(message)
	if len(edgeFields) == 0 {
		g.P("return nil")
	} else {
		g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "entgo.io/ent/schema/edge"})
		g.P("return []ent.Edge{")
		for _, field := range edgeFields {
			err := writeEdge(g, field)
			if err != nil {
				return err
			}
		}
		g.P("}")
	}
	g.P(fmt.Sprintf("}"))
	return nil
}

func getEdgeFields(message *protogen.Message) []*protogen.Field {
	fields := []*protogen.Field{}
	for _, field := range message.Fields {
		if shouldIncludeFieldInEdges(field) {
			glog.Infof("getEdgeFields() field %s on message %s is an edge", getFieldProtoName(field), getFieldParentMessageType(field))
			fields = append(fields, field)
		}
	}

	return fields
}

func shouldIncludeFieldInEdges(field *protogen.Field) bool {
	if fieldIsIgnored(field) {
		return false
	}
	kind := getFieldKind(field)
	return kind == protoreflect.MessageKind
}

func writeEdge(g *protogen.GeneratedFile, field *protogen.Field) error {
	builder := &strings.Builder{}
	options := getEdgeOptions(field)
	if options.Ref != "" {
		// if there's a ref then we write a From() edge
		writeFrom(builder, field)
		writeRef(builder, field)
	} else {
		// else we write a To() edge
		writeTo(builder, field)
	}
	err := writeUnique(builder, field)
	if err != nil {
		return err
	}
	builder.WriteString(",")
	g.P(builder.String())
	return nil
}

func getEdgeOptions(field *protogen.Field) *ent.EntEdgeOptions {
	options := getFieldOptions(field)
	return options.EdgeOptions
}

func writeFrom(builder *strings.Builder, field *protogen.Field) {
	edgeName := getEdgeName(field)
	edgeType := getEdgeType(field)
	builder.WriteString("edge.From(\"")
	builder.WriteString(edgeName)
	builder.WriteString("\", ")
	builder.WriteString(edgeType)
	builder.WriteString(".Type)")
}

func getEdgeName(field *protogen.Field) string {
	fieldGoName := getFieldGoName(field)
	return strcase.ToSnake(fieldGoName)
}

func getEdgeType(field *protogen.Field) string {
	return getMessageProtoName(getFieldMessage(field))
}

func writeRef(builder *strings.Builder, field *protogen.Field) {
	ref := getEdgeRef(field)
	if ref != "" {
		builder.WriteString(".Ref(\"")
		builder.WriteString(ref)
		builder.WriteString("\")")
	}
}

func getEdgeRef(field *protogen.Field) string {
	options := getEdgeOptions(field)
	return options.Ref
}

func writeTo(builder *strings.Builder, field *protogen.Field) {
	edgeName := getEdgeName(field)
	edgeType := getEdgeType(field)
	builder.WriteString("edge.To(\"")
	builder.WriteString(edgeName)
	builder.WriteString("\", ")
	builder.WriteString(edgeType)
	builder.WriteString(".Type)")
}

func writeUnique(builder *strings.Builder, field *protogen.Field) error {
	glog.Infof("writeUnique() handling field %s on message %s", getFieldProtoName(field), getFieldParentMessageType(field))
	otherEdgeField, err := getOtherEdgeField(field)
	if err != nil {
		return err
	}
	// one to one and many to one relationships require the unique() specifier
	glog.Infof("field: %s.%s", getFieldParentMessageType(field), getFieldProtoName(field))
	glog.Infof("otherEdgeField: %s.%s", getFieldParentMessageType(otherEdgeField), getFieldProtoName(otherEdgeField))
	oneToOne := isOneToOne(field, otherEdgeField)
	oneToMany := isOneToMany(field, otherEdgeField)
	if oneToOne || oneToMany {
		builder.WriteString(".Unique()")
	}
	return nil
}

func isOneToOne(field, otherEdgeField *protogen.Field) bool {
	thisFieldIsRepeated := fieldIsRepeated(field)
	otherEdgeFieldIsRepeated := fieldIsRepeated(otherEdgeField)
	return !thisFieldIsRepeated && !otherEdgeFieldIsRepeated
}

func isOneToMany(field, otherEdgeField *protogen.Field) bool {
	thisFieldIsRepeated := fieldIsRepeated(field)
	otherEdgeFieldIsRepeated := fieldIsRepeated(otherEdgeField)
	return !thisFieldIsRepeated && otherEdgeFieldIsRepeated
}

func getOtherEdgeField(field *protogen.Field) (*protogen.Field, error) {
	otherEdgeMessage := getFieldMessage(field)
	otherEdgeType := getMessageProtoName(otherEdgeMessage)
	thisFieldName := getFieldProtoName(field)
	thisFieldType := getMessageProtoName(otherEdgeMessage)
	thisFieldParentMessage := getFieldParentMessage(field)
	thisFieldParentMessageType := getMessageProtoName(thisFieldParentMessage)
	glog.Infof("getOtherEdgeField(): this field name: %s, this field type: %s, thisFieldParentMessageType: %s, other edge type: %s", thisFieldName, thisFieldType, thisFieldParentMessageType, otherEdgeType)
	possibleFields, err := getOtherEdgeFieldsOfSameType(field)
	if err != nil {
		return nil, err
	}
	glog.Infof("getOtherEdgeField(): possible edge fields")
	for _, possibleField := range possibleFields {
		logFieldTypeAndName(possibleField)
	}
	glog.Infof("getOtherEdgeField(): end possible edge fields")
	if len(possibleFields) == 1 {
		return possibleFields[0], nil
	}
	//otherEdgeFieldsOfThisType := getOtherEdgeFieldsOfSameType(field)
	//var otherEdgeField *protogen.Field
	//// see if we need the option specified
	//if len(otherEdgeFields) == 0 {
	//	return nil, errors.New(fmt.Sprintf("field %s on type %s references type %s, but type %s has no fields of type %s", thisFieldName, thisFieldParentMessageType, otherType, otherType, thisFieldParentMessageType))
	//} else if len(otherEdgeFields) == 1 {
	//	otherEdgeField = otherEdgeFields[0]
	//} else {
	//	// multiple fields of the same type, that means a ref is required either on this field referencing  a field on
	//	// the other edge, or on the other edge referencing this field
	//	otherEdgeField = getOtherEdgeFieldReferencedByThisField(field)
	//	// if the other edge is nil, then this field doesn't have a ref to the other edge so we need to check the other
	//	// edge to see if it has a ref to this field
	//	if otherEdgeField == nil {
	//		otherEdgeField = getOtherEdgeReferencingField(field)
	//	}
	//}
	//if otherEdgeField == nil {
	//	return nil, errors.New(fmt.Sprintf(fmt.Sprintf("unable to determine edge, field %s on type %s references type %s, but type %s has more than one field of type %s and no ref is specified, when a message has more than one field of the same message type, a ref must be specified on one of the edges because it can't be inferred", thisFieldName, thisFieldParentMessageType, otherType, otherType, thisFieldParentMessageType)))
	//}
	//return otherEdgeField, nil

	return nil, nil
}

func getOtherEdgeFieldsOfSameType(field *protogen.Field) ([]*protogen.Field, error) {
	otherEdgeMessage := getFieldMessage(field)
	otherEdgeType := getMessageProtoName(otherEdgeMessage)
	thisFieldType := getFieldParentMessageType(field)
	thisFieldParentMessage := getFieldParentMessage(field)
	thisFieldParentMessageType := getMessageProtoName(thisFieldParentMessage)
	glog.Infof("getOtherEdgeFieldsOfSameType(): this field type: %s otherEdge type: %s", thisFieldType, otherEdgeType)
	fieldsOfThisType := []*protogen.Field{}
	for _, otherEdgeMessageField := range otherEdgeMessage.Fields {
		if !fieldTypeIsMessage(otherEdgeMessageField) {
			continue
		}
		otherEdgeFieldType := getMessageProtoName(getFieldMessage(otherEdgeMessageField))
		glog.Infof(
			"%s field %s is of type %s, thisField parent is of type %s",
			getMessageProtoName(otherEdgeMessage),
			getFieldProtoName(otherEdgeMessageField),
			getMessageProtoName(getFieldMessage(otherEdgeMessageField)),
			thisFieldParentMessageType,
		)
		if otherEdgeFieldType == thisFieldType {
			glog.Infof("getOtherEdgeFieldsOfSameType() matched field")
			fieldsOfThisType = append(fieldsOfThisType, otherEdgeMessageField)
		}
	}

	if len(fieldsOfThisType) == 0 {
		return nil, errors.New(fmt.Sprintf(
			"Failed to create edge for field %s.%s, type %s has no fields of type %s",
			getFieldProtoName(field),
			getFieldParentMessageType(field),
			getMessageProtoName(otherEdgeMessage),
			getFieldParentMessageType(field),
		))
	}
	return fieldsOfThisType, nil
}

func logFieldTypeAndName(field *protogen.Field) {
	glog.Infof("field: %s.%s", getFieldParentMessageType(field), getFieldProtoName(field))
}
