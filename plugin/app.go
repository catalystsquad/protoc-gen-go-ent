package plugin

import (
	"fmt"
	"github.com/golang/glog"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

const appPackageName = "app"

func GenerateApp(gen *protogen.Plugin) error {
	writeGenerate(gen)
	writeEntc(gen)
	writeGqlGen(gen)
	writeResolver(gen)
	writeEntResolvers(gen)
	writeServer(gen)
	writeGoMod(gen)
	return nil
}

func getAppDirectory() string {
	return "app"
}

func getEntDirectory() string {
	return "ent"
}

func getAppFileName(paths ...string) string {
	path := []string{getAppDirectory()}
	path = append(path, paths...)
	return strings.Join(path, "/")
}

func writeGenerate(gen *protogen.Plugin) {
	fileName := getAppFileName("generate.go")
	g := gen.NewGeneratedFile(fileName, "")
	g.P("package ", appPackageName)
	g.P()
	g.P(`//go:generate go run -mod=mod ./ent/entc.go`)
	g.P(`//go:generate go run -mod=mod github.com/99designs/gqlgen`)
}

func writeEntc(gen *protogen.Plugin) {
	fileName := getAppFileName(getEntDirectory(), "entc.go")
	g := gen.NewGeneratedFile(fileName, "")
	g.P(entcContent)
}

func writeGqlGen(gen *protogen.Plugin) {
	fileName := getAppFileName("gqlgen.yml")
	g := gen.NewGeneratedFile(fileName, "")
	g.P(gqlgenContent)
	g.P("autobind:")
	g.P("  - app/ent")
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		for _, m := range f.Messages {
			if getMessageOptions(m).Gen {
				messageProtoName := getMessageProtoName(m)
				thing := fmt.Sprintf("  - app/ent/%s", strings.ToLower(messageProtoName))
				glog.Infof(thing)
				g.P(thing)
			} else {
			}
		}
	}
	g.P("schema:")
	g.P("  - ent.graphql")
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		for _, m := range f.Messages {
			if getMessageOptions(m).Gen {
				messageProtoName := getMessageProtoName(m)
				thing := fmt.Sprintf("  - %s.graphql", strings.ToLower(messageProtoName))
				glog.Infof(thing)
				g.P(thing)
			} else {
			}
		}
	}
}

func writeResolver(gen *protogen.Plugin) {
	fileName := getAppFileName("resolver.go")
	g := gen.NewGeneratedFile(fileName, "")
	g.P(resolverContent)
}

func writeEntResolvers(gen *protogen.Plugin) {
	fileName := getAppFileName("ent.resolvers.go")
	g := gen.NewGeneratedFile(fileName, "")
	g.P(resolversContent)
	for _, f := range gen.Files {
		if !f.Generate {
			glog.Infof("writeEntResolvers skipping file: %s", f.Desc.FullName())
			continue
		}
		glog.Infof("writeEntResolvers handling file: %s", f.Desc.FullName())
		for _, m := range f.Messages {
			glog.Infof("writeEntResolvers handling message %s", getMessageProtoName(m))
			if getMessageOptions(m).Gen {
				messageProtoName := getMessageProtoName(m)
				definition := fmt.Sprintf("func (r *queryResolver) %ss(ctx context.Context, after *ent.Cursor, first *int, before *ent.Cursor, last *int, orderBy *ent.TodoOrder) (*ent.%sConnection, error) {return r.client.%s.Query().Paginate(ctx, after, first, before, last, ent.With%sOrder(orderBy), ent.With%sFilter(where.Filter))}", messageProtoName, messageProtoName, messageProtoName, messageProtoName, messageProtoName)
				glog.Infof(definition)
				g.P(definition)
			} else {
			}
		}
	}
}

func writeServer(gen *protogen.Plugin) {
	fileName := getAppFileName("cmd", "graphql", "main.go")
	g := gen.NewGeneratedFile(fileName, "")
	g.P(serverContent)
}

func writeGoMod(gen *protogen.Plugin) {
	fileName := getAppFileName("go.mod")
	g := gen.NewGeneratedFile(fileName, "")
	g.P(goModContent)
}

var resolversContent = `
package app

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"app/ent"
	"fmt"
)

// Node is the resolver for the node field.
func (r *queryResolver) Node(ctx context.Context, id int) (ent.Noder, error) {
	return r.client.Noder(ctx, id)
}

// Nodes is the resolver for the nodes field.
func (r *queryResolver) Nodes(ctx context.Context, ids []int) ([]ent.Noder, error) {
	return r.client.Noders(ctx, ids)
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
`
var goModContent = `
module app

go 1.21

require (
	entgo.io/contrib v0.4.5
	entgo.io/ent v0.13.1
	github.com/99designs/gqlgen v0.17.44
	github.com/mattn/go-sqlite3 v1.14.16
)

require (
	ariga.io/atlas v0.19.1-0.20240203083654-5948b60a8e43 // indirect
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl/v2 v2.13.0 // indirect
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/sosodev/duration v1.2.0 // indirect
	github.com/vektah/gqlparser/v2 v2.5.11 // indirect
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.9 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	github.com/zclconf/go-cty v1.8.0 // indirect
	golang.org/x/exp v0.0.0-20221230185412-738e83a70c30 // indirect
	golang.org/x/mod v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.18.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
`

var serverContent = `
package main

import (
	"context"
	"app"
	"app/ent"
	"log"
	"net/http"
	"entgo.io/contrib/entgql"

	"entgo.io/ent/dialect"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Create ent.Client and run the schema migration.
	client, err := ent.Open(dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatal("opening ent client", err)
	}
	if err := client.Schema.Create(
		context.Background(),
	); err != nil {
		log.Fatal("opening ent client", err)
	}

	// Configure the server and start listening on :8085.
	srv := handler.NewDefaultServer(app.NewSchema(client))
    srv.Use(entgql.Transactioner{TxOpener: client})
	http.Handle("/",
		playground.Handler("Todo", "/graphql"),
	)
	http.Handle("/graphql", srv)
	log.Println("listening on :8085")
	if err := http.ListenAndServe(":8085", nil); err != nil {
		log.Fatal("http server terminated", err)
	}
}
`

var resolverContent = `
package app

import (
    "app/ent"
    
    "github.com/99designs/gqlgen/graphql"
)

// Resolver is the resolver root.
type Resolver struct{ client *ent.Client }

// NewSchema creates a graphql executable schema.
func NewSchema(client *ent.Client) graphql.ExecutableSchema {
    return NewExecutableSchema(Config{
        Resolvers: &Resolver{client},
    })
}
`

var gqlgenContent = `
# resolver reports where the resolver implementations go.
resolver:
  layout: follow-schema
  dir: .

# gqlgen will search for any type names in the schema in these go packages
# if they match it will use them, otherwise it will generate them.
# This section declares type mapping between the GraphQL and Go type systems.
models:
  # Defines the ID field as Go 'int'.
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.UUID
  Node:
    model:
      - app/ent.Noder
`

var entcContent = `
//go:build ignore

package main

import (
    "log"

    "entgo.io/ent/entc"
    "entgo.io/ent/entc/gen"
    "entgo.io/contrib/entgql"
)

func main() {
    ex, err := entgql.NewExtension(
		entgql.WithWhereInputs(true),
		entgql.WithConfigPath("gqlgen.yml"),
        entgql.WithSchemaGenerator(),
        entgql.WithSchemaPath("ent.graphql"),
    )
    if err != nil {
        log.Fatalf("creating entgql extension: %v", err)
    }
    opts := []entc.Option{
        entc.Extensions(ex),
    }
    if err := entc.Generate("./ent/schema", &gen.Config{}, opts...); err != nil {
        log.Fatalf("running ent codegen: %v", err)
    }
}
`
