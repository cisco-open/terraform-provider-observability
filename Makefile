TOP_LEVEL=$(shell git rev-parse --show-toplevel)
TOOLSDIR := $(TOP_LEVEL)/.tools
ADDLICENSE := $(TOOLSDIR)/addlicense
FSOC := $(TOOLSDIR)/fsoc
FSOC_VERSION := 0.67.0
ADDLICENSE_VERSION := 1.1.1

$(ADDLICENSE):
	mkdir -p $(TOOLSDIR)
	wget https://github.com/google/addlicense/releases/download/v$(ADDLICENSE_VERSION)/addlicense_$(ADDLICENSE_VERSION)_Linux_x86_64.tar.gz
	tar -xvf addlicense_$(ADDLICENSE_VERSION)_Linux_x86_64.tar.gz -C $(TOOLSDIR) addlicense
	rm addlicense_$(ADDLICENSE_VERSION)_Linux_x86_64.tar.gz 

$(FSOC):
	mkdir -p $(TOOLSDIR)
	wget https://github.com/cisco-open/fsoc/releases/download/v$(FSOC_VERSION)/fsoc-linux-amd64.tar.gz
	tar -xvf fsoc-linux-amd64.tar.gz -C $(TOOLSDIR) fsoc-linux-amd64
	mv $(TOOLSDIR)/fsoc-linux-amd64 $(TOOLSDIR)/fsoc
	rm fsoc-linux-amd64.tar.gz 


.PHONY: all
all: lint test check-license

.PHONY: check-license
check-license: $(ADDLICENSE)
	@echo "verifying license headers"
	$(TOOLSDIR)/addlicense -check .

.PHONY: add-license
add-license: $(ADDLICENSE)
	@echo "adding license headers, please commit any modified files"
	$(ADDLICENSE) -s -v -c "Cisco Systems, Inc. and its affiliates" -l apache .

.PHONY: lint
lint:
	@echo "linting placeholder"

.PHONY: test
test:
	@echo "testing placeholder"

.PHONY: clean
clean:
	rm -rf $(TOOLSDIR)
