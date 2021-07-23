# For the version tagging, you must use annotated tags (or git gets unhappy).
# So:
# git tag -a vX.Y.Z
# <enter annotation - doesn't matter what you say, there just has to be one>

SRC = $(shell find . -name "*.go")
PROG = dbp

GIT_VERSION = $(shell git describe --long --dirty || echo wtf)

empty:
	@echo "Make targets:"
	@echo "make dbp"
	@echo "make clean"
	@echo "make start_pg"
	@echo "make stop_pg"

$(PROG): $(SRC) #api/server/tm.pb.go
	go build -o $(PROG) -v -ldflags "-w -s -X main.Version=$(GIT_VERSION)"

.PHONY: clean
clean:
	@rm -f $(PROG)

.PHONY: start_pg
# Start postgres in a docker-compose session in daemon mode
start_pg:
	(cd testdata/env/pg && \
	sh startdb.sh -d && \
	sleep 10 && \
	../../../dbp apply -e postgres -z dbcreds.js -f patches \
	)

.PHONY: stop_pg
stop_pg:
	(cd testdata/env/pg && docker-compose down)


