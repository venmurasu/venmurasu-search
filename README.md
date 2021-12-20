# Venmurasu-search
** Beta **

A sample application showing Venmurasu Search functionality.

## Dependencies
Uses Blevesearch library

## Installation


Install Go using [the instructions](https://golang.org/doc/install). This application requires Go 1.14 or above.

After installing Go, Then clone this repo.

```
git config --global submodule.recurse true
git clone git@github.com:venmurasu/venmurasu-search.git
cd venmurasu-search
git submodule update --init --recursive
```

### To generate bleve index

To generate bleve index file from the venmurasu json directory.

```
make genindex
```

This generated index directory is required to run the search web server.

### To build binary

```
make
```

Go supports cross compilation, if you want to build executable for linux from Mac or windows, then

```
env GOOS=linux GOARCH=amd64 go build -o bin/server-amd64 *.go
```

### To run the server

start the app from

```
make run
```
Before this, build the index. Then the app is accessible from `http://localhost:8094`

### To run as docker

```
docker-compose up -d
```

docker expectes mounting the index directory.


## REST API

The application has REST API for integration with other applications.

```
POST /api/search
Request JSON: { "size": 10 , "from":0, "search": "search:'இளைய யாதவர்'"}
```
Where `size` -> no. of results
       `from` -> Offset, used for pagination, for next page, offset by the size * no. of pages (e.g., For page 2, from -> 20)
