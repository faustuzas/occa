BIN_DIR=./bin
CONFIG_DIR=./deploy/config

setup-env:
	docker-compose -p occa -f deploy/local/docker-compose.yml down
	docker-compose -p occa -f deploy/local/docker-compose.yml up -d --build

gateway:
	go build -o ${BIN_DIR}/$@ src/cmd/gateway/main.go

run-gateway: gateway
	${BIN_DIR}/gateway -f ${CONFIG_DIR}/gateway.yml

run-cli:
	go run src/cmd/cliclient/main.go