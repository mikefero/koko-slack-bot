.DEFAULT_GOAL := build

APP_NAME := koko-slack-bot
APP_WORKDIR := $(shell pwd)
APP_DATE_FORMAT := +'%Y-%m-%dT%H:%M:%SZ'
APP_LOG_FMT := $(shell date "$(APP_DATE_FORMAT) [$(APP_NAME)]")

include mk/build.mk
include mk/test.mk
include mk/tools.mk
