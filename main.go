package main

import (
	"flag"
	"github.com/catalystsquad/app-utils-go/env"
	"github.com/catalystsquad/protoc-gen-go-ent/plugin"
	"github.com/golang/glog"
	"google.golang.org/protobuf/compiler/protogen"
)

var logLevel = env.GetEnvOrDefault("LOG_LEVEL", "ERROR")

func main() {
	flag.Set("stderrthreshold", logLevel)
	flag.Parse()
	defer glog.Flush()
	protogen.Options{ParamFunc: flag.CommandLine.Set}.Run(func(gen *protogen.Plugin) error {
		glog.Infof("*****************************************************************")
		for _, f := range gen.Files {
			if !f.Generate {
				glog.Infof("main() skipping file: %s", f.Desc.FullName())
				continue
			}
			glog.Infof("main() handling file: %s", f.Desc.FullName())
			err := plugin.HandleProtoFile(gen, f)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
