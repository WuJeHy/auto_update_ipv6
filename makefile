
APP_NAME      := ipv6_ddns_$(shell date "+%Y%m%d%H" )
APP_TIME	  := $(shell date "+%Y%m%d%H" )
build_ipv6_app:
		GOOS=linux GOARCH=amd64 go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH -ldflags '-w -s -linkmode "external" -extldflags "-static"' -o build/${APP_NAME}-amd64_linux main.go
build_ipv6_app_ubuntu:
		GOOS=linux GOARCH=amd64 go build     -ldflags '-w -s ' -o build/${APP_NAME}-amd64_ubuntu main.go



build_ipv6_mgr_ubuntu:
		GOOS=linux GOARCH=amd64 go build -o build/mgr_${APP_TIME}-amd64_ubuntu cmd/run_mgr/main.go
		strip build/mgr_${APP_TIME}-amd64_ubuntu
