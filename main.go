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
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			return plugin.HandleProtoFile(gen, f)
		}
		return nil
	})
}
