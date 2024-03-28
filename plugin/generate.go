package plugin

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func Generate(gen *protogen.Plugin) error {
	if shouldGenerate(gen) {
		logImportant("generating for : %s", gen.Request.FileToGenerate)
		objects, err := parseFiles(gen)
		if err != nil {
			return err
		}
		generateEntSchemas(gen, objects)
		generateGraphqlSchemaFiles(gen, objects)
		generateResolvers(gen, objects)
		generateGqlgenYaml(gen, objects)
		generateGqlgencYaml(gen)
	}

	return nil
}

func shouldGenerate(gen *protogen.Plugin) bool {
	numFilesToGenerate := 0
	for _, file := range gen.Files {
		if getFileOptions(file).Generate {
			numFilesToGenerate++
		}
	}

	return numFilesToGenerate > 0
}
