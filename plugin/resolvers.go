package plugin

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

func generateResolvers(gen *protogen.Plugin, objects []SchemaObject) {
	generateRootResolver(gen)
	generateEntResolvers(gen, objects)
	generateObjectResolvers(gen, objects)
}

func generateObjectResolvers(gen *protogen.Plugin, objects []SchemaObject) {
	for _, object := range objects {
		generateObjectResolver(gen, object)
	}
}

func generateObjectResolver(gen *protogen.Plugin, object SchemaObject) {
	g := createObjectResolverFile(gen, object)
	def := replaceResolverPackagePath(objectResolverTemplate)
	def = replaceObjectName(def, object)
	def = replaceObjectPluralName(def, object)
	g.P(def)
}

var objectResolverTemplate = `
package ResolverPackage

// CreateObjectName is the resolver for the createObjectName field.
func (r *mutationResolver) CreateObjectName(ctx context.Context, input ent.CreateObjectNameInput) (*ent.ObjectName, error) {
	return r.client.ObjectName.Create().SetInput(input).Save(ctx)
}

// CreateObjectPluralName is the resolver for the createObjectPluralName field.
func (r *mutationResolver) CreateObjectPluralName(ctx context.Context, input []*ent.CreateObjectNameInput) ([]*ent.ObjectName, error) {
	bulkInput := []*ent.ObjectNameCreate{}
	for _, crateInput := range input {
		bulkInput = append(bulkInput, r.client.ObjectName.Create().SetInput(*crateInput))
	}
	return r.client.ObjectName.CreateBulk(bulkInput...).Save(ctx)
}

// UpdateObjectName is the resolver for the updateObjectName field.
func (r *mutationResolver) UpdateObjectName(ctx context.Context, id uuid.UUID, input ent.UpdateObjectNameInput) (*ent.ObjectName, error) {
	return r.client.ObjectName.UpdateOneID(id).SetInput(input).Save(ctx)
}

// UpdateObjectPluralName is the resolver for the updateObjectPluralName field.
func (r *mutationResolver) UpdateObjectPluralName(ctx context.Context, input []*UpdateObjectPluralNameInput) ([]*ent.ObjectName, error) {
	result := []*ent.ObjectName{}
	for _, updateInput := range input {
		updated, err := r.client.ObjectName.UpdateOneID(updateInput.ID).SetInput(*updateInput.ObjectName).Save(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, updated)
	}
	return result, nil
}

// DeleteObjectName is the resolver for the deleteObjectName field.
func (r *mutationResolver) DeleteObjectName(ctx context.Context, id uuid.UUID) (bool, error) {
	err := r.client.ObjectName.DeleteOneID(id).Exec(ctx)
	return err == nil, err
}

// DeleteObjectPluralName is the resolver for the deleteObjectPluralName field.
func (r *mutationResolver) DeleteObjectPluralName(ctx context.Context, ids []uuid.UUID) (bool, error) {
	for _, id := range ids {
		err := r.client.ObjectName.DeleteOneID(id).Exec(ctx)
		if err != nil {
			return false, err
		}
	}
	
	return true, nil
}
`

func generateEntResolvers(gen *protogen.Plugin, objects []SchemaObject) {
	g := createObjectResolversFile(gen)
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "context"})
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/google/uuid"})
	base := replaceResolverPackagePath(objectResolversContent)
	g.P(base)
	for _, object := range objects {
		optsDef := getResolverOptsDef(object)
		def := fmt.Sprintf(objectResolverDefinitionTemplate, optsDef)
		def = replaceObjectPluralName(def, object)
		def = replaceObjectName(def, object)
		g.P(def, "\n")
	}
}

func getResolverOptsDef(object SchemaObject) string {
	opts := []string{"ent.WithObjectNameFilter(where.Filter)"}
	if objectHasOrderByAnnotation(object) {
		opts = append(opts, "ent.WithObjectNameOrder(orderBy)")
	}
	if len(opts) > 0 {
		return fmt.Sprintf(", %s", strings.Join(opts, ", "))
	}

	return ""
}

func objectHasOrderByAnnotation(object SchemaObject) bool {
	for _, edge := range object.EntEdges {
		for _, edgeAnnotation := range edge.Annotations {
			if strings.Contains(edgeAnnotation, "entgql.OrderField") {
				return true
			}
		}
	}
	return false
}

func createObjectResolversFile(gen *protogen.Plugin) *protogen.GeneratedFile {
	fileName := getResolverFileName("ent.resolvers")
	return gen.NewGeneratedFile(fileName, ".")
}

func createObjectResolverFile(gen *protogen.Plugin, object SchemaObject) *protogen.GeneratedFile {
	return gen.NewGeneratedFile(object.ResolverFileName, ".")
}

func generateRootResolver(gen *protogen.Plugin) {
	g := createRootResolverFile(gen)
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "github.com/99designs/gqlgen/graphql"})
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "app/ent"})
	def := replaceResolverPackagePath(resolverContent)
	def = replaceEntPackagePath(def)
	g.P(def)
}

func createRootResolverFile(gen *protogen.Plugin) *protogen.GeneratedFile {
	fileName := getRootResolverFileName()
	return gen.NewGeneratedFile(fileName, ".")
}

func getRootResolverFileName() string {
	return getResolverFileName("resolver")
}

func getResolverFileName(resolverName string) string {
	return fmt.Sprintf("resolvers/%s.go", resolverName)
}

var resolverContent = `
package ResolverPackage

// Resolver is the resolver root.
type Resolver struct{ client *ent.Client }

// NewSchema creates a graphql executable schema.
func NewSchema(client *ent.Client) graphql.ExecutableSchema {
    return ent.NewExecutableSchema(ent.Config{
        Resolvers: &Resolver{client},
    })
}`

var objectResolversContent = `
package ResolverPackage

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

// Node is the resolver for the node field.
func (r *queryResolver) Node(ctx context.Context, id uuid.UUID) (ent.Noder, error) {
	return r.client.Noder(ctx, id)
}

// Nodes is the resolver for the nodes field.
func (r *queryResolver) Nodes(ctx context.Context, ids []uuid.UUID) ([]ent.Noder, error) {
	return r.client.Noders(ctx, ids)
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
`

var objectResolverDefinitionTemplate = `
func (r *queryResolver) ObjectPluralName(ctx context.Context, after *ent.Cursor, first *int, before *ent.Cursor, last *int, orderBy *ent.ObjectNameOrder) (*ent.ObjectNameConnection, error) {return r.client.ObjectName.Query().Paginate(ctx, after, first, before, last%s)}
`
