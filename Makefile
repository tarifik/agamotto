test: 
	@go test -race

cover:
	@gopherbadger -md="README.md"

build:
	@echo "building mac os version"
	@go build -o builds/masker_proxy_mac -v cmd/main.go && \
	echo "build SUCCESS" || \
	echo "build FAILED"
	@echo ""

	@echo "building windows version"
	@GOOS=windows GOARCH=386 go build -o builds/masker_proxy_mac.exe -v cmd/main.go && \
	echo "build SUCCESS" || \
	echo "build FAILED"
	@echo ""

	@echo "building linux version"
	@GOOS=linux GOARCH=amd64 go build -o builds/masker_proxy_linux -v cmd/main.go && \
	echo "build SUCCESS" || \
	echo "build FAILED"