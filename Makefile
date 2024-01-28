BIN_DIR=./bin
TMP_DIR=./tmp
CONFIG_DIR=./deploy/config

gen-mocks:
	go generate 

generate-rsa-key-pair:
	mkdir -p $(TMP_DIR)/keys
	openssl genpkey -algorithm RSA -out $(TMP_DIR)/keys/occa &> /dev/null
	openssl rsa -pubout -in $(TMP_DIR)/keys/occa -out $(TMP_DIR)/keys/occa_pub &> /dev/null

setup-env: clean-env
	docker-compose -p occa -f deploy/local/docker-compose.yml up -d --build

clean-env:
	docker-compose -p occa -f deploy/local/docker-compose.yml down

gateway:
	go build -o ${BIN_DIR}/$@ src/cmd/gateway/main.go

run-gateway: gateway
	${BIN_DIR}/gateway -f ${CONFIG_DIR}/gateway.yml

run-cli:
	go run src/cmd/cliclient/main.go