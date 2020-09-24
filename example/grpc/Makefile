.PHONY: compile

compile:
	@protoc --go_out=plugins=grpc:./service --go_opt=paths=source_relative -I ./service/ ./service/*.proto
