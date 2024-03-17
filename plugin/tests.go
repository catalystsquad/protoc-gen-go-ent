package plugin

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

var testFile *protogen.GeneratedFile

func initTestFile(gen *protogen.Plugin) {
	if testFile == nil {
		fileName := getAppFileName("test/graphql_test.go")
		testFile = gen.NewGeneratedFile(fileName, "")
	}
}

func writeTestPackage() {
	testFile.P("package test")
}

func generateTests(gen *protogen.Plugin, message *protogen.Message) error {
	initTestFile(gen)
	writeTestPackage()
	writeTestImports()
	writeTestClient()
	err := generateCreateTest(message)
	if err != nil {
		return err
	}

	return nil
}

func generateCreateTest(message *protogen.Message) error {
	objectName := getCreateObjectName(message)
	testFile.P("func Test", objectName, "(t *testing.T) {")
	testFile.P(indent, "fake := client.", objectName, "{", objectName, ": client.", objectName, "_", objectName, "{}}")
	err := writeFieldFakeData(message)
	if err != nil {
		return err
	}
	testFile.P("ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)")
	testFile.P("defer cancel()")
	testFile.P("response, err := gqlClient.", objectName, "(ctx, ", getCreateArgs(message), ")")
	testFile.P("require.NoError(t, err)")
	writeFieldAssertions(message)
	testFile.P("}")
	testFile.P()
	return nil
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
	objectName := getCreateObjectName(getFieldParentMessage(field))
	fieldGoName := getRepeatedFieldName(field)
	testFile.P("for _, expected := range fake.", objectName, ".", fieldGoName, "{")
	testFile.P(indent, "require.True(t, ", "lo.Contains(fake.", objectName, ".", fieldGoName, ", expected))")
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
	objectName := getCreateObjectName(getFieldParentMessage(field))
	fieldGoName := getFieldGoName(field)
	testFile.P("require.Equal(t, fake.", objectName, ".", fieldGoName, ", response.", objectName, ".", fieldGoName, ")")
}

func writeFieldFakeDefinition(field *protogen.Field) error {
	goFieldName := getFieldGoName(field)
	objectName := getCreateObjectName(getFieldParentMessage(field))
	gofakeitFunc, err := getGoFakeItFunctionForFieldBasedOnType(field)
	if err != nil {
		return err
	}
	testFile.P("fake.", objectName, ".", goFieldName, " = ", gofakeitFunc)
	return nil
}

func getFakeFieldPath(field *protogen.Field) string {
	var goFieldName string
	if fieldIsRepeated(field) {
		goFieldName = getRepeatedFieldName(field)
	} else {
		goFieldName = getFieldName(field)
	}
	objectName := getCreateObjectName(getFieldParentMessage(field))

	return fmt.Sprintf("fake.%s.%s", objectName, goFieldName)
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

func getCreateObjectName(message *protogen.Message) string {
	return fmt.Sprintf("Create%s", getMessageProtoName(message))
}

func getCreateArgs(message *protogen.Message) string {
	fields := getNonMessageFields(message)
	args := []string{}
	objectName := getCreateObjectName(message)
	for _, field := range fields {
		if !isIdField(field) {
			var fieldGoName string
			if fieldIsRepeated(field) {
				fieldGoName = getRepeatedFieldName(field)
			} else {
				fieldGoName = getFieldGoName(field)
			}
			args = append(args, fmt.Sprintf("fake.%s.%s", objectName, fieldGoName))
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
