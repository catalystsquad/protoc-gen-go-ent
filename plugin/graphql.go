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
		queryFile = gen.NewGeneratedFile("queries.graphql", "")
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
	queryFile.P(indent, "{", getGraphqlFieldNamesString(message, false), "}")
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
	queryFile.P(indent, "{", getGraphqlFieldNamesString(message, false), "}")
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
		name := getMessageProtoName(message)
		pluralName := name + "s"
		queryFile.P("query ", pluralName, "(", getQueryVars(message), ") {")
		queryFile.P(indent, strings.ToLower(strcase.ToLowerCamel(pluralName)), "(", getQueryArgs(), ") {")
		queryFile.P(indent, indent, "edges {")
		queryFile.P(indent, indent, indent, "node {")
		queryFile.P(indent, indent, indent, indent, getGraphqlFieldNamesString(message, false))
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
	fields := getMutationFields(message)
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
	fields := getMutationFields(message)
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
	fields := getMutationFields(message)
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
	varName := strcase.ToLowerCamel(getFieldProtoName(field))
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
	} else {
		kind := field.Desc.Kind()
		var ok bool
		graphqlType, ok = protoToGraphqlTypes[kind]
		if !ok {
			return "", errors.New(fmt.Sprintf("unknown graphql type for proto type %s", kind))
		}
	}

	if fieldIsRepeated(field) {
		graphqlType = fmt.Sprintf("[%s]", graphqlType)
	}
	if !fieldIsOptional(field) {
		graphqlType = fmt.Sprintf("%s!", graphqlType)
	}

	return graphqlType, nil
}

func getCreateMutationInput(message *protogen.Message) string {
	fields := getMutationFieldNames(message)
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
	fields := getMutationFieldNames(message)
	builder := strings.Builder{}
	builder.WriteString("{")
	inputFields := []string{}
	for _, field := range fields {
		inputField := fmt.Sprintf("%s: $%s", field, field)
		inputFields = append(inputFields, inputField)
	}
	return fmt.Sprintf("{%s}", strings.Join(inputFields, ","))
}

func getGraphqlFieldNamesString(message *protogen.Message, includeMessages bool) string {
	return strings.Join(getGraphqlFieldNames(message, includeMessages), " ")
}
func getGraphqlFieldNames(message *protogen.Message, includeMessages bool) []string {
	fieldNames := []string{}
	for _, field := range getNonIgnoredFields(message) {
		if !includeMessages && fieldTypeIsMessage(field) {
			continue
		}
		graphqlFieldName := getGraphqlFieldName(field)
		fieldNames = append(fieldNames, graphqlFieldName)
	}

	return fieldNames
}

func getMutationFieldNames(message *protogen.Message) []string {
	fields := getMutationFields(message)
	mutationFieldNames := []string{}
	for _, field := range fields {
		mutationFieldNames = append(mutationFieldNames, getGraphqlFieldName(field))
	}

	return mutationFieldNames
}

func getMutationFields(message *protogen.Message) []*protogen.Field {
	mutationFields := []*protogen.Field{}
	for _, field := range getNonIgnoredFields(message) {
		if !fieldTypeIsMessage(field) {
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
