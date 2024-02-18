BIN_DIR=./bin
TMP_DIR=./tmp
CONFIG_DIR=./deploy/config

gen-mocks:
	go generate

gen-proto:
	protoc --proto_path=./ --go_out=. --go-grpc_out=. \
		./src/eventserver/generated/proto/eventserverpb/*.proto \
		./src/pkg/generated/proto/rteventspb/*.proto

generate-rsa-key-pair:
	mkdir -p $(TMP_DIR)/keys
	openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out $(TMP_DIR)/keys/occa &> /dev/null
	openssl rsa -pubout -in $(TMP_DIR)/keys/occa -out $(TMP_DIR)/keys/occa_pub &> /dev/null

setup-env: clean-env
	docker-compose -p occa -f deploy/local/docker-compose.yml up -d --build

clean-env:
	docker-compose -p occa -f deploy/local/docker-compose.yml down

gateway:
	go build -o ${BIN_DIR}/$@ src/cmd/gateway/main.go

eventserver:
	go build -o ${BIN_DIR}/$@ src/cmd/eventserver/main.go

run-gateway: gateway
	${BIN_DIR}/gateway -f ${CONFIG_DIR}/gateway.yml

run-eventserver: eventserver
	${BIN_DIR}/eventserver -f ${CONFIG_DIR}/event-server.yml

run-eventserver-cluster: eventserver
	SERVER_ID=event_server_1 HTTP_LISTEN_ADDRESS=9101 GRPC_LISTEN_ADDRESS=9102 ${BIN_DIR}/eventserver -f ${CONFIG_DIR}/event-server.yml &
	SERVER_ID=event_server_2 HTTP_LISTEN_ADDRESS=9103 GRPC_LISTEN_ADDRESS=9104 ${BIN_DIR}/eventserver -f ${CONFIG_DIR}/event-server.yml &
	SERVER_ID=event_server_3 HTTP_LISTEN_ADDRESS=9105 GRPC_LISTEN_ADDRESS=9106 ${BIN_DIR}/eventserver -f ${CONFIG_DIR}/event-server.yml

run-cli:
	go run src/cmd/cliclient/main.go