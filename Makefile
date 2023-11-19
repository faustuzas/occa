BIN_DIR=./bin
CONFIG_DIR=./deploy/config

gateway:
	go build -o ${BIN_DIR}/$@ src/cmd/gateway/main.go

gateway-run: gateway
	${BIN_DIR}/gateway -f ${CONFIG_DIR}/gateway.yaml