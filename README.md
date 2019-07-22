# MinDAV
![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/totoval/mindav.svg)
![GitHub last commit](https://img.shields.io/github/last-commit/totoval/mindav.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/totoval/mindav)](https://goreportcard.com/report/github.com/totoval/mindav)
![Travis (.org)](https://img.shields.io/travis/totoval/mindav.svg)
![GitHub top language](https://img.shields.io/github/languages/top/totoval/mindav.svg)
![GitHub](https://img.shields.io/github/license/totoval/mindav.svg)
![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/totoval/mindav.svg)

## About MinDAV
MinDAV is a self-hosted file backup server which bridges WebDAV protocol with Minio.  
  
**WebDAV ❤️ Minio**

## Why them?
### WebDAV

> Web Distributed Authoring and Versioning (WebDAV) is an extension of the Hypertext Transfer Protocol (HTTP) that allows clients to perform remote Web content authoring operations.   

There're many cloud storages that support WebDAV protocol, such as **dropbox**, **owncloud**, **nextcloud**, etc.   
  
***WebDAV provides a simple port for your files.***

### Minio
> The 100% Open Source, Enterprise-Grade, Amazon S3 Compatible Object Storage  
  
***Minio is [reliable](https://docs.min.io/docs/minio-erasure-code-quickstart-guide.html) for your files.***

## Architecture

<img src="https://raw.githubusercontent.com/totoval/mindav/master/readme_assets/architecture.png" alt="mindav architecture" width="800" />

## One Click Start
```bash
git clone git@github.com:totoval/mindav.git
cd mindav
docker-compose up -d
```
Now you can connect the MinDAV by using your favorite WebDAV clients, such as [Cyberduck](http://cyberduck.io):  
<img src="https://raw.githubusercontent.com/totoval/mindav/master/readme_assets/37E56D20-FCA7-41FB-B8B2-3B5E390A6DBC.png" alt="cyberduck client" width="600" />

## Getting Started
> Assumed that you already have your [Minio](https://github.com/minio/minio) server running. Or [Quick Run Minio Server](#quick-run-minio-server) 
* `cp .env.example.json .env.json`
* Config your Minio in your `.env.json` file
    ```json
    {
      "WEBDAV_DRIVER": "minio",
      "WEBDAV_USER": "totoval",
      "WEBDAV_PASSWORD": "passw0rd",
      "MINIO_ENDPOINT": "play.min.io:9000",
      "MINIO_ACCESS_KEY_ID": "access_key_id",
      "MINIO_SECRET_ACCESS_KEY": "secret_access_key",
      "MINIO_BUCKET": "bucket_name",
      "MINIO_USE_SSL": false,
      "MEMORY_UPLOAD_MODE": false
    }
    ```
* Run `go run main.go` or the run the binary
* Now you can connect the MinDAV by using your favorite WebDAV clients

## Quick Run Minio Server
```sh
docker run --name minio --rm -it \ 
    -p "9000:9000" \ 
    -v "./minio/data:/data" \ 
    -v "./minio/config:/root/.minio" \ 
    minio/minio:latest \ 
    server /data
```

## Supported Clients(KNOWN):   
* [Cyberduck](http://cyberduck.io) for `macOS`, `Windows`  
* [PhotoSync](http://www.photosync-app.com) for `iOS`, `Android`  
* [FE File Explorer](http://www.skyjos.com) for `iOS`, `Android`, `macOS`  
* [Filezilla](https://filezilla-project.org/) for `macOS`, `Windows`, `Linux`
* And More...
> `OSX`'s `finder` is not support for `rename` operate!

## MEMORY_UPLOAD_MODE
> If the host has a large memory, then set to `true` could improve upload performance.

## Roadmap
- [x] Memory filesystem support
- [x] File filesystem support
- [x] Minio filesystem support
- [x] User system

## Thanks
* [Totoval](https://github.com/totoval/totoval)
