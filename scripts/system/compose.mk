APP_PATH?=/go/src/$(shell grep ^module go.mod | awk '{print $$2}')
APP_PROJECT=${APP_NAME}_${MODE}

DOCKER_COMPOSE_BUILD_ARGS=\
	GOPATH="${GOPATH}" \
	APP_NAME="${APP_NAME}" \
	APP_PATH="${APP_PATH}" \
	APP_PROJECT="${APP_PROJECT}" \
	MODE="${MODE}"

DOCKER_COMPOSE_DIR=${PWD}/build/docker

DOCKER_COMPOSE_CMD=\
	${DOCKER_COMPOSE_BUILD_ARGS} docker-compose \
		-p ${APP_PROJECT} \
		-f ${DOCKER_COMPOSE_DIR}/docker-compose.yaml

RED_COLOR=\033[0;31m
NO_COLOR=\033[0m


docker-test-up:	## Up compose test docker images
docker-test-up: .mode-test
	${DOCKER_COMPOSE_CMD} up

docker-test-build:	## Build compose test docker images
docker-test-build: .mode-test
	${DOCKER_COMPOSE_CMD} build

docker-test:	## Run compose test
docker-test: cmd?=test
docker-test: .mode-test
	${DOCKER_COMPOSE_CMD} exec app make ${cmd}

docker-test-down:	## Stop docker-compose to test and remove db
docker-test-down: .mode-test
	${DOCKER_COMPOSE_CMD} down -v

docker-dev-up:	## Up compose development docker images
docker-dev-up: .mode-dev
	${DOCKER_COMPOSE_CMD} up

docker-dev-build:	## Build compose development docker images
docker-dev-build: .mode-dev
	${DOCKER_COMPOSE_CMD} build

docker-dev-down:	## Stop development mode docker-compose
docker-dev-down: .mode-dev
	${DOCKER_COMPOSE_CMD} down -v

docker-dev-reup:	## Down and up compose development docker images
docker-dev-reup: docker-dev-down docker-dev-up

docker-dev:	## Run compose in development mode with live reload
docker-dev: .mode-dev
	${DOCKER_COMPOSE_CMD} run --rm --name ${MODE}_run_app app make ${cmd}

.mode-test:
	$(eval MODE=test)

.mode-dev:
	$(eval MODE=dev)

kafka-users:	## Make kafka users
	docker exec -it protokaf_dev_kafka /opt/kafka/bin/kafka-configs.sh \
		--bootstrap-server 0.0.0.0:9092 --alter \
  		--add-config 'SCRAM-SHA-256=[password=secret],SCRAM-SHA-512=[password=secret]' \
  		--entity-type users --entity-name admin
