export GO111MODULE=on
fmt:
	gofmt -w -s .
	go mod tidy
test:
	go test tkk/*
	go test tk/*
bench:
	go test tkk/* -bench=. -run=NONE