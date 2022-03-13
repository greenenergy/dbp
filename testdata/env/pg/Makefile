.PHONY: all
all:
	@echo "Usage:"
	@echo "make docker - build docker image necessary to initialize test db"
	@echo "make start - start postgres test environment. Make sure you 'make docker' first"
	@echo "make install - install patches, after 'make start'"
	@echo "make stop - stop postgres test environment"
	@echo "make clean - remove all trace of the pg db. Otherwise it survives docker-compose restarts"

.PHONY: docker
docker:
	(cd docker && docker build . -t postgres-setup:latest)

.PHONY: start
start:
	sh startdb.sh -d

.PHONY: install
install:
	../../../dbp apply -e postgres \
	--db.host="localhost" \
	--db.port=5432 \
	--db.name="patchtest" \
	--db.username="pgpatchery" \
	--db.password="badpasswordx" \
	-f patches -v

stop:
	docker-compose down

clean:
	rm -rf data
