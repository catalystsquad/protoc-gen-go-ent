package plugin

import "google.golang.org/protobuf/compiler/protogen"

func HandleProtoFile(gen *protogen.Plugin, file *protogen.File) error {
	for _, message := range file.Messages {
		err := HandleProtoMessage(gen, file, message)
		if err != nil {
			return err
		}
	}
	return nil
}
