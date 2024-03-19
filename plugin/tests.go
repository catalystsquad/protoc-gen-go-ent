package plugin

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

var testFile *protogen.GeneratedFile

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
	err := generateCreateTest(message)
	if err != nil {
		return err
	}
	err = generateGetByIdTest(message)
	if err != nil {
		return err
	}
	err = generateHelperFunctions(message)
	if err != nil {
		return err
	}

	return nil
}

func generateCreateTest(message *protogen.Message) error {
	objectName := getCreateObjectName(message)
	testFile.P("func Test", objectName, "(t *testing.T) {")
	newFakeObjectFunctionName := getNewFakeFunctionName(message)
	createObjectFunctionName := getCreateFunctionName(message)
	testFile.P(indent, "fake := ", newFakeObjectFunctionName, "()")
	testFile.P("actual, err := ", createObjectFunctionName, "(fake)")
	testFile.P("require.NoError(t, err)")
	assertFunctionName := getAssertCreateEqualityFunctionName(message)
	testFile.P(assertFunctionName, "(t, fake, actual)")
	testFile.P("}")
	testFile.P()
	return nil
}

func generateGetByIdTest(message *protogen.Message) error {
	objectName := getMessageProtoName(message)
	testFile.P("func TestGet", objectName, "ById(t *testing.T) {")
	newFakeObjectFunctionName := getNewFakeFunctionName(message)
	createObjectFunctionName := getCreateFunctionName(message)
	testFile.P(indent, "fake := ", newFakeObjectFunctionName, "()")
	testFile.P("actual, err := ", createObjectFunctionName, "(fake)")
	testFile.P("require.NoError(t, err)")
	testFile.P("fetched, err := ", getGetByIdFunctionName(message), "(actual.ID)")
	testFile.P("require.NoError(t, err)")
	assertFunctionName := getAssertGetByIdEqualityFunctionName(message)
	testFile.P(assertFunctionName, "(t, actual, fetched)")
	testFile.P("}")
	testFile.P()
	return nil
}

func generateHelperFunctions(message *protogen.Message) error {
	generateCreateFunction(message)
	err := generateNewFakeObjectFunction(message)
	if err != nil {
		return err
	}
	generateAssertCreateEqualityFunction(message)
	generateAssertGetByIdEqualityFunction(message)
	generateGetByIdFunction(message)
	return nil
}

func generateCreateFunction(message *protogen.Message) {
	objectName := getCreateObjectName(message)
	inputVarName := strcase.ToLowerCamel(objectName)
	clientType := getCreateObjectClientType(message)
	functionName := getCreateFunctionName(message)
	testFile.P("func ", functionName, "(", inputVarName, " client.", clientType, ") (client.", clientType, ", error) {")
	testFile.P("ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)")
	testFile.P("defer cancel()")
	testFile.P("response, err := gqlClient.", objectName, "(ctx, ", getCreateArgs(message), ")")
	testFile.P("if err != nil { return client.", clientType, "{}, err }")
	testFile.P("return response.", objectName, ", nil")
	testFile.P("}")
}

func getCreateFunctionName(message *protogen.Message) string {
	protoName := getMessageProtoName(message)
	return fmt.Sprintf("create%s", protoName)
}

func generateNewFakeObjectFunction(message *protogen.Message) error {
	objectName := getCreateObjectClientType(message)
	functionName := getNewFakeFunctionName(message)
	testFile.P("func ", functionName, "() client.", objectName, "{")
	testFile.P(indent, "fake := client.", objectName, "{}")
	err := writeFieldFakeData(message)
	if err != nil {
		return err
	}
	testFile.P(indent, "return fake")
	testFile.P("}")
	return nil
}

func generateAssertCreateEqualityFunction(message *protogen.Message) {
	functionName := getAssertCreateEqualityFunctionName(message)
	expectedType := getCreateObjectClientType(message)
	def := fmt.Sprintf("func %s(t *testing.T, expected, actual client.%s) {", functionName, expectedType)
	glog.Infof("create def: %s", def)
	testFile.P(def)
	writeFieldAssertions(message)
	testFile.P("}")
}

func generateAssertGetByIdEqualityFunction(message *protogen.Message) {
	functionName := getAssertGetByIdEqualityFunctionName(message)
	expectedType := getCreateObjectClientType(message)
	actualType := getGetObjectByIdClientType(message)
	def := fmt.Sprintf("func %s(t *testing.T, expected client.%s, actual *client.%s) {", functionName, expectedType, actualType)
	glog.Infof("get def: %s", def)
	testFile.P(def)
	writeFieldAssertions(message)
	testFile.P("}")
}

func generateGetByIdFunction(message *protogen.Message) {
	testFile.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/google/uuid"})
	messageTypeName := getMessageProtoName(message)
	clientName := getGetByIdObjectName(message)
	functionName := getGetByIdFunctionName(message)
	testFile.P("func ", functionName, "(id uuid.UUID) (*client.", clientName, "_", messageTypeName, ", error) {")
	testFile.P("ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)")
	testFile.P("defer cancel()")
	testFile.P("response, err := gqlClient.", messageTypeName, "ByID(ctx, id)")
	testFile.P("if err != nil { return nil, err }")
	testFile.P("return response.", messageTypeName, ", nil")
	testFile.P("}")
}

func getNewFakeFunctionName(message *protogen.Message) string {
	objectName := getCreateObjectName(message)
	return fmt.Sprintf("newFake%s", objectName)
}

func getAssertCreateEqualityFunctionName(message *protogen.Message) string {
	return fmt.Sprintf("assertCreate%sEquality", getMessageProtoName(message))
}

func getAssertGetByIdEqualityFunctionName(message *protogen.Message) string {
	return fmt.Sprintf("assert%sByIdEquality", getMessageProtoName(message))
}

func getGetByIdFunctionName(message *protogen.Message) string {
	messageType := getMessageProtoName(message)
	return fmt.Sprintf("get%sById", messageType)
}

func writeFieldFakeData(message *protogen.Message) error {
	fields := getNonMessageFields(message)
	var err error
	for _, field := range fields {
		if !isIdField(field) {
			if fieldIsRepeated(field) {
				err = writeRepeatedFieldFakeDefinition(field)
			} else {
				err = writeFieldFakeDefinition(field)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func writeFieldAssertions(message *protogen.Message) {
	fields := getNonMessageFields(message)
	for _, field := range fields {
		if !isIdField(field) {

			if fieldIsRepeated(field) {
				writeRepeatedFieldEqualityAssertion(field)
			} else {
				writeFieldEqualityAssertion(field)
			}

		}
	}
}

func writeRepeatedFieldEqualityAssertion(field *protogen.Field) {
	fieldGoName := getRepeatedFieldName(field)
	testFile.P("for _, expected := range expected.", fieldGoName, "{")
	testFile.P(indent, "require.True(t, ", "lo.Contains(actual.", fieldGoName, ", expected))")
	testFile.P("}")
}

func getRepeatedFieldName(field *protogen.Field) string {
	fieldGoName := getFieldGoName(field)
	lastChar := fieldGoName[len(fieldGoName)-1:]
	fieldGoName = fieldGoName[:len(fieldGoName)-1]
	fieldGoName = fieldGoName + strings.ToLower(lastChar)

	return fieldGoName
}

func writeFieldEqualityAssertion(field *protogen.Field) {
	fieldGoName := getFieldGoName(field)
	testFile.P("require.Equal(t, expected.", fieldGoName, ", actual.", fieldGoName, ")")
}

func writeFieldFakeDefinition(field *protogen.Field) error {
	goFieldName := getFieldGoName(field)
	gofakeitFunc, err := getGoFakeItFunctionForFieldBasedOnType(field)
	if err != nil {
		return err
	}
	testFile.P("fake.", goFieldName, " = ", gofakeitFunc)
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

func writeRepeatedFieldFakeDefinition(field *protogen.Field) error {
	fieldPath := getFakeFieldPath(field)
	gofakeitFunc, err := getGoFakeItFunctionForFieldBasedOnType(field)
	if err != nil {
		return err
	}
	testFile.P("for i := 0; i < v6.Number(1, 3); i++ {")
	testFile.P(indent, fieldPath, " = append(", fieldPath, ", ", gofakeitFunc, ")")
	testFile.P("}")
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
