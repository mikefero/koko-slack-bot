# --------------------------------------------------
# Test Targets
# --------------------------------------------------

COVERAGE_DIR := $(APP_WORKDIR)/coverage
UNIT_TEST_DIR := $(COVERAGE_DIR)/unit
UNIT_TEST_WEBPAGE := $(UNIT_TEST_DIR)/index.html
UNIT_TEST_COVERAGE := $(UNIT_TEST_DIR)/coverage.out

.PHONY: test
test:
	@echo $(APP_LOG_FMT) "executing unit test suite"
	@mkdir -p $(UNIT_TEST_DIR)
	@go test -v \
		-race \
		-covermode=atomic \
		-coverprofile=$(UNIT_TEST_COVERAGE) \
		./internal/...
	@go tool cover -func=$(UNIT_TEST_COVERAGE)
	@go tool cover -html=$(UNIT_TEST_COVERAGE) -o $(UNIT_TEST_WEBPAGE)