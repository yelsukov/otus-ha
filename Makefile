PROJECT=otus-ha
GOOS=linux
CGO_ENABLED?=0

include .env

buildBackend: clean
	cd backend && CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} go build -a \
		-ldflags "-s -w -X main.VERSION=${VERSION}" \
		-o ../../bin/server

buildFrontend: clean
	cd frontend && npm install && ng build --prod

up:
	sudo docker-compose -f docker-compose.yml up --build -d

down:
	sudo docker-compose -f docker-compose.yml down

fmt:
	@echo "+ $@"
	@goimports -w -l src

clean:
	rm -f backend/bin/*
	rm -f frontend/dist/*