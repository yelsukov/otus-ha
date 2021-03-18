PROJECT=otus-habash
GOOS=linux
CGO_ENABLED?=0

include .env

TYPE?=standalone

SEED_QTY?=1000
SEED_HOST?=127.0.0.1:3336
SEED_USER?=root
SEED_PASS?=rO0t3zRtga

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
	sudo docker-compose -f docker-compose.news.yml -f docker-compose.dialogue.yml -f docker-compose.standalone.yml -f docker-compose.queue.yml up --build -d

downFull:
	sudo docker-compose -f docker-compose.news.yml -f docker-compose.dialogue.yml -f docker-compose.standalone.yml -f docker-compose.queue.yml down -v

startReplica:
	sudo docker-compose -f docker-compose.proxysql.yml up --build -d

stopReplica:
	sudo docker-compose -f docker-compose.proxysql.yml down -v

startTaran:
	sudo docker-compose -f docker-compose.tarantool.yml -f docker-compose.queue.yml up --build -d

stopTaran:
	sudo docker-compose -f docker-compose.tarantool.yml -f docker-compose.queue.yml down -v

upCluster:
	sudo docker-compose -f docker-compose.cluster.yml -f docker-compose.haproxy.yml -f docker-compose.queue.yml up --build -d

downCluster:
	sudo docker-compose -f docker-compose.cluster.yml -f docker-compose.haproxy.yml -f docker-compose.queue.yml down -v

startMonitor:
	sudo docker-compose -f deployment/monitoring/docker-compose.yml up --build -d

stopMonitor:
	sudo docker-compose -f deployment/monitoring/docker-compose.yml down -v --rmi=local

upWithConsul:
	sudo docker-compose -f docker-compose.dialogue.yml -f docker-compose.standalone.yml -f docker-compose.queue.yml up --build -d

downWithConsul:
	sudo docker-compose -f docker-compose.dialogue.yml -f docker-compose.standalone.yml -f docker-compose.queue.yml down -v

fmt:
	@echo "+ $@"
	@goimports -w -l src

seed:
	 cd seeder && go build && ./seeder -dbHost "${SEED_HOST}" -u ${SEED_USER} -p "${SEED_PASS}" -q ${SEED_QTY}
