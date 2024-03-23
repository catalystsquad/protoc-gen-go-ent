package plugin

import "strings"

func replaceObjectName(template string, object SchemaObject) string {
	return strings.ReplaceAll(template, "ObjectName", object.GoType)
}

func replaceObjectPluralName(template string, object SchemaObject) string {
	return strings.ReplaceAll(template, "ObjectPluralName", object.PluralName)
}
