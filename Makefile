build: clean
	go build -o bin/server *.go

clean:
	go clean
	rm -rf bin

run: build
	./bin/server -addr 0.0.0.0:8084 -index ${PWD}/vensearch.bleve/ -static ${PWD}/static/

cleanindex:
	rm -rf vensearch.bleve

genindex: cleanindex
	go run generator/*.go -index ${PWD}/vensearch.bleve -jsonDir ${PWD}/data/venmurasu-json/
	chmod -R 755 ${PWD}/vensearch.bleve

docker:
	docker-compose build search

all: build