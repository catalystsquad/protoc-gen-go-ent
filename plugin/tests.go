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
		testFile = gen.NewGeneratedFile("test/graphql_test.go", "")
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
	testFile.P(indent, "fake := gqlclient.", objectName, "{", objectName, ": gqlclient.", objectName, "_", objectName, "{}}")
	err := writeFieldFakeData(message)
	if err != nil {
		return err
	}
	testFile.P("ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)")
	testFile.P("defer cancel()")
	testFile.P("response, err := client.", objectName, "(ctx, ", getCreateArgs(message), ")")
	testFile.P("require.NoError(t, err)")
	writeFieldAssertions(message)
	testFile.P("}")
	testFile.P()
	return nil
}

func writeFieldFakeData(message *protogen.Message) error {
	fields := getNonMessageFields(message)
	for _, field := range fields {
		err := writeFieldFakeDefinition(field)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeFieldAssertions(message *protogen.Message) {
	fields := getNonMessageFields(message)
	for _, field := range fields {
		objectName := getCreateObjectName(getFieldParentMessage(field))
		fieldGoName := getFieldGoName(field)
		testFile.P("require.Equal(t, fake.", objectName, ".", fieldGoName, ", response.", objectName, ".", fieldGoName, ")")
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
	switch getFieldKind(field) {
	case protoreflect.StringKind:
		definition = "gofakeit.Name()"
	case protoreflect.Int64Kind:
		definition = "int64(gofakeit.Number(0, 1000))"
	default:
		err = errors.New(fmt.Sprintf("unable to get gofakeit function for field kind: %s", getFieldKind(field)))
	}

	return
}

func getCreateObjectName(message *protogen.Message) string {
	return fmt.Sprintf("Create%s", getMessageProtoName(message))
}

func getCreateArgs(message *protogen.Message) string {
	fields := getNonMessageFields(message)
	args := []string{}
	objectName := getCreateObjectName(message)
	for _, field := range fields {
		fieldGoName := getFieldGoName(field)
		args = append(args, fmt.Sprintf("fake.%s.%s", objectName, fieldGoName))
	}

	return strings.Join(args, ",")
}

var testHeader = `
package test

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

var client = gqlclient.NewClient(http.DefaultClient, "http://localhost:8085/graphql", nil)
`
