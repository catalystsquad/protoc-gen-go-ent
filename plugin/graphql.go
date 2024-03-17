package plugin

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

var queryFile *protogen.GeneratedFile

const indent = "  "

// initQueryFile initializes the query file if it's not initialized. The query file can be used with tools like gqlgenc
// to generate clients. The queryFile var is used when handling messages to generate queries for them
func initQueryFile(gen *protogen.Plugin) {
	if queryFile == nil {
		fileName := getAppFileName("queries.graphql")
		queryFile = gen.NewGeneratedFile(fileName, "")
	}
}

func generateQueriesAndMutations(gen *protogen.Plugin, message *protogen.Message) error {
	initQueryFile(gen)
	if shouldGenerateMutations(message) {
		err := generateCreateMutation(message)
		if err != nil {
			return err
		}
		err = generateUpdateMutation(message)
		if err != nil {
			return err
		}
		err = generateDeleteMutation(message)
		if err != nil {
			return err
		}
	}
	generateQueries(message)

	return nil
}

func generateCreateMutation(message *protogen.Message) error {
	name := getMessageProtoName(message)
	varsDefinition, err := getCreateMutationVarsDefinition(message)
	if err != nil {
		return err
	}
	queryFile.P("mutation Create", name, varsDefinition, "{")
	queryFile.P(indent, "create", name, "(input:", getCreateMutationInput(message), ")")
	queryFile.P(indent, "{", getGraphqlFieldNamesString(message, true, false), "}")
	queryFile.P("}")
	return nil
}

func generateUpdateMutation(message *protogen.Message) error {
	name := getMessageProtoName(message)
	varsDefinition, err := getUpdateMutationVarsDefinition(message)
	if err != nil {
		return err
	}
	queryFile.P("mutation Update", name, varsDefinition, "{")
	queryFile.P(indent, "update", name, "(id: $id, input:", getUpdateMutationInput(message), ")")
	queryFile.P(indent, "{", getGraphqlFieldNamesString(message, true, false), "}")
	queryFile.P("}")
	return nil
}

func generateDeleteMutation(message *protogen.Message) error {
	name := getMessageProtoName(message)
	queryFile.P("mutation Delete", name, "($id: ID!) {")
	queryFile.P(indent, "delete", name, "(id: $id)")
	queryFile.P("}")
	return nil
}

func generateQueries(message *protogen.Message) {
	if len(getNonMessageFields(message)) > 0 {
		name := getPluralMessageProtoName(message)
		queryFile.P("query ", name, "(", getQueryVars(message), ") {")
		queryFile.P(indent, strings.ToLower(strcase.ToLowerCamel(name)), "(", getQueryArgs(), ") {")
		queryFile.P(indent, indent, "edges {")
		queryFile.P(indent, indent, indent, "node {")
		queryFile.P(indent, indent, indent, indent, getGraphqlFieldNamesString(message, false, false))
		queryFile.P(indent, indent, indent, "}")
		queryFile.P(indent, indent, "}")
		queryFile.P(indent, "}")
		queryFile.P("}")
	}
}

func getQueryVars(message *protogen.Message) string {
	name := getMessageProtoName(message)
	return strings.Replace("$after: Cursor, $first:Int, $before:Cursor,$last:Int,$orderBy:[BlarfOrder!],$where:BlarfWhereInput", "Blarf", name, -1)
}

func getQueryArgs() string {
	return "after:$after,first:$first, before:$before,last:$last,orderBy:$orderBy,where:$where"
}

func getCreateMutationVarsDefinition(message *protogen.Message) (string, error) {
	builder := strings.Builder{}
	fields := getMutationFields(false, message)
	fieldVarDefinitions := []string{}
	builder.WriteString("(")
	for _, field := range fields {
		fieldVarDefinition, err := getFieldVarDefinition(field)
		if err != nil {
			return "", err
		}
		fieldVarDefinitions = append(fieldVarDefinitions, fieldVarDefinition)
	}
	builder.WriteString(strings.Join(fieldVarDefinitions, ","))
	builder.WriteString(")")

	return builder.String(), nil
}

func getUpdateMutationVarsDefinition(message *protogen.Message) (string, error) {
	builder := strings.Builder{}
	fields := getMutationFields(false, message)
	fieldVarDefinitions := []string{}
	builder.WriteString("($id:ID!, ")
	for _, field := range fields {
		fieldVarDefinition, err := getFieldVarDefinition(field)
		if err != nil {
			return "", err
		}
		fieldVarDefinitions = append(fieldVarDefinitions, fieldVarDefinition)
	}
	builder.WriteString(strings.Join(fieldVarDefinitions, ","))
	builder.WriteString(")")

	return builder.String(), nil
}

func getQueryMutationVarsDefinition(message *protogen.Message) (string, error) {
	builder := strings.Builder{}
	fields := getMutationFields(true, message)
	fieldVarDefinitions := []string{}
	builder.WriteString("(")
	for _, field := range fields {
		fieldVarDefinition, err := getFieldVarDefinition(field)
		if err != nil {
			return "", err
		}
		fieldVarDefinitions = append(fieldVarDefinitions, fieldVarDefinition)
	}
	builder.WriteString(strings.Join(fieldVarDefinitions, ","))
	builder.WriteString(")")

	return builder.String(), nil
}

func getFieldVarDefinition(field *protogen.Field) (string, error) {
	varName := getGraphqlFieldName(field)
	graphqlType, err := getGraphqlType(field)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("$%s: %s", varName, graphqlType), nil
}

func getGraphqlType(field *protogen.Field) (string, error) {
	var graphqlType string

	if fieldTypeIsMessage(field) {
		graphqlType = getFieldMessageType(field)
	} else if fieldTypeIsEnum(field) {
		graphqlType = getFieldEnumType(field)
	} else {
		kind := field.Desc.Kind()
		var ok bool
		if fieldIsRepeated(field) {
			graphqlType, ok = repeatedProtoToGraphqlTypes[kind]
		} else {
			graphqlType, ok = protoToGraphqlTypes[kind]
		}
		if !ok {
			return "", errors.New(fmt.Sprintf("unknown graphql type for proto type %s", kind))
		}
	}

	if fieldIsRepeated(field) {
		graphqlType = fmt.Sprintf("[%s!]", graphqlType)
	}
	if !fieldIsOptional(field) {
		graphqlType = fmt.Sprintf("%s!", graphqlType)
	}

	return graphqlType, nil
}

func getCreateMutationInput(message *protogen.Message) string {
	fields := getMutationFieldNames(false, message)
	builder := strings.Builder{}
	builder.WriteString("{")
	inputFields := []string{}
	for _, field := range fields {
		inputField := fmt.Sprintf("%s: $%s", field, field)
		inputFields = append(inputFields, inputField)
	}
	return fmt.Sprintf("{%s}", strings.Join(inputFields, ","))
}

func getUpdateMutationInput(message *protogen.Message) string {
	fields := getMutationFieldNames(false, message)
	builder := strings.Builder{}
	builder.WriteString("{")
	inputFields := []string{}
	for _, field := range fields {
		inputField := fmt.Sprintf("%s: $%s", field, field)
		inputFields = append(inputFields, inputField)
	}
	return fmt.Sprintf("{%s}", strings.Join(inputFields, ","))
}

func getGraphqlFieldNamesString(message *protogen.Message, includeIdField, includeMessages bool) string {
	return strings.Join(getGraphqlFieldNames(message, includeIdField, includeMessages), " ")
}
func getGraphqlFieldNames(message *protogen.Message, includeIdField, includeMessages bool) []string {
	fieldNames := []string{}
	for _, field := range getNonIgnoredFields(message) {
		if !includeMessages && fieldTypeIsMessage(field) {
			continue
		}
		if !includeIdField && getFieldProtoName(field) == "id" {
			continue
		}
		graphqlFieldName := getGraphqlFieldName(field)
		fieldNames = append(fieldNames, graphqlFieldName)
	}

	return fieldNames
}

func getMutationFieldNames(includeId bool, message *protogen.Message) []string {
	fields := getMutationFields(includeId, message)
	mutationFieldNames := []string{}
	for _, field := range fields {
		mutationFieldNames = append(mutationFieldNames, getGraphqlFieldName(field))
	}

	return mutationFieldNames
}

func getMutationFields(includeId bool, message *protogen.Message) []*protogen.Field {
	mutationFields := []*protogen.Field{}
	for _, field := range getNonIgnoredFields(message) {
		if !fieldTypeIsMessage(field) {
			if !includeId && getFieldProtoName(field) == "id" {
				continue
			}
			mutationFields = append(mutationFields, field)
		}
	}

	return mutationFields
}

var protoToGraphqlTypes = map[protoreflect.Kind]string{
	protoreflect.StringKind:   "String",
	protoreflect.BoolKind:     "Boolean",
	protoreflect.Int32Kind:    "Int",
	protoreflect.Sint32Kind:   "Int",
	protoreflect.Uint32Kind:   "Uint32",
	protoreflect.Sfixed32Kind: "Int",
	protoreflect.Fixed32Kind:  "Uint32",
	protoreflect.Int64Kind:    "Int",
	protoreflect.Sint64Kind:   "Int",
	protoreflect.Uint64Kind:   "Uint64",
	protoreflect.Sfixed64Kind: "Int",
	protoreflect.Fixed64Kind:  "Uint64",
	protoreflect.FloatKind:    "Float",
	protoreflect.DoubleKind:   "Float",
}

var repeatedProtoToGraphqlTypes = map[protoreflect.Kind]string{
	protoreflect.StringKind:   "String",
	protoreflect.Int32Kind:    "Int",
	protoreflect.Sint32Kind:   "Int",
	protoreflect.Uint32Kind:   "Int",
	protoreflect.Sfixed32Kind: "Int",
	protoreflect.Fixed32Kind:  "Int",
	protoreflect.Int64Kind:    "Int",
	protoreflect.Sint64Kind:   "Int",
	protoreflect.Uint64Kind:   "Int",
	protoreflect.Sfixed64Kind: "Int",
	protoreflect.Fixed64Kind:  "Int",
	protoreflect.FloatKind:    "Float",
	protoreflect.DoubleKind:   "Float",
}

func shouldGenerateMutations(message *protogen.Message) bool {
	return len(getNonMessageFields(message)) > 0
}
