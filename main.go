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
		if *config.GenerateApp {
			err := plugin.GenerateApp(gen)
			if err != nil {
				return err
			}
		}
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			err := plugin.HandleProtoFile(gen, f)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func setupFlags() flag.FlagSet {
	var flags flag.FlagSet
	config.LogLevel = flags.String("loglevel", "error", "logging level")
	config.GenerateApp = flags.Bool("genapp", false, "set to true to generate an ent graphql app")
	return flags
}

func setupLogging() {
	flag.Set("stderrthreshold", *config.LogLevel)
	flag.Set("logtostderr", "true")
	flag.Parse()
}
