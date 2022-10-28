AWS_PROFILE=personal

BUILDIR=bin
API_CMD_DIR=cmd/api
DEPLOYMENT_DIR=deployments

CONFIG_FILE=configs/config.yaml

TARGET=optimized-m3u-iptv-list-server

$(BUILDIR):
	mkdir -p $(BUILDIR)

$(TARGET): $(BUILDIR)
	env GOOS=linux GOARCH=amd64 go build -o $(BUILDIR)/$(TARGET) $(API_CMD_DIR)

.PHONY: build
build: $(TARGET)
	cp $(CONFIG_FILE) $(BUILDIR)/

.PHONY: deploy
deploy: build
	cd $(DEPLOYMENT_DIR) && AWS_PROFILE=$(AWS_PROFILE) terraform apply -var-file="secret.tfvars"

.PHONY: clean
clean:
	rm -rf $(BUILDIR)