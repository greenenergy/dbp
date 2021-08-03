# For the version tagging, you must use annotated tags (or git gets unhappy).
# So:
# git tag -a vX.Y.Z
# <enter annotation - doesn't matter what you say, there just has to be one>

SRC = $(shell find . -name "*.go")
PROG = dbp

GIT_VERSION = $(shell git describe --long --dirty || echo wtf)

empty:
	@echo "Make targets:"
	@echo "make $(PROG)"
	@echo "make clean"
	@echo "make install"

$(PROG): $(SRC) #api/server/tm.pb.go
	go build -o $(PROG) -v -ldflags "-w -s -X main.Version=$(GIT_VERSION)"

.PHONY: install
install:
	go install -v -ldflags "-w -s -X main.Version=$(GIT_VERSION)"

.PHONY: clean
clean:
	@rm -f $(PROG)


