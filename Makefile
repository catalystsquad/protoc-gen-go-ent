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
	go install github.com/lyft/protoc-gen-star/protoc-gen-debug
	go install
	buf generate --template example/example/buf.gen.yaml --path example/example
clean:
	rm -f example/cockroachdb/*.go
	rm -f example/postgres/*.go
	rm -f options/*.go
generate: clean build-options build-example
test: generate
	go test -v ./test