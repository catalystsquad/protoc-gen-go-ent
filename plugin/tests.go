package plugin

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

var testFile *protogen.GeneratedFile

const newLine = "\n"

func initTestFile(gen *protogen.Plugin, message *protogen.Message) {
	fileName := getAppFileName("test", fmt.Sprintf("%s_graphql_test.go", strcase.ToSnake(getMessageProtoName(message))))
	testFile = gen.NewGeneratedFile(fileName, "")
}

func writeTestPackage() {
	testFile.P("package test")
}

func generateTests(gen *protogen.Plugin, message *protogen.Message) error {
	initTestFile(gen, message)
	writeTestPackage()
	writeTestImports()
	writeTestClient()
	generateCreateTest(message)
	generateUpdateTest(message)
	generateGetByIdTest(message)
	generateDeleteTest(message)
	err := generateHelperFunctions(message)
	if err != nil {
		return err
	}

	return nil
}

func generateCreateTest(message *protogen.Message) {
	testFile.P(templateMessageType(createTestTemplate, message, true))
}

func generateUpdateTest(message *protogen.Message) {
	testFile.P(templateMessageType(updateTestTemplate, message, true))
}

func generateGetByIdTest(message *protogen.Message) {
	testFile.P(templateMessageType(getByIdTestTemplate, message, true))
}

func generateDeleteTest(message *protogen.Message) {
	testFile.P(templateMessageType(deleteTestTemplate, message, true))
}

func generateHelperFunctions(message *protogen.Message) error {
	generateCreateFunction(message)
	generateUpdateFunction(message)
	err := generateNewFakeCreateObjectFunction(message)
	if err != nil {
		return err
	}
	err = generateNewFakeUpdateObjectFunction(message)
	if err != nil {
		return err
	}
	generateAssertCreateEqualityFunction(message)
	generateAssertUpdateEqualityFunction(message)
	generateAssertGetByIdAfterCreateEqualityFunction(message)
	generateAssertGetByIdAfterUpdateEqualityFunction(message)
	generateGetByIdFunction(message)
	generateDeleteFunction(message)
	return nil
}

func generateCreateFunction(message *protogen.Message) {
	def := templateMessageType(createFunctionTemplate, message, true)
	def = fmt.Sprintf(def, getCreateArgs(message))
	testFile.P(def, "\n")
}

func generateUpdateFunction(message *protogen.Message) {
	def := templateMessageType(updateFunctionTemplate, message, true)
	def = fmt.Sprintf(def, getUpdateArgs(message))
	testFile.P(def, "\n")
}

func generateNewFakeCreateObjectFunction(message *protogen.Message) error {
	def := templateMessageType(newFakeCreateFunctionTemplate, message, true)
	body, err := getFieldFakeDataContent(message)
	if err != nil {
		return err
	}
	testFile.P(fmt.Sprintf(def, body))
	return nil
}

func generateNewFakeUpdateObjectFunction(message *protogen.Message) error {
	def := templateMessageType(newFakeUpdateFunctionTemplate, message, true)
	body, err := getFieldFakeDataContent(message)
	if err != nil {
		return err
	}
	testFile.P(fmt.Sprintf(def, body))
	return nil
}

func generateAssertCreateEqualityFunction(message *protogen.Message) {
	def := templateMessageType(assertCreateEqualityTemplate, message, true)
	body := getFieldAssertions(message)
	testFile.P(fmt.Sprintf(def, body))
}

func generateAssertUpdateEqualityFunction(message *protogen.Message) {
	def := templateMessageType(assertUpdateEqualityTemplate, message, true)
	body := getFieldAssertions(message)
	testFile.P(fmt.Sprintf(def, body))
}

func generateAssertGetByIdAfterCreateEqualityFunction(message *protogen.Message) {
	def := templateMessageType(assertGetByIdAfterCreateEqualityTemplate, message, true)
	body := getFieldAssertions(message)
	testFile.P(fmt.Sprintf(def, body))
}

func generateAssertGetByIdAfterUpdateEqualityFunction(message *protogen.Message) {
	def := templateMessageType(assertGetByIdAfterUpdateEqualityTemplate, message, true)
	body := getFieldAssertions(message)
	testFile.P(fmt.Sprintf(def, body))
}

func generateGetByIdFunction(message *protogen.Message) {
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/google/uuid"})
	testFile.P(templateMessageType(getByIdFunctionTemplate, message, true))
}

func generateDeleteFunction(message *protogen.Message) {
	testFile.P(templateMessageType(deleteFunctionTemplate, message, true))
}

func getFieldFakeDataContent(message *protogen.Message) (string, error) {
	builder := &strings.Builder{}
	fields := getNonMessageFields(message)
	var err error
	for _, field := range fields {
		if !isIdField(field) {
			if fieldIsRepeated(field) {
				err = writeRepeatedFieldFakeDefinition(builder, field)
			} else {
				err = writeFieldFakeDefinition(builder, field)
			}
			if err != nil {
				return "", err
			}
		}
	}

	return builder.String(), nil
}

func getFieldAssertions(message *protogen.Message) string {
	builder := &strings.Builder{}
	fields := getNonMessageFields(message)
	for _, field := range fields {
		if !isIdField(field) {
			if fieldIsRepeated(field) {
				writeRepeatedFieldEqualityAssertion(builder, field)
			} else {
				writeFieldEqualityAssertion(builder, field)
			}
		}
	}

	return builder.String()
}

func writeRepeatedFieldEqualityAssertion(builder *strings.Builder, field *protogen.Field) {
	fieldGoName := getRepeatedFieldName(field)
	builder.WriteString(fmt.Sprintf("for _, expected := range expected.%s {\n", fieldGoName))
	builder.WriteString(fmt.Sprintf("require.True(t, lo.Contains(actual.%s, expected))\n", fieldGoName))
	builder.WriteString("}\n")
}

func getRepeatedFieldName(field *protogen.Field) string {
	fieldGoName := getFieldGoName(field)
	lastChar := fieldGoName[len(fieldGoName)-1:]
	fieldGoName = fieldGoName[:len(fieldGoName)-1]
	fieldGoName = fieldGoName + strings.ToLower(lastChar)

	return fieldGoName
}

func writeFieldEqualityAssertion(builder *strings.Builder, field *protogen.Field) {
	fieldGoName := getFieldGoName(field)
	builder.WriteString(fmt.Sprintf("require.Equal(t, expected.%s, actual.%s)\n", fieldGoName, fieldGoName))
}

func writeFieldFakeDefinition(builder *strings.Builder, field *protogen.Field) error {
	goFieldName := getFieldGoName(field)
	gofakeitFunc, err := getGoFakeItFunctionForFieldBasedOnType(field)
	if err != nil {
		return err
	}
	builder.WriteString(fmt.Sprintf("fake.%s = %s\n", goFieldName, gofakeitFunc))
	return nil
}

func getFakeFieldPath(field *protogen.Field) string {
	var goFieldName string
	if fieldIsRepeated(field) {
		goFieldName = getRepeatedFieldName(field)
	} else {
		goFieldName = getFieldName(field)
	}

	return fmt.Sprintf("fake.%s", goFieldName)
}

func writeRepeatedFieldFakeDefinition(builder *strings.Builder, field *protogen.Field) error {
	fieldPath := getFakeFieldPath(field)
	gofakeitFunc, err := getGoFakeItFunctionForFieldBasedOnType(field)
	if err != nil {
		return err
	}
	builder.WriteString("for i := 0; i < v6.Number(1, 3); i++ {\n")
	builder.WriteString(fmt.Sprintf("%s = append(%s, %s)\n", fieldPath, fieldPath, gofakeitFunc))
	builder.WriteString("}\n")
	return nil
}

func getGoFakeItFunctionForFieldBasedOnType(field *protogen.Field) (definition string, err error) {
	if fieldIsEnum(field) {
		packageName := strings.ToLower(getFieldParentMessageType(field))
		testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: protogen.GoImportPath(fmt.Sprintf("app/ent/%s", packageName))})
		EnumName := getFieldGoName(field)
		values := getFieldEnumValues(field)
		values = lo.Map(values, func(item string, index int) string {
			return fmt.Sprintf("\"%s\"", item)
		})
		definition = fmt.Sprintf("%s.%s(v6.RandomString([]string{%s}))", packageName, EnumName, strings.Join(values, ","))
	} else {
		var ok bool
		if fieldIsRepeated(field) {
			definition, ok = protoKindRepeatedFakeDefinitionMap[getFieldKind(field)]
		} else {
			definition, ok = protoKindFakeDefinitionMap[getFieldKind(field)]
		}
		if ok {
			if fieldIsOptional(field) {
				definition = fmt.Sprintf("lo.ToPtr(%s)", definition)
			}
			if strings.Contains(definition, "strconv.") {
				testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "strconv"})
			}
		} else {
			err = errors.New(fmt.Sprintf("unable to get gofakeit function for field kind: %s", getFieldKind(field)))
		}
	}

	return
}

var protoKindFakeDefinitionMap = map[protoreflect.Kind]string{
	protoreflect.BoolKind:     "v6.Bool()",
	protoreflect.EnumKind:     "v6.Bool()",
	protoreflect.Int32Kind:    "v6.Number(0, 1000)",
	protoreflect.Sint32Kind:   "v6.Number(0, 1000)",
	protoreflect.Uint32Kind:   "uint32(v6.Number(0, 1000))",
	protoreflect.Sfixed32Kind: "v6.Number(0, 1000)",
	protoreflect.Fixed32Kind:  "uint32(v6.Number(0, 1000))",
	protoreflect.Sint64Kind:   "v6.Number(0, 1000)",
	protoreflect.Uint64Kind:   "uint64(v6.Number(0, 1000))",
	protoreflect.Sfixed64Kind: "v6.Number(0, 1000)",
	protoreflect.Fixed64Kind:  "uint64(v6.Number(0, 1000))",
	protoreflect.FloatKind:    "float64(v6.Number(0, 1000))",
	protoreflect.DoubleKind:   "float64(v6.Number(0, 1000))",
	protoreflect.BytesKind:    "[]byte(v6.Name())",
	protoreflect.StringKind:   "v6.Name()",
	protoreflect.Int64Kind:    "v6.Number(0, 1000)",
}

var protoKindRepeatedFakeDefinitionMap = map[protoreflect.Kind]string{
	protoreflect.BoolKind:     "v6.Bool()",
	protoreflect.EnumKind:     "v6.Bool()",
	protoreflect.Int32Kind:    "v6.Number(0, 1000)",
	protoreflect.Sint32Kind:   "v6.Number(0, 1000)",
	protoreflect.Uint32Kind:   "v6.Number(0, 1000)",
	protoreflect.Sfixed32Kind: "v6.Number(0, 1000)",
	protoreflect.Fixed32Kind:  "v6.Number(0, 1000)",
	protoreflect.Sint64Kind:   "v6.Number(0, 1000)",
	protoreflect.Uint64Kind:   "v6.Number(0, 1000)",
	protoreflect.Sfixed64Kind: "v6.Number(0, 1000)",
	protoreflect.Fixed64Kind:  "v6.Number(0, 1000)",
	protoreflect.FloatKind:    "float64(v6.Number(0, 1000))",
	protoreflect.DoubleKind:   "float64(v6.Number(0, 1000))",
	protoreflect.BytesKind:    "[]byte(v6.Name())",
	protoreflect.StringKind:   "v6.Name()",
	protoreflect.Int64Kind:    "v6.Number(0, 1000)",
}

func getCreateObjectClientType(message *protogen.Message) string {
	objectName := getCreateObjectName(message)
	return fmt.Sprintf("%s_%s", objectName, objectName)
}

func getGetObjectByIdClientType(message *protogen.Message) string {
	messageType := getMessageProtoName(message)
	objectName := getGetByIdObjectName(message)
	return fmt.Sprintf("%s_%s", objectName, messageType)
}

func getCreateObjectName(message *protogen.Message) string {
	protoName := getMessageProtoName(message)
	return fmt.Sprintf("Create%s", protoName)
}

func getUpdateObjectName(message *protogen.Message) string {
	protoName := getMessageProtoName(message)
	return fmt.Sprintf("Update%s", protoName)
}

func getGetByIdObjectName(message *protogen.Message) string {
	protoName := getMessageProtoName(message)
	return fmt.Sprintf("%sById", protoName)
}

func getCreateArgs(message *protogen.Message) string {
	objectName := getCreateObjectName(message)
	inputVarName := strcase.ToLowerCamel(objectName)
	fields := getNonMessageFields(message)
	args := []string{}
	for _, field := range fields {
		if !isIdField(field) {
			var fieldGoName string
			if fieldIsRepeated(field) {
				fieldGoName = getRepeatedFieldName(field)
			} else {
				fieldGoName = getFieldGoName(field)
			}
			args = append(args, fmt.Sprintf("%s.%s", inputVarName, fieldGoName))
		}
	}

	return strings.Join(args, ",")
}

func getUpdateArgs(message *protogen.Message) string {
	objectName := getUpdateObjectName(message)
	inputVarName := strcase.ToLowerCamel(objectName)
	fields := getNonMessageFields(message)
	args := []string{}
	for _, field := range fields {
		if !isIdField(field) {
			var fieldGoName string
			if fieldIsRepeated(field) {
				fieldGoName = getRepeatedFieldName(field)
			} else {
				fieldGoName = getFieldGoName(field)
			}
			args = append(args, fmt.Sprintf("%s.%s", inputVarName, fieldGoName))
		}
	}

	return strings.Join(args, ",")
}

func writeTestImports() {
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "context"})
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/brianvoe/gofakeit/v6"})
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/stretchr/testify/require"})
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "net/http"})
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "app/client"})
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "time"})
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/samber/lo"})
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "testing"})
}

func writeTestClient() {
	testFile.P("var gqlClient = client.NewClient(http.DefaultClient, \"http://localhost:8085/graphql\", nil)")
}
