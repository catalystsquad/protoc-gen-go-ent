build-options:
	buf generate --template proto/options/buf.gen.yaml --path proto/options
build-example:
	go install
	buf generate --template example/example/buf.gen.yaml --path example/example
clean:
	rm -f example/cockroachdb/*.go
	rm -f example/postgres/*.go
	rm -f options/*.go
generate: clean build-options build-example
test: generate
	go test -v ./test