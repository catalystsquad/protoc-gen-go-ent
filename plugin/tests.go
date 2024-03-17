package plugin

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

var testFile *protogen.GeneratedFile

func initTestFile(gen *protogen.Plugin) {
	if testFile == nil {
		fileName := getAppFileName("test/graphql_test.go")
		testFile = gen.NewGeneratedFile(fileName, "")
		testFile.P(testHeader)
	}
}

func generateTests(gen *protogen.Plugin, message *protogen.Message) error {
	initTestFile(gen)
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
	for _, field := range fields {
		if !isIdField(field) {
			err := writeFieldFakeDefinition(field)
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
			objectName := getCreateObjectName(getFieldParentMessage(field))
			fieldGoName := getFieldGoName(field)
			testFile.P("require.Equal(t, fake.", objectName, ".", fieldGoName, ", response.", objectName, ".", fieldGoName, ")")
		}
	}
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

func getGoFakeItFunctionForFieldBasedOnType(field *protogen.Field) (definition string, err error) {
	definition, ok := protoKindFakeDefinitionMap[getFieldKind(field)]
	if !ok {
		err = errors.New(fmt.Sprintf("unable to get gofakeit function for field kind: %s", getFieldKind(field)))
	}

	return
}

var protoKindFakeDefinitionMap = map[protoreflect.Kind]string{
	protoreflect.BoolKind:     "gofakeit.Bool()",
	protoreflect.EnumKind:     "gofakeit.Bool()",
	protoreflect.Int32Kind:    "int32(gofakeit.Number(0, 1000))",
	protoreflect.Sint32Kind:   "int32(gofakeit.Number(0, 1000))",
	protoreflect.Uint32Kind:   "uint32(gofakeit.Number(0, 1000))",
	protoreflect.Sfixed32Kind: "int32(gofakeit.Number(0, 1000))",
	protoreflect.Fixed32Kind:  "uint32(gofakeit.Number(0, 1000))",
	protoreflect.Sint64Kind:   "int64(gofakeit.Number(0, 1000))",
	protoreflect.Uint64Kind:   "uint64(gofakeit.Number(0, 1000))",
	protoreflect.Sfixed64Kind: "int64(gofakeit.Number(0, 1000))",
	protoreflect.Fixed64Kind:  "uint64(gofakeit.Number(0, 1000))",
	protoreflect.FloatKind:    "float32(gofakeit.Number(0, 1000))",
	protoreflect.DoubleKind:   "float64(gofakeit.Number(0, 1000))",
	protoreflect.BytesKind:    "[]byte(gofakeit.Name())",
	protoreflect.StringKind:   "gofakeit.Name()",
	protoreflect.Int64Kind:    "int64(gofakeit.Number(0, 1000))",
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
			fieldGoName := getFieldGoName(field)
			args = append(args, fmt.Sprintf("fake.%s.%s", objectName, fieldGoName))
		}
	}

	return strings.Join(args, ",")
}

var testHeader = `
package test

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"app/client"
	"time"
)

var gqlClient = client.NewClient(http.DefaultClient, "http://localhost:8085/graphql", nil)
`
