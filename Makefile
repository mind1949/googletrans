export GO111MODULE=on
fmt:
	gofmt -w -s .
	go mod tidy
test:
	go test tkk/*
	go test tk/*
	go test transcookie/*
	go test .
bench:
	go test tkk/* -bench=. -run=NONE -benchmem
	go test . -bench=. -run=NONE -benchmem