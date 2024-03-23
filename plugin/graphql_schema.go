package plugin

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func generateGraphqlSchemaFiles(gen *protogen.Plugin, objects []SchemaObject) {
	for _, object := range objects {
		generateGraphqlSchemaFile(gen, object)
	}
}

func generateGraphqlSchemaFile(gen *protogen.Plugin, object SchemaObject) {
	g := createGraphqlSchemaFile(gen, object)
	def := replaceObjectName(mutationTemplate, object)
	def = replaceObjectPluralName(def, object)
	g.P(def)
}

func createGraphqlSchemaFile(gen *protogen.Plugin, object SchemaObject) *protogen.GeneratedFile {
	g := gen.NewGeneratedFile(object.GraphqlSchemaFileName, object.SchemaFileGoImportPath)
	return g
}

var mutationTemplate = `
extend type Mutation {
  createObjectName(input: CreateObjectNameInput!): ObjectName
  createObjectPluralName(input: [CreateObjectNameInput!]!): [ObjectName!]
  updateObjectName(id: ID!, input: UpdateObjectNameInput!): ObjectName
  updateObjectPluralName(input: [UpdateObjectPluralNameInput!]!): [ObjectName!]
  deleteObjectName(id: ID!): Boolean!
  deleteObjectPluralName(ids: [ID!]!): Boolean!
}

input UpdateObjectPluralNameInput {
  id: ID!
  ObjectName: UpdateObjectNameInput!
}
`
