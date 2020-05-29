APP     := gomvc
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output
export APP_PATH:= $(HOMEDIR)

# init command params
GO      := go
GOMOD   := $(GO) mod
GOBUILD := $(GO) build
GOTEST  := $(GO) test -gcflags="-N -l"
GOPKGS  := $$($(GO) list ./...| grep -vE "vendor")
all: prepare compile package

# set proxy env
set-env:
	$(GO) env -w GO111MODULE=on
	$(GO) env -w GONOSUMDB=\*

#make prepare, download dependencies
prepare: gomod
gomod: set-env
	$(GOMOD) download

#make compile
compile: build
build:
	$(GO) fmt $(GOPKGS)
	$(GO) vet -composites=false $(GOPKGS)
	$(GOBUILD) -o $(HOMEDIR)/$(APP)

# make test, test your code
test: test-case
test-case:
	$(GO) fmt $(GOPKGS)
	$(GOTEST) -v -cover $(GOPKGS)

# make package
package: package-bin
package-bin:
	mkdir -p $(OUTDIR)/bin
	mkdir -p $(OUTDIR)/log
	mkdir -p $(OUTDIR)/conf
	mv $(APP) $(OUTDIR)/bin
	cp conf/db.toml  $(OUTDIR)/conf/
	cp conf/$(APP).toml $(OUTDIR)/conf

# make clean
clean:
	$(GO) clean
	rm -rf $(OUTDIR)
	rm -rf $(HOMEDIR)/$(APP)
	rm -rf $(GOPATH)/pkg/darwin_amd64
# avoid filename conflict and speed up build 
.PHONY: all prepare compile test package clean build
