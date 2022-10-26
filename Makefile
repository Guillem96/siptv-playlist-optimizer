AWS_PROFILE=personal
BUILDIR=bin
SOURCEDIR=src
CONFIG_FILE=config.yaml
TARGET=optimized-m3u-iptv-list-server

$(BUILDIR):
	mkdir -p $(BUILDIR)

$(TARGET): $(BUILDIR)
	cd $(SOURCEDIR) && env GOOS=linux GOARCH=amd64 go build -o ../$(BUILDIR)/$(TARGET)

.PHONY: build
build: $(TARGET)
	cp $(CONFIG_FILE) bin/$(CONFIG_FILE)

.PHONY: deploy
deploy: build
	cd infrastructure && AWS_PROFILE=$(AWS_PROFILE) terraform apply -var-file="secret.tfvars"

.PHONY: clean
clean:
	rm -rf $(BUILDIR)