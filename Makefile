
.PHONY: build expo

build:
	docker buildx build --platform linux/amd64 -t demirbey05/blog_detorch:latest --load . --push