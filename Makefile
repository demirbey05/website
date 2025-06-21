
.PHONY: build expo

build:
	docker buildx build --platform linux/amd64 -t website --load .
expo:
	docker save website -o website.tar