include mk/header.mk

COSMOS_PLUGINS ?=
export COSMOS_PLUGINS

$(d)/preload.go: d:=$(d)
$(d)/preload.go: $(d)/preload_list $(d)/preload.sh ALWAYS
	$(d)/preload.sh > $@
	go fmt $@ >/dev/null

DEPS_GO += $(d)/preload.go

include mk/footer.mk
