all: install_glide check get_vendor_deps install


########################################
### Glide

GLIDE = github.com/tendermint/glide
GLIDE_CHECK := $(shell command -v glide 2> /dev/null)

check:
ifndef GLIDE_CHECK
	@echo "No glide in path.  Install with 'make install_glide'."
else
	@echo "Found glide in path."
endif

install_glide:
ifdef GLIDE_CHECK
	@echo "Glide is already installed.  Run 'make update_glide' to update."
else
	@echo "$(ansi_grn)Installing glide$(ansi_end)"
	go get -v $(GLIDE)
endif

update_glide:
	@echo "$(ansi_grn)Updating glide$(ansi_end)"
	go get -u -v $(GLIDE)


########################################
### Install tools


get_vendor_deps: check
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install

install: get_vendor_deps
	@echo "$(ansi_grn)Installing tools$(ansi_end)"
	@echo "$(ansi_yel)Install go-vendorinstall$(ansi_end)"
	go build -o bin/go-vendorinstall go-vendorinstall/*.go
	
	@echo "$(ansi_yel)Install gometalinter.v2$(ansi_end)"
	GOBIN=$(CURDIR)/bin ./bin/go-vendorinstall gopkg.in/alecthomas/gometalinter.v2
	
	@echo "$(ansi_yel)Install shelldown$(ansi_end)"
	GOBIN=$(CURDIR)/bin ./bin/go-vendorinstall github.com/rigelrozanski/shelldown/cmd/shelldown

	@echo "$(ansi_grn)Done installing tools$(ansi_end)"


########################################
# ANSI colors

ansi_red=\033[0;31m
ansi_grn=\033[0;32m
ansi_yel=\033[0;33m
ansi_end=\033[0m


# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: check install_glide update_glide get_vendor_deps install
