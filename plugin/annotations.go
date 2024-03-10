package plugin

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
)

func WriteAnnotations(g *protogen.GeneratedFile, message *protogen.Message) {
	messageGoName := getMessageGoName(message)
	g.P(fmt.Sprintf("func (%s) Annotations() []schema.Annotation { return []schema.Annotation {entgql.QueryField(),entgql.Mutations(entgql.MutationCreate()),", messageGoName))
	g.P(fmt.Sprintf("}"))
	g.P(fmt.Sprintf("}"))
}
