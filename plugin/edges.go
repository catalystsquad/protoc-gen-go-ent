package plugin

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

func writeEdges(g *protogen.GeneratedFile, message *ParsedMessage) {
	g.P(fmt.Sprintf("func (%s) Edges() []ent.Edge {", message.StructName))
	if len(message.Edges) == 0 {
		g.P("return nil")
	} else {
		g.P("return []ent.Edge{")
		for _, edge := range message.Edges {
			edgeDefinition := getEdgeDefinition(edge)
			g.P(edgeDefinition, ",")
		}
		g.P("}")
	}
	g.P(fmt.Sprintf("}"))
}

// getEdgeDefinition builds the edge definition. If backref is false then this edge is defined as owning the relationship,
//
//	thus To() is used. If backref is true then this edge is defined as not owning the relationship, thus From() is used
func getEdgeDefinition(edge *Edge) string {
	builder := &strings.Builder{}
	builder.WriteString("edge.")
	switch edge.EdgeType {
	case OneToOneTwoTypes:
		writeTo(edge, builder)
		writeUnique(builder)
	case OneToOneSameType:
		if edge.To != "" {
			writeTo(edge, builder)
		} else {
			writeFrom(edge, builder)
			writeRef(edge, builder)
		}
		writeUnique(builder)
	}
	return builder.String()
}

func writeFrom(edge *Edge, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("From(\"%s\", %s.Type)", edge.From, edge.FromType))
}

func writeRef(edge *Edge, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("Ref(\"%s\")", edge.Ref))
}

func writeTo(edge *Edge, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("To(\"%s\", %s.Type)", edge.To, edge.ToType))
}

func writeUnique(builder *strings.Builder) {
	builder.WriteString(".Unique()")
}
