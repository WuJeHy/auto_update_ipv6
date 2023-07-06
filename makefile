
APP_NAME      := ipv6_ddns_$(shell date "+%Y%m%d%H" )

build_ipv6_app:
		GOOS=linux GOARCH=amd64 go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH -ldflags '-w -s -linkmode "external" -extldflags "-static"' -o build/${APP_NAME}-amd64_linux main.go
build_ipv6_app_ubuntu:
		GOOS=linux GOARCH=amd64 go build     -ldflags '-w -s ' -o build/${APP_NAME}-amd64_ubuntu main.go
