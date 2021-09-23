GOPATH?=$(shell go env GOPATH)
FIRST_GOPATH:=$(firstword $(subst :, ,$(GOPATH)))
GOBIN:=$(FIRST_GOPATH)/bin
GOSRC:=$(FIRST_GOPATH)/src

# Delete the default suffixes
.SUFFIXES:
