# imdb2meta

A service for getting movie and TV show metadata for an IMDb ID via HTTP or gRPC

## Protocol buffer generation

To re-generate the `meta.pg.go` file from the `meta.proto` file, run: `protoc --go_out=. --go_opt=paths=source_relative meta.proto`

## ⚠ Warning

`IMDb.com, Inc` is the copyright owner of the data in the IMDb datasets. You may only use the data for personal and non-commercial use. For more info see ["Can I use IMDb data in my software?"](https://help.imdb.com/article/imdb/general-information/can-i-use-imdb-data-in-my-software/G5JTRESSHJBBHTGX) and their [copyright/conditions of use](https://www.imdb.com/conditions) statement.
