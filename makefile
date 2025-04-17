
server: web/app/*
	go build -o server ./cmd/server

problems:
	go run cmd/problem_importer testdata/data/leetcode_dataset.csv database.db

test:
	go test ./...

test_short:
	go test -short ./...

test_v:
	go test -v ./...
