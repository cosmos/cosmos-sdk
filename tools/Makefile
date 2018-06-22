all: get_tools


########################################
### DEP

DEP = github.com/golang/dep/cmd/dep
GOLINT = github.com/tendermint/lint/golint
GOMETALINTER = gopkg.in/alecthomas/gometalinter.v2
DEP_CHECK := $(shell command -v dep 2> /dev/null)
GOLINT_CHECK := $(shell command -v golint 2> /dev/null)
GOMETALINTER_CHECK := $(shell command -v gometalinter.v2 2> /dev/null)

check_tools:
ifndef DEP_CHECK
	@echo "No dep in path.  Install with 'make get_tools'."
else
	@echo "Found dep in path."
endif
ifndef GOLINT_CHECK
	@echo "No golint in path.  Install with 'make get_tools'."
else
	@echo "Found golint in path."
endif
ifndef GOMETALINTER_CHECK
	@echo "No gometalinter in path.  Install with 'make get_tools'."
else
	@echo "Found gometalinter in path."
endif

get_tools:
ifdef DEP_CHECK
	@echo "Dep is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing dep"
	go get -v $(DEP)
endif
ifdef GOLINT_CHECK
	@echo "Golint is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing golint"
	go get -v $(GOLINT)
endif
ifdef GOMETALINTER_CHECK
	@echo "Gometalinter.v2 is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing gometalinter.v2"
	go get -v $(GOMETALINTER)
endif

update_tools:
	@echo "Updating dep"
	go get -u -v $(DEP)
	@echo "Updating tendermint/golint"
	go get -u -v $(GOLINT)
	@echo "Updating gometalinter.v2"
	go get -u -v $(GOMETALINTER)

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: check_tools get_tools update_tools
