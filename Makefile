GROUP_NAME=platform
PROJECT_NAME=arc-consumer
API_ROOT=/data/v1/realtime

.PHONY: image swagger-client swagger-server

image:
	docker build -t${GROUP_NAME}/${PROJECT_NAME}:dev .

run:
	docker run --rm \
	-p 8972:8972 \
	-e BASIC_INSWARM=false \
	-e TAOS_TAOSENABLE=false \
	${GROUP_NAME}/${PROJECT_NAME}:dev

