# Mindav
![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/totoval/mindav.svg)
![GitHub last commit](https://img.shields.io/github/last-commit/totoval/mindav.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/totoval/mindav)](https://goreportcard.com/report/github.com/totoval/mindav)
![Travis (.org)](https://img.shields.io/travis/totoval/mindav.svg)
![GitHub top language](https://img.shields.io/github/languages/top/totoval/mindav.svg)
![GitHub](https://img.shields.io/github/license/totoval/mindav.svg)

## About Mindav
Mindav is a webdav server which is supported multi file backends such as minio, memory and file.

## Getting Started
> Assumed that you already have your Minio server running. Or [Quick Run Minio Server](#quick-run-minio-server) 
* `cp .env.example.json .env.json`
* Config your Minio in your `.env.json` file
    ```json
    {
      "WEBDAV_DRIVER": "minio",
      "MINIO_ENDPOINT": "play.min.io:9000",
      "MINIO_ACCESS_KEY_ID": "access_key_id",
      "MINIO_SECRET_ACCESS_KEY": "secret_access_key",
      "MINIO_BUCKET": "bucket_name",
      "MINIO_USE_SSL": false
    }
    ```
* Run `go run main.go` or the run the binary
* Now you can connect the Mindav by using webdav clients, such as [Cyberduck](http://cyberduck.io):  
<img src="https://raw.githubusercontent.com/totoval/mindav/master/readme_assets/D133AD6B-755F-4878-826F-FC9A6A0BA273.png" alt="cyberduck client" width="300" />

## Supported Clients(KNOWN):   
* [Cyberduck](http://cyberduck.io) for `OSX`  
* [PhotoSync](http://www.photosync-app.com) for `iOS`
* and More...
> `OSX`'s `finder` is not support for `rename` operate!

## Quick Run Minio Server
```sh
docker run --name minio --rm -it \ 
    -p "9000:9000" \ 
    -v "./minio/data:/data" \ 
    -v "./minio/config:/root/.minio" \ 
    minio/minio:latest \ 
    server /data
```

## Roadmap
- [x] Memory filesystem support
- [x] File filesystem support
- [x] Minio filesystem support
- [ ] User system

## Thanks
* [Totoval](https://github.com/totoval/totoval)