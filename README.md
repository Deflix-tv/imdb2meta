# imdb2meta

A service for getting movie and TV show metadata for an IMDb ID via HTTP or gRPC, using the official IMDb datasets

## Content

- [Content](#content)
- [Usage](#usage)
  1. [Import data](#1-import-data)
  2. [Run service](#2-run-service)
  3. [Query service](#3-query-service)
- [Protocol buffer generation](#protocol-buffer-generation)
- [⚠ Warning](#⚠-warning)

## Usage

First you need import the data of the IMDb dataset into a database, then you need to start the web service which is backed by the database and finally you can query it via HTTP or gRPC.

### 1. Import data

First you need import the data of the IMDb dataset into a database. We support [BadgerDB](https://github.com/dgraph-io/badger) and [bbolt](https://github.com/etcd-io/bbolt).

Steps:

1. Download the `title.basics.tsv.gz` dataset from <https://datasets.imdbws.com>
   - For more info about IMDb datasets see <https://www.imdb.com/interfaces/>
   - > ⚠ Warning: `IMDb.com, Inc` is the copyright owner of the data in the IMDb datasets. You may only use the data for personal and non-commercial use. For more info see ["Can I use IMDb data in my software?"](https://help.imdb.com/article/imdb/general-information/can-i-use-imdb-data-in-my-software/G5JTRESSHJBBHTGX) and their [copyright/conditions of use](https://www.imdb.com/conditions) statement.

2. Exract the TSV file somewhere
3. Run the import tool with the appropriate CLI arguments
   - Example: `imdb2meta-import -tsvPath "/home/john/Downloads/data.tsv" -badgerPath "/home/john/imdb2meta/badger"`

> Note: The import takes a while (and much longer with bbolt than with BadgerDB), the process requires a lot of memory and the final DB size is fairly big.  
> With a 6-core, 12-thread CPU and a mid-range SSD, an import of all data (7351639 rows as of 2020-11-21) into BadgerDB takes 4 minutes, up to 1.03 GB memory and the final DB size is 1.29 GB.  
> When skipping TV episodes and storing only the minimal metadata it takes 1 minute and 5 seconds, up to 530 MB memory and the final DB size is 314 MB.

CLI reference:

```text
Usage of imdb2meta-import:
  -badgerPath string
        Path to the directory with the BadgerDB files
  -boltPath string
        Path to the bbolt DB file
  -limit int
        Limit the number of rows to process (excluding the header row)
  -minimal
        Only store minimal metadata (ID, type, title, release/start year)
  -skipEpisodes
        Skip storing individual TV episodes
  -tsvPath string
        Path to the "data.tsv" file that's inside the "title.basics.tsv.gz" archive
```

### 2. Run service

After importing the data you can start the web service.

Example: `imdb2meta-service -badgerPath "/home/john/imdb2meta/badger"`

CLI reference:

```text
Usage of imdb2meta-service:
  -badgerPath string
        Path to the directory with the BadgerDB files
  -bindAddr string
        Local interface address to bind to. "localhost" only allows access from the local host. "0.0.0.0" binds to all network interfaces. (default "localhost")
  -boltPath string
        Path to the bbolt DB file
  -grpcPort int
        Port to listen on for gRPC requests (default 8081)
  -httpPort int
        Port to listen on for HTTP requests (default 8080)
```

#### Docker

You can also run the service as Docker container.

1. Update the image: `docker pull doingodswork/imdb2meta-service`
2. Start the container: `docker run --name imdb2meta -v /path/to/badger:/data -p 8080:8080 -p 8081:8081 doingodswork/imdb2meta-service -badgerPath "/data"`
   - > Note: `Ctrl-C` only detaches from the container. It doesn't stop it.
   - When detached, you can attach again with `docker attach imdb2meta`
3. To stop the container: `docker stop imdb2meta`
4. To start the (still existing) container again: `docker start imdb2meta`

### 3. Query service

After starting the web service you can query it via HTTP or gRPC:

#### HTTP

Example request: `curl "http://localhost:8080/meta/tt1254207"`

Example response:

```json
{
    "id": "tt1254207",
    "titleType": "SHORT",
    "primaryTitle": "Big Buck Bunny",
    "startYear": 2008,
    "runtime": 10,
    "genres": [
        "Animation",
        "Comedy",
        "Short"
    ]
}
```

#### gRPC

Example request (using [grpcurl](https://github.com/fullstorydev/grpcurl)): `grpcurl -plaintext -d '{"id":"tt1254207"}' localhost:8081 imdb2meta.MetaFetcher/Get`  
(In Windows/PowerShell you have to use `'{\"id\":\"tt1254207\"}'`)

Example response:

```json
{
    "id": "tt1254207",
    "titleType": "SHORT",
    "primaryTitle": "Big Buck Bunny",
    "startYear": 2008,
    "runtime": 10,
    "genres": [
        "Animation",
        "Comedy",
        "Short"
    ]
}
```

## Protocol buffer generation

To re-generate the `meta.pb.go` file from the `meta.proto` file, run: `protoc -I="./protos" --go_out=./pb --go_opt=paths=source_relative meta.proto`

To re-generate the `service.pb.go` and `service_grpc.pb.go` files from the `service.proto` file, run: `protoc -I="./protos" --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative service.proto`

## ⚠ Warning

`IMDb.com, Inc` is the copyright owner of the data in the IMDb datasets. You may only use the data for personal and non-commercial use. For more info see ["Can I use IMDb data in my software?"](https://help.imdb.com/article/imdb/general-information/can-i-use-imdb-data-in-my-software/G5JTRESSHJBBHTGX) and their [copyright/conditions of use](https://www.imdb.com/conditions) statement.
