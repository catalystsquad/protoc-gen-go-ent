package plugin

import (
	"github.com/catalystsquad/protoc-gen-go-ent/config"
	"strings"
)

func replaceResolverPackagePath(template string) string {
	return strings.ReplaceAll(template, "ResolverPackage", *config.ResolverPackage)
}

func replaceEntPackagePath(template string) string {
	return strings.ReplaceAll(template, "EntPackagePath", *config.EntPackagePath)
}

func replaceObjectName(template string, object SchemaObject) string {
	return strings.ReplaceAll(template, "ObjectName", object.GoType)
}

func replaceObjectPluralName(template string, object SchemaObject) string {
	return strings.ReplaceAll(template, "ObjectPluralName", object.PluralName)
}
