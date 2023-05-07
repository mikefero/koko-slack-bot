# --------------------------------------------------
# Tools/Misc. Targets
# --------------------------------------------------

.PHONY: install-tools
install-tools:
	@echo $(APP_LOG_FMT) "installing developer tools"
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %