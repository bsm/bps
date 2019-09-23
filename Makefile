default: vet test

vet/%: %
	@cd $< && go vet ./...

test/%: %
	@cd $< && go test ./... --ginkgo.noisySkippings=true --ginkgo.noisyPendings=true

test-verbose/%: %
	@cd $< && go test ./... -v

bench/%: %
	@cd $< && go test ./... -run=NONE -bench=. -benchmem

bump-deps/%: %
	@cd $< && go get -u ./... && go mod tidy

vet: vet/. $(patsubst %/go.mod,vet/%,$(wildcard */go.mod))
test: test/. $(patsubst %/go.mod,test/%,$(wildcard */go.mod))
test-verbose: test-verbose/. $(patsubst %/go.mod,test-verbose/%,$(wildcard */go.mod))
bench: bench/. $(patsubst %/go.mod,bench/%,$(wildcard */go.mod))
bump-deps: bump-deps/. $(patsubst %/go.mod,bump-deps/%,$(wildcard */go.mod))

# go get -u github.com/davelondon/rebecca/cmd/becca
README.md: README.md.tpl
	becca -package github.com/bsm/bps
