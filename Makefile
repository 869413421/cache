.PHONY: proto
proto:
	protoc --go_out=. ./proto/*.proto
