# tm - timemanager executale
# Make targets:
# make  		 - builds linux version
#
# For the version tagging, you must use annotated tags (or get gets unhappy).
# So:
# git tag -a vX.Y.Z
# <enter annotation - doesn't matter what you say, there just has to be one>

SRC = $(shell find . -name "*.go")
PROG = dbp

GIT_VERSION = $(shell git describe --long --dirty || echo wtf)

$(PROG): $(SRC) #api/server/tm.pb.go
	go build -o $(PROG) -v -ldflags "-w -s -X main.Version=$(GIT_VERSION)"

.PHONY: clean
clean:
	@rm -f $(PROG)
