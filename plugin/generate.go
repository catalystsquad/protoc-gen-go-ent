package plugin

import "google.golang.org/protobuf/compiler/protogen"

func Generate(gen *protogen.Plugin) error {
	objects, err := parseFiles(gen)
	if err != nil {
		return err
	}
	generateEntSchemas(gen, objects)
	generateGraphqlSchemaFiles(gen, objects)
	generateResolvers(gen, objects)
	generateGqlgenYaml(gen, objects)
	generateGqlgencYaml(gen)
	return nil
}
