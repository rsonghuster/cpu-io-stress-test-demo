GIT_COMMIT	:=	$(shell git describe --always 2> /dev/null || true)
TIMESTAMP	:=	$(shell date '+%y%m%d-%H%M%S')
VERSION		:=	$(shell echo "${GIT_COMMIT}-${TIMESTAMP}" )

hangzhou_tag	:=	$(shell echo "registry.cn-hangzhou.aliyuncs.com/fc-demo2/cpu-io-test-demo:${VERSION}")

build: 
	docker build -t ${hangzhou_tag} -f Dockerfile . --no-cache

push:
	docker push ${hangzhou_tag}

all: build push