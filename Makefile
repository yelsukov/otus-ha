PROJECT=otus-habash
GOOS=linux
CGO_ENABLED?=0

include .env

TYPE?=standalone

buildBackend: clean
	cd backend && CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} go build -a \
		-ldflags "-s -w -X main.VERSION=${VERSION}" \
		-o ../../bin/server

buildFrontend: clean
	cd frontend && npm install && ng build --prod

up:
	sudo docker-compose -f docker-compose.${TYPE}.yml up --build -d

down:
	sudo docker-compose -f docker-compose.${TYPE}.yml down -v

upFull:
	sudo docker-compose -f docker-compose.standalone.yml -f docker-compose.news.yml -f docker-compose.infra.yml up --build -d

downFull:
	sudo docker-compose -f docker-compose.standalone.yml -f docker-compose.news.yml -f docker-compose.infra.yml down -v

startReplica:
	sudo docker-compose -f docker-compose.proxysql.yml up --build -d

stopReplica:
	sudo docker-compose -f docker-compose.proxysql.yml down -v

startTaran:
	sudo docker-compose -f docker-compose.tarantool.yml -f docker-compose.infra.yml up --build -d

stopTaran:
	sudo docker-compose -f docker-compose.tarantool.yml -f docker-compose.infra.yml down -v

startMonitor:
	sudo docker-compose -f deployment/monitoring/docker-compose.yml up --build -d

stopMonitor:
	sudo docker-compose -f deployment/monitoring/docker-compose.yml down -v --rmi=local

fmt:
	@echo "+ $@"
	@goimports -w -l src

clean:
	rm -f backend/bin/*
	rm -f frontend/dist/*