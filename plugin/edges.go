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

type EdgeType int

const (
	UnknownEdgeType EdgeType = iota
	SameType
	TwoTypes
	Bidirectional
)

func (t EdgeType) String() string {
	return []string{"UnknownEdgeType", "SameType", "TwoTypes", "Bidirectional"}[t]
}

type EdgeCardinality int

const (
	UnknownEdgeCardinality EdgeCardinality = iota
	OneToOne
	OneToMany
	ManyToOne
	ManyToMany
)

func (c EdgeCardinality) String() string {
	return []string{"UnknownEdgeCardinality", "OneToOne", "OneToMany", "ManyToOne", "ManyToMany"}[c]
}

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

func writeEdge(g *protogen.GeneratedFile, edge *protogen.Field) error {
	builder := &strings.Builder{}
	writeEdgeBase(builder, edge)
	err := writeEdgeUnique(builder, edge)
	if err != nil {
		return err
	}
	g.P(builder.String(), ",")
	return nil
}

func writeEdgeBase(builder *strings.Builder, edge *protogen.Field) {
	if edgeHasRef(edge) && !edgeIsBidirectional(edge) {
		// if this edge has a ref specified then it's either a From() edge or a bidirectional edge
		writeFrom(builder, edge)
		writeRef(builder, edge)
	} else {
		writeTo(builder, edge)
	}
}

func writeEdgeUnique(builder *strings.Builder, edge *protogen.Field) error {
	cardinality, err := getEdgeCardinality(edge)
	if err != nil {
		return err
	}
	if cardinality == OneToOne || cardinality == ManyToOne {
		builder.WriteString(".Unique()")
	}

	return nil
}

func edgeHasRef(edge *protogen.Field) bool {
	return getEdgeRef(edge) != ""
}

func edgeIsBidirectional(edge *protogen.Field) bool {
	// an edge is bidirectional when the field is referencing itself and the field and parent types are equal
	return getEdgeRef(edge) == getFieldProtoName(edge) && getFieldParentMessageType(edge) == getFieldMessageType(edge)
}

func edgeIsReferenced(edge *protogen.Field) bool {
	fieldReferencingEdge := getFieldReferencingEdge(edge)
	return fieldReferencingEdge != nil
}

func writeEdgeBasebak(builder *strings.Builder, edge *protogen.Field) error {
	cardinality, err := getEdgeCardinality(edge)
	glog.Infof("edge cardinality: %s", cardinality)
	if err != nil {
		return err
	}
	edgeType, err := getEdgeEdgeType(edge)
	glog.Infof("edge type: %s", edgeType)
	if err != nil {
		return err
	}
	switch cardinality {
	case OneToOne:
		switch edgeType {
		case SameType:
			err = writeOneToOneSameType(builder, edge)
			if err != nil {
				return err
			}
		case TwoTypes:
			err = writeOneToOneTwoTypes(builder, edge)
			if err != nil {
				return err
			}
		case Bidirectional:
			err = writeOneToOneBidirectional(builder, edge)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown edge type")
		}
	case OneToMany:
		switch edgeType {
		case SameType:
			err = writeOneToManySameType(builder, edge)
			if err != nil {
				return err
			}
		case TwoTypes:
			err = writeOneToManyTwoTypes(builder, edge)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown edge type")
		}
	case ManyToOne:
		switch edgeType {
		case SameType:
			err = writeManyToOneSameType(builder, edge)
			if err != nil {
				return err
			}
		case TwoTypes:
			err = writeManyToOneTwoTypes(builder, edge)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown edge type")
		}
	case ManyToMany:
		switch edgeType {
		case SameType:
			err = writeManyToManySameType(builder, edge)
			if err != nil {
				return err
			}
		case TwoTypes:
			err = writeManyToManyTwoTypes(builder, edge)
			if err != nil {
				return err
			}
		case Bidirectional:
			err = writeManyToManyBidirectional(builder, edge)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown edge type")
		}
	default:
		return errors.New("unkown edge cardinality")
	}

	return nil
}

func writeOneToOneSameType(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeOneToOneTwoTypes(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeOneToOneBidirectional(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeOneToManySameType(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeOneToManyTwoTypes(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeManyToOneSameType(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeManyToOneTwoTypes(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeManyToManySameType(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeManyToManyTwoTypes(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func writeManyToManyBidirectional(builder *strings.Builder, edge *protogen.Field) error {
	return nil
}

func getEdgeOptions(field *protogen.Field) *ent.EntEdgeOptions {
	options := getFieldOptions(field)
	return options.EdgeOptions
}

func writeFrom(builder *strings.Builder, field *protogen.Field) {
	builder.WriteString("edge.From(\"")
	writeEdgeNameAndType(builder, field)
	builder.WriteString(".Type)")
}

func writeRef(builder *strings.Builder, field *protogen.Field) {
	ref := getEdgeRef(field)
	builder.WriteString(".Ref(\"")
	builder.WriteString(ref)
	builder.WriteString("\")")
}

func writeTo(builder *strings.Builder, field *protogen.Field) {
	builder.WriteString("edge.To(\"")
	writeEdgeNameAndType(builder, field)
	builder.WriteString(".Type)")
}

func writeEdgeNameAndType(builder *strings.Builder, field *protogen.Field) {
	edgeName := getEdgeName(field)
	edgeType := getEdgeType(field)
	builder.WriteString(edgeName)
	builder.WriteString("\", ")
	builder.WriteString(edgeType)
}

func getEdgeName(field *protogen.Field) string {
	fieldGoName := getFieldGoName(field)
	return strcase.ToSnake(fieldGoName)
}

func getEdgeType(field *protogen.Field) string {
	return getMessageProtoName(getFieldMessage(field))
}

func getEdgeRef(field *protogen.Field) string {
	options := getEdgeOptions(field)
	return options.Ref
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
	//oneToOne := isOneToOne(field, otherEdgeField)
	//oneToMany := isOneToMany(field, otherEdgeField)
	//if oneToOne || oneToMany {
	//	builder.WriteString(".Unique()")
	//}
	return nil
}

//func isOneToOne(field, otherEdgeField *protogen.Field) bool {
//	thisFieldIsRepeated := fieldIsRepeated(field)
//	otherEdgeFieldIsRepeated := fieldIsRepeated(otherEdgeField)
//	return !thisFieldIsRepeated && !otherEdgeFieldIsRepeated
//}

//func isOneToMany(field, otherEdgeField *protogen.Field) bool {
//	thisFieldIsRepeated := fieldIsRepeated(field)
//	otherEdgeFieldIsRepeated := fieldIsRepeated(otherEdgeField)
//	return !thisFieldIsRepeated && otherEdgeFieldIsRepeated
//}

func getFieldReferencingEdge(edge *protogen.Field) (fieldReferencingEdge *protogen.Field) {
	edgeName := getFieldProtoName(edge)
	fieldMessage := getFieldMessage(edge)
	for _, field := range fieldMessage.Fields {
		ref := getEdgeRef(field)
		if ref == edgeName {
			fieldReferencingEdge = field
		}
	}

	return
}

func getFieldReferencedByEdge(edge *protogen.Field) (fieldReferencedByEdge *protogen.Field) {
	ref := getEdgeRef(edge)
	otherEdgeMessage := getFieldMessage(edge)
	for _, field := range otherEdgeMessage.Fields {
		if ref == getFieldProtoName(field) {
			fieldReferencedByEdge = field
		}
	}

	return
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

func getEdgeMessages(edge *protogen.Field) (edgeMessage, fieldMessage *protogen.Message) {
	edgeMessage = getEdgeMessage(edge)
	fieldMessage = getFieldMessage(edge)

	return
}

func getEdgeMessage(edge *protogen.Field) *protogen.Message {
	return getFieldParentMessage(edge)
}

func getOtherEdgeByRef(edge *protogen.Field) (otherEdge *protogen.Field, err error) {
	otherEdge = getFieldReferencedByEdge(edge)
	if otherEdge == nil {
		otherEdge = getFieldReferencingEdge(edge)
	}
	if otherEdge == nil {
		err = errors.New(fmt.Sprintf("unable to find edge referencing, or referenced by field %s", getQualifiedProtoFieldName(edge)))
	}
	return
}

func isOneToOne(edge *protogen.Field) (bool, error) {
	otherEdge, err := getOtherEdgeByRef(edge)
	if err != nil {
		return false, err
	}
	return !fieldIsRepeated(edge) && !fieldIsRepeated(otherEdge), nil

}

func isOneToMany(edge *protogen.Field) (bool, error) {
	otherEdge, err := getOtherEdgeByRef(edge)
	if err != nil {
		return false, err
	}
	return fieldIsRepeated(edge) && !fieldIsRepeated(otherEdge), nil
}

func isManyToOne(edge *protogen.Field) (bool, error) {
	otherEdge, err := getOtherEdgeByRef(edge)
	if err != nil {
		return false, err
	}
	return !fieldIsRepeated(edge) && fieldIsRepeated(otherEdge), nil
}

func isManyToMany(edge *protogen.Field) (bool, error) {
	otherEdge, err := getOtherEdgeByRef(edge)
	if err != nil {
		return false, err
	}
	return fieldIsRepeated(edge) && fieldIsRepeated(otherEdge), nil
}

func isBidirectional(edge *protogen.Field) bool {
	a, b := getEdgeMessages(edge)
	aHasOneFieldOfTypeB := messageHasFieldsOfOtherMessage(1, a, b)
	return aHasOneFieldOfTypeB
}

func isSameType(edge *protogen.Field) bool {
	a, b := getEdgeMessages(edge)
	return messagesAreSameType(a, b)
}

func messagesAreSameType(a, b *protogen.Message) bool {
	return getMessageProtoName(a) == getMessageProtoName(b)
}

func messageHasFieldsOfOtherMessage(num int, a, b *protogen.Message) bool {
	count := 0
	bType := getMessageProtoName(b)
	for _, field := range a.Fields {
		if fieldTypeIsMessage(field) && getMessageProtoName(getFieldMessage(field)) == bType {
			count++
		}
	}

	return count == num
}

func messageHasRepeatedFieldsOfOtherMessage(num int, a, b *protogen.Message) bool {
	count := 0
	bType := getMessageProtoName(b)
	for _, field := range a.Fields {
		if fieldTypeIsMessage(field) && fieldIsRepeated(field) && getMessageProtoName(getFieldMessage(field)) == bType {
			count++
		}
	}

	return count == num
}

func getEdgeEdgeType(edge *protogen.Field) (edgeType EdgeType, err error) {
	if isSameType(edge) {
		if isBidirectional(edge) {
			edgeType = Bidirectional
		} else {
			edgeType = SameType
		}
	} else {
		edgeType = TwoTypes
	}
	if edgeType == UnknownEdgeType {
		err = errors.New(fmt.Sprintf("unknown edge type for field: %s.%s", getFieldParentMessageType(edge), getFieldName(edge)))
	}

	return
}

func getEdgeCardinality(edge *protogen.Field) (EdgeCardinality, error) {
	oneToOne, err := isOneToOne(edge)
	if err != nil {
		return UnknownEdgeCardinality, err
	}
	if oneToOne {
		return OneToOne, nil
	}
	oneToMany, err := isOneToMany(edge)
	if err != nil {
		return UnknownEdgeCardinality, err
	}
	if oneToMany {
		return OneToMany, nil
	}
	manyToOne, err := isManyToOne(edge)
	if err != nil {
		return UnknownEdgeCardinality, err
	}
	if manyToOne {
		return ManyToOne, nil
	}
	manyToMany, err := isManyToMany(edge)
	if err != nil {
		return UnknownEdgeCardinality, err
	}
	if manyToMany {
		return ManyToMany, nil
	}

	return UnknownEdgeCardinality, errors.New(fmt.Sprintf("unknown cardinality for field: %s", getQualifiedProtoFieldName(edge)))
}

// getSpecifiedEdgeCardinality checks the edge fields for a ref option and determines the cardinality based on the ref
// if no ref is found then the cardinality is not specified
//func getSpecifiedEdgeCardinality(edge *protogen.Field) (EdgeCardinality, error) {
//	edgeRef := getEdgeRef(edge)
//	if edgeRef != "" {
//		// this edge has a ref to another edge, get the other edge field
//		fieldMessage := getFieldMessage(edge)
//		matchedField := getMessageField(edgeRef, fieldMessage)
//		if matchedField == nil {
//			return UnknownEdgeCardinality, errors.New(fmt.Sprintf("field: %s specifies ref: %s but type %s has no field named %s", getQualifiedProtoFieldName(edge), edgeRef, getMessageProtoName(fieldMessage), edgeRef))
//		}
//	} else {
//	}
//}

func getCardinalityOfFields(a, b *protogen.Field) (cardinality EdgeCardinality, err error) {
	cardinality = UnknownEdgeCardinality
	if fieldsAreOneToOne(a, b) {
		cardinality = OneToOne
	} else if fieldsAreOneToMany(a, b) {
		cardinality = OneToMany
	} else if fieldsAreManyToOne(a, b) {
		cardinality = ManyToOne
	} else if fieldsAreManyToMany(a, b) {
		cardinality = OneToOne
	} else {
		err = errors.New(fmt.Sprintf("unable to determine cardinality between fields %s and %s", getQualifiedProtoFieldName(a), getQualifiedProtoFieldName(b)))
	}

	return
}

func fieldsAreOneToOne(a, b *protogen.Field) bool {
	return !fieldIsRepeated(a) && !fieldIsRepeated(b)
}

func fieldsAreOneToMany(a, b *protogen.Field) bool {
	return !fieldIsRepeated(a) && fieldIsRepeated(b)
}

func fieldsAreManyToOne(a, b *protogen.Field) bool {
	return fieldIsRepeated(a) && !fieldIsRepeated(b)
}

func fieldsAreManyToMany(a, b *protogen.Field) bool {
	return fieldIsRepeated(a) && fieldIsRepeated(b)
}
