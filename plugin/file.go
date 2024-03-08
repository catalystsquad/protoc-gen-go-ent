package plugin

import (
	"github.com/golang/glog"
	"google.golang.org/protobuf/compiler/protogen"
)

func HandleProtoFile(gen *protogen.Plugin, file *protogen.File) error {
	for _, message := range file.Messages {
		if !shouldHandleMessage(message) {
			glog.Infof("skipping message: %s", getMessageProtoName(message))
			continue
		}
		glog.Infof("handling message: %s", getMessageProtoName(message))
		err := HandleProtoMessage(gen, file, message)
		if err != nil {
			return err
		}
	}
	return nil
}
