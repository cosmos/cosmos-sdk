all: get_tools


########################################
### DEP

DEP = github.com/golang/dep/cmd/dep
GOLINT = github.com/tendermint/lint/golint
GOMETALINTER = gopkg.in/alecthomas/gometalinter.v2
UNCONVERT = github.com/mdempsky/unconvert
INEFFASSIGN = github.com/gordonklaus/ineffassign
MISSPELL = github.com/client9/misspell/cmd/misspell
ERRCHECK = github.com/kisielk/errcheck
UNPARAM = mvdan.cc/unparam
GOCYCLO = github.com/alecthomas/gocyclo

DEP_CHECK := $(shell command -v dep 2> /dev/null)
GOLINT_CHECK := $(shell command -v golint 2> /dev/null)
GOMETALINTER_CHECK := $(shell command -v gometalinter.v2 2> /dev/null)
UNCONVERT_CHECK := $(shell command -v unconvert 2> /dev/null)
INEFFASSIGN_CHECK := $(shell command -v ineffassign 2> /dev/null)
MISSPELL_CHECK := $(shell command -v misspell 2> /dev/null)
ERRCHECK_CHECK := $(shell command -v errcheck 2> /dev/null)
UNPARAM_CHECK := $(shell command -v unparam 2> /dev/null)
GOCYCLO_CHECK := $(shell command -v gocyclo 2> /dev/null)

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
ifndef UNCONVERT_CHECK
	@echo "No unconvert in path.  Install with 'make get_tools'."
else
	@echo "Found unconvert in path."
endif
ifndef INEFFASSIGN_CHECK
	@echo "No ineffassign in path.  Install with 'make get_tools'."
else
	@echo "Found ineffassign in path."
endif
ifndef MISSPELL_CHECK
	@echo "No misspell in path.  Install with 'make get_tools'."
else
	@echo "Found misspell in path."
endif
ifndef ERRCHECK_CHECK
	@echo "No errcheck in path.  Install with 'make get_tools'."
else
	@echo "Found errcheck in path."
endif
ifndef UNPARAM_CHECK
	@echo "No unparam in path.  Install with 'make get_tools'."
else
	@echo "Found unparam in path."
endif
ifndef GOCYCLO_CHECK
	@echo "No gocyclo in path.  Install with 'make get_tools'."
else
	@echo "Found gocyclo in path."
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
ifdef UNCONVERT_CHECK
	@echo "Unconvert is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing unconvert"
	go get -v $(UNCONVERT)
endif
ifdef INEFFASSIGN_CHECK
	@echo "Ineffassign is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing ineffassign"
	go get -v $(INEFFASSIGN)
endif
ifdef MISSPELL_CHECK
	@echo "misspell is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing misspell"
	go get -v $(MISSPELL)
endif
ifdef ERRCHECK_CHECK
	@echo "errcheck is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing errcheck"
	go get -v $(ERRCHECK)
endif
ifdef UNPARAM_CHECK
	@echo "unparam is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing unparam"
	go get -v $(UNPARAM)
endif
ifdef GOYCLO_CHECK
	@echo "goyclo is already installed.  Run 'make update_tools' to update."
else
	@echo "Installing goyclo"
	go get -v $(GOCYCLO)
endif

update_tools:
	@echo "Updating dep"
	go get -u -v $(DEP)
	@echo "Updating tendermint/golint"
	go get -u -v $(GOLINT)
	@echo "Updating gometalinter.v2"
	go get -u -v $(GOMETALINTER)
	@echo "Updating unconvert"
	go get -u -v $(UNCONVERT)
	@echo "Updating ineffassign"
	go get -u -v $(INEFFASSIGN)
	@echo "Updating misspell"
	go get -u -v $(MISSPELL)
	@echo "Updating errcheck"
	go get -u -v $(ERRCHECK)
	@echo "Updating unparam"
	go get -u -v $(UNPARAM)
	@echo "Updating goyclo"
	go get -u -v $(GOCYCLO)

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: check_tools get_tools update_tools
