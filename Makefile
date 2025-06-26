
.PHONY: build expo

build:
	docker buildx build --platform linux/amd64 -t website --load . --push