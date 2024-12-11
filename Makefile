server:
	export PORT=8080 &&\
	export MODE=server &&\
	go run main.go

worker:
	export PORT=8081 &&\
	export MODE=consumer &&\
	go run main.go