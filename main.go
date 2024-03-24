package main

import (
	"flag"
	"github.com/catalystsquad/protoc-gen-go-ent/config"
	"github.com/catalystsquad/protoc-gen-go-ent/plugin"
	"github.com/golang/glog"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	flags := setupFlags()
	setupLogging()
	defer glog.Flush()

	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		// TODO: Refactor to not use workspaces so this doesn't matter
		if lo.Contains(gen.Request.FileToGenerate, "options/ent.proto") {
			// skip options
			return nil
		}
		return plugin.Generate(gen)
	})
}

func setupFlags() flag.FlagSet {
	var flags flag.FlagSet
	config.LogLevel = flags.String("loglevel", "error", "logging level")
	config.GenerateApp = flags.Bool("genapp", false, "set to true to generate an ent graphql app")
	config.EntSchemaDir = flags.String("entschemadir", "schema", "directory to output ent schemas to")
	config.GraphqlSchemaDir = flags.String("graphqlschemadir", "graphql_schema", "directory to output graphql schemas to")
	config.GraphqlResolverDir = flags.String("graphqlresolverdir", "resolvers", "directory to output graphql resolvers to")
	config.EntPackagePath = flags.String("entpackagepath", "ent", "package path to use for ent generated files")
	config.ResolverPackage = flags.String("resolverpackage", "resolvers", "package path to use for ent generated resolver files")
	return flags
}

func setupLogging() {
	flag.Set("stderrthreshold", *config.LogLevel)
	flag.Set("logtostderr", "true")
	flag.Parse()
}
