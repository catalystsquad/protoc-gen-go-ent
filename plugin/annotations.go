package plugin

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
)

func WriteSchemaFileAnnotations(g *protogen.GeneratedFile, message *protogen.Message) {
	messageGoName := getMessageGoName(message)
	g.P(fmt.Sprintf("func (%s) Annotations() []schema.Annotation { return []schema.Annotation {entgql.QueryField(),entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),", messageGoName))
	g.P(fmt.Sprintf("}"))
	g.P(fmt.Sprintf("}"))
}
