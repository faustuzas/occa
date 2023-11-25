BIN_DIR=./bin
CONFIG_DIR=./deploy/config

setup-env:
	docker-compose -f deploy/local/docker-compose.yaml up -d

gateway:
	go build -o ${BIN_DIR}/$@ src/cmd/gateway/main.go

gateway-run: gateway
	${BIN_DIR}/gateway -f ${CONFIG_DIR}/gateway.yaml

start-cli:
	go run src/cmd/cliclient/main.go