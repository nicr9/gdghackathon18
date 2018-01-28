build:
	docker build -t nicr9/beacon .

push:
	docker push nicr9/beacon

.PHONY: build push
