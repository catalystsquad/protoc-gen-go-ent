package plugin

import (
	"fmt"
	"github.com/emirpasic/gods/sets/hashset"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

func generateGqlgencYaml(gen *protogen.Plugin) {
	g := createGqlGencYamlFile(gen)
	g.P(gqlgencTemplate)
}

func generateGqlgenYaml(gen *protogen.Plugin, objects []SchemaObject) {
	g := createGqlGenYamlFile(gen)
	writeResolverConfig(g, objects)
	writeAutoBind(g, objects)
	writeModels(g, objects)
	writeGqlGenSchemaYaml(g, objects)
}

func writeResolverConfig(g *protogen.GeneratedFile, objects []SchemaObject) {
	g.P(resolverTemplate)
}

func createGqlGenYamlFile(gen *protogen.Plugin) *protogen.GeneratedFile {
	fileName := getConfigFileName("gqlgen.yml")
	return gen.NewGeneratedFile(fileName, ".")
}

func createGqlGencYamlFile(gen *protogen.Plugin) *protogen.GeneratedFile {
	fileName := getConfigFileName("gqlgenc.yml")
	return gen.NewGeneratedFile(fileName, ".")
}

func getConfigFileName(name string) string {
	return fmt.Sprintf("config/%s", name)
}

func writeAutoBind(g *protogen.GeneratedFile, objects []SchemaObject) {
	g.P("autobind:")
	g.P(indent, "- app/ent")
	for _, object := range objects {
		g.P(indent, fmt.Sprintf("- app/ent/%s", strings.ToLower(object.GoType)))
	}
}

func writeModels(g *protogen.GeneratedFile, objects []SchemaObject) {
	g.P("models:")
	writeIdModel(g)
	writeNodeModel(g)
	writeFieldModels(g, objects)
}

func writeGqlGenSchemaYaml(g *protogen.GeneratedFile, objects []SchemaObject) {
	g.P("schema:")
	g.P(indent, "- ent.graphql")
	for _, object := range objects {
		g.P(indent, fmt.Sprintf("- %s", object.GraphqlSchemaFileName))
	}
}

func writeFieldModels(g *protogen.GeneratedFile, objects []SchemaObject) {
	modelTypes := hashset.New()
	for _, object := range objects {
		for _, graphqlType := range object.GraphqlTypes.Values() {
			modelTypes.Add(graphqlType)
		}
	}
	for _, modelType := range modelTypes.Values() {
		modelTypeString := modelType.(string)
		ref, ok := gqlgenModelTypeMap[modelTypeString]
		if ok {
			writeModel(g, modelTypeString, ref)
		}
	}
}

var gqlgenModelTypeMap = map[string]string{
	"Uint32": "github.com/99designs/gqlgen/graphql.Uint32",
	"Uint64": "github.com/99designs/gqlgen/graphql.Uint64",
	"Float":  "github.com/99designs/gqlgen/graphql.Float",
}

func writeIdModel(g *protogen.GeneratedFile) {
	writeModel(g, "ID", "github.com/99designs/gqlgen/graphql.UUID")
}

func writeNodeModel(g *protogen.GeneratedFile) {
	writeModel(g, "Node", "app/ent.Noder")
}

func writeModel(g *protogen.GeneratedFile, name, reference string) {
	g.P(indent, name, ":")
	g.P(indent, indent, "model:")
	g.P(indent, indent, indent, "- ", reference)
}

var resolverTemplate = `resolver:
  layout: follow-schema
  dir: .`

var gqlgencTemplate = `model:
  filename: client/models_gen.go # https://github.com/99designs/gqlgen/tree/master/plugin/modelgen
client:
  filename: client/client.go # Where should any generated client go?
query:
  - client/queries/*.graphql # Where are all the query files located?
schema:
  - schema/*.graphql # Where are all the schema files located?
generate:
  clientV2: true # Generate a Client that provides a new signature
  clientInterfaceName: "GraphQLClient" # Determine the name of the generated client interface`
