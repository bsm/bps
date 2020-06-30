default: vet test

.common.makefile:
	curl -fsSL -o $@ https://gitlab.com/bsm/misc/raw/master/make/go/common.makefile
include .common.makefile

# go get -u github.com/davelondon/rebecca/cmd/becca
README.md: README.md.tpl *example_test.go
	becca -package github.com/bsm/bps
