all: install


########################################
### DEP

DEP = github.com/golang/dep/cmd/dep
DEP_CHECK := $(shell command -v dep 2> /dev/null)

check_tools:
ifndef DEP_CHECK
	@echo "No dep in path.  Install with 'make get_tools'."
else
	@echo "Found dep in path."
endif

get_tools:
ifdef DEP_CHECK
	@echo "Dep is already installed.  Run 'make update_tools' to update."
else
	@echo "$(ansi_grn)Installing dep$(ansi_end)"
	go get -v $(DEP)
endif

update_tools:
	@echo "$(ansi_grn)Updating dep$(ansi_end)"
	go get -u -v $(DEP)


########################################
### Install tools


get_vendor_deps: check_tools
	@rm -rf vendor/
	@echo "--> Running dep ensure"
	@dep ensure -v

install: get_vendor_deps
	@echo "$(ansi_grn)Installing tools$(ansi_end)"
	@echo "$(ansi_yel)Install go-vendorinstall$(ansi_end)"
	go build -o bin/go-vendorinstall go-vendorinstall/*.go

	@echo "$(ansi_yel)Install gometalinter.v2$(ansi_end)"
	GOBIN="$(CURDIR)/bin" ./bin/go-vendorinstall github.com/alecthomas/gometalinter

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
.PHONY: check_tools install_tools update_tools get_vendor_deps install
