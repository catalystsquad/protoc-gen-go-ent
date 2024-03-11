package plugin

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
)

func WriteSchemaFileAnnotations(g *protogen.GeneratedFile, message *protogen.Message) {
	messageGoName := getMessageGoName(message)
	g.P(fmt.Sprintf("func (%s) Annotations() []schema.Annotation { ", messageGoName))
	g.P("  return []schema.Annotation {")
	g.P("  entgql.QueryField(),")
	g.P("  entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),")
	g.P("  entgql.MultiOrder(),")
	g.P("  entgql.RelayConnection(),")
	g.P("}")
	g.P("}")
}
