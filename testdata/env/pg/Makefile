.PHONY: all
all:
	@echo "This folder is intended to be used as an install verification platform. Copy your patch files into the patches"
	@echo "folder and then run 'make start && make install'. dbp will try to run all your patches into a fresh database."
	@echo "'make stop' will then shut down the db and cleanup."
	@echo "Usage:"
	@echo "make start - start postgres test environment. Make sure you 'make docker' first"
	@echo "make install - install patches, after 'make start'"
	@echo "make stop - stop postgres test environment"
	@echo "make clean - remove all trace of the pg db. Otherwise it survives docker-compose restarts"

.PHONY: start
start:
	#sh startdb.sh -d
	docker-compose up -d

.PHONY: install
install:
	../../../dbp apply -e postgres \
	--db.host="localhost" \
	--db.port=5434 \
	--db.name="patchtest" \
	--db.username="postgres" \
	--db.password="postgres" \
	--db.tls=false \
	-f patches -v -d \
	-r 100

.PHONY: dry
dry:
	../../../dbp apply -e postgres \
	--db.host="localhost" \
	--db.port=5434 \
	--db.name="patchtest" \
	--db.username="postgres" \
	--db.password="postgres" \
	--db.tls=false \
	-f patches -v -d \
	-r 100 --dry


stop:
	docker-compose down

