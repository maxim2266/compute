.PHONY: all

.DELETE_ON_ERROR:

# main target
all: functions.go functions_test.go
	go test

# code generation rule
%.go: tools/%.php
	php -d error_reporting=E_ALL -d display_errors=On -f $^ > $@
	goimports -w $@
