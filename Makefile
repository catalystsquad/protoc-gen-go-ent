build-debug-request:
	go install github.com/pubg/protoc-gen-debug@latest
	protoc \
        --debug_out=./ \
        --debug_opt=dump_binary=true \
        --debug_opt=dump_json=true \
        --debug_opt=file_binary=request.pb.bin \
        --debug_opt=file_json=request.pb.json \
        -I ./ \
        example/example/example.proto
build-options:
	buf generate --template proto/options/buf.gen.yaml --path proto/options
build-example:
	go install
	buf generate --template example/buf.gen.yaml
clean:
	rm -f options/*.go
	rm -rf example/config
	rm -rf example/graphql_schema
	rm -rf example/resolvers
	rm -rf example/schema
generate-ent:
	cd example/app && go generate
generate-gql-client:
	cd example/app && go get github.com/Yamashou/gqlgenc/clientv2 &&  gqlgenc
generate: clean build-options build-example
test: generate
	go test -v ./test
build-demo:
	cp -r example/schema/* demo_app/ent/schema
	cp example/config/* demo_app
	cp example/resolvers/* demo_app
	cp example/graphql_schema/* demo_app/graphql_schema
	cd demo_app && go generate