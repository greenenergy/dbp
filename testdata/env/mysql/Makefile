.PHONY: all
all:
	@echo "This folder is intended to be used as an install verification platform. Copy your patch files into the patches"
	@echo "folder and then run 'make start && make install'. dbp will try to run all your patches into a fresh database."
	@echo "'make stop' will then shut down the db and cleanup."
	@echo "Usage:"
	@echo "make docker - build docker image necessary to initialize test db"
	@echo "make start - start mysql test environment. Make sure you 'make docker' first"
	@echo "make install - install patches, after 'make start'"
	@echo "make stop - stop mysql test environment"
	@echo "make clean - remove all trace of the pg db. Otherwise it survives docker-compose restarts"

.PHONY: docker
docker:
	(cd docker && docker build . -t mysql-setup:latest)

.PHONY: start
start:
	#sh startdb.sh -d
	docker-compose up -d

.PHONY: install
install:
	../../../dbp apply -e mysql \
	--db.host="localhost" \
	--db.port=3306 \
	--db.name="test_db" \
	--db.username="root" \
	--db.password="password" \
	--db.tls=false \
	-f patches -v -d \
	-r 100

.PHONY: dry
dry:
	../../../dbp apply -e mysql \
	--db.host="localhost" \
	--db.port=3306 \
	--db.name="test_db" \
	--db.username="root" \
	--db.password="password" \
	--db.tls=false \
	-f patches -v -d \
	-r 100 --dry


stop:
	docker-compose down

clean:
	rm -rf data

