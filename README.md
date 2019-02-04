#Wiki documentation

## Docker

### build
docker build -t gowiki .

### run
docker run --name gowiki -p 8080:8080 gowiki

### stop
docker stop gowiki

### remove
docker rm gowiki

## Go

### run (development)
go run wiki.go

### build
go build wiki.go

### run build (production)
./wiki