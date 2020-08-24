export GO111MODULE=on
fmt:
	gofmt -w -s .
	go mod tidy
test:
	go test tkk/* -v
	go test tk/* -v
	go test . -v
bench:
	go test tkk/* -bench=. -run=NONE