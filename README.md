# Venmurasu-search
** Alpha Status **

A sample application showing Venmurasu Search functionality.

## Dependencies
Uses Blevesearch library

## Installation


Install Go using [the instructions](https://golang.org/doc/install). This application requires Go 1.13 or above.

After installing Go, run the following commands to download and install:

```shell
go get github.com/venmurasu/venmurasu-search

go mod init

```
then
```
go build
```
In the `main.go` code, point the JSON files from https://github.com/venmurasu/venmurasu-source/tree/search-source/content/bleve_data

start the app from
```
./venmurasu-search
```
First run should build the index and start the app and is accessible from `http://localhost:8094`

## REST API

The application has REST API for integration with other applications.

```
POST /api/search  
Request JSON: { "size": 10 , "from":0, "search": "search:'இளைய யாதவர்'"}
```
Where `size` -> no. of results
       `from` -> Offset, used for pagination, for next page, offset by the size * no. of pages (e.g., For page 2, from -> 20)
