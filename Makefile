.PHONY: all clean

.DELETE_ON_ERROR:

# files to generate
GO_FILES := functions.go functions_test.go

# main target
all: $(GO_FILES)
	go test

# code generation rule
%.go: tools/%.php
	php -f $^ > $@
	goimports -w $@

# cleanup
clean:
	$(RM) $(GO_FILES)
