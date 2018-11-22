SHELL = /bin/bash

IMAGE_NAME = az82/showcase-webhook
NAMESPACE = admissioncontrol

CA_SUBJ = "/C=DE/O=az82/OU=webhook-showcase/CN=ca"
CERT_SUBJ = "/C=DE/O=az82/OU=webhook-showcase/CN=webhook.$(NAMESPACE).svc"
VALID_DAYS = 365
EC_CURVE = "secp384r1"
CERTS_DIR = ".certs"

CA_KEYFILE = "$(CERTS_DIR)/ca-key.pem"
CA_CERTFILE = "$(CERTS_DIR)/ca-cert.pem"
KEYFILE = "$(CERTS_DIR)/tls.key"
CERTFILE = "$(CERTS_DIR)/tls.crt"
CSR_FILE = "$(CERTS_DIR)/csr.pem"


.PHONY: deploy
deploy: .certs container
	@echo -e "\nCreating namespace..."
	 kubectl create namespace $(NAMESPACE) --dry-run -o yaml | kubectl apply -f -

	@echo -e "\nCreating secret..."
	kubectl create secret tls -n $(NAMESPACE) webhook-tls --cert="$(CERTFILE)" --key="$(KEYFILE)" --dry-run -o yaml | kubectl apply -f -

	@echo -e "\nDeploying webhook..."
	sed -E "s/(caBundle:).*\$$/\1 $$(base64 "$(CA_CERTFILE)" | tr -d '\r\n')/" webhook.yaml | kubectl apply -n $(NAMESPACE) -f -


.PHONY: container
container: main-linux-amd64
	@echo -e "\nBuilding docker image..."
	docker build -t $(IMAGE_NAME) .


main: main.go
	@echo -e "\nCompiling application..."
	go get -d ./...
	go test
	go build -o main .


main-linux-amd64: main.go
	@echo -e "\nCompiling application..."
	go get -d ./...
	go test
	env GOOS=linux GOARCH=amd64 go build -o main-linux-amd64 .


.certs:
	mkdir -p $(CERTS_DIR)

	@echo -e "\nGenerating CA..."
	openssl req -new -newkey ec:<(openssl ecparam -name $(EC_CURVE)) -days $(VALID_DAYS) -nodes -x509 -subj "$(CA_SUBJ)" -keyout "$(CA_KEYFILE)" -out "$(CA_CERTFILE)"

	@echo -e "\nGenerating certificate..."
	openssl ecparam -name $(EC_CURVE) -genkey -noout -out "$(KEYFILE)"
	openssl req -new -newkey ec:<(openssl ecparam -name $(EC_CURVE)) -subj "$(CERT_SUBJ)" -key "$(KEYFILE)" -out "$(CSR_FILE)"
	openssl x509 -req -in "$(CSR_FILE)" -CA "$(CA_CERTFILE)" -CAkey "$(CA_KEYFILE)" -CAcreateserial -out "$(CERTFILE)" -days "$(VALID_DAYS)"


.PHONY: clean
clean:
	rm -rf $(CERTS_DIR) .srl main main-linux-amd64


.PHONY: clean-kubernetes
clean-kubernetes:
	@echo -e "\nDeleting docker image..."
	docker rmi -f $(IMAGE_NAME) 2> /dev/null

	@echo -e "\nDeleting webhook deployments..."
	kubectl delete namespace $(NAMESPACE)
