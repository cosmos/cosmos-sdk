MODULES := group poa

all:
	@for m in $(MODULES); do \
		echo ">>> $$m"; \
		$(MAKE) -C $$m || exit 1; \
	done

# Run target in all modules: make all-<target>
all-%:
	@for m in $(MODULES); do \
		echo ">>> $$m"; \
		$(MAKE) -C $$m $* || exit 1; \
	done

# Run target in specific module: make <module>-<target>
group-%:
	@$(MAKE) -C group $*

poa-%:
	@$(MAKE) -C poa $*
