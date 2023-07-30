# taitan

[![Build Status](https://travis-ci.org/datasektionen/taitan.svg?branch=master)](https://travis-ci.org/datasektionen/taitan)
[![Coverage Status](https://coveralls.io/repos/datasektionen/taitan/badge.svg?branch=master&service=github)](https://coveralls.io/github/datasektionen/taitan?branch=master)
[![GoDoc](https://godoc.org/github.com/datasektionen/taitan?status.svg)](https://godoc.org/github.com/datasektionen/taitan)
[![Go Report Card](http://goreportcard.com/badge/datasektionen/taitan)](http://goreportcard.com/report/datasektionen/taitan)

*Taitan* is a RESTful markdown to HTML supplier of pages for [bawang](http://github.com/datasektionen/bawang).

*Taitan(タイタン) is romaji for Titan.*

## API

Retrieve a markdown document.

GET /:path

## Response

```json
{
  "title": "Om Oss",
  "slug": "om-oss",
  "updated_at": "2015-11-06T02:04:58Z",

  "image": "unimplemented",

  "body": "<h1>...",
  "sidebar": "<ul>...",
  "anchors": [{"id":"id", "value":"asdf"}],
}
```

## Running

### Environment variables

| Name        | Description                                                                             |
|-------------|-----------------------------------------------------------------------------------------|
| PORT        | The port to listen to requests on                                                       |
| TOKEN       | GitHub Personal Access Token used for authorization when pulling the content repository |
| CONTENT_URL | The repository to get content from                                                      |

### Flags

| Name | Description                                  |
|------|----------------------------------------------|
| -v   | Print info messages                          |
| -vv  | Print more info messages                     |
| -w   | Reload the contents when they change on disk |
| -c   | Directory to serve contents from. Setting this disables the automatic fetching using git and makes the `TOKEN` and `CONTENT_URL` unused. |

## API documentation

http://godoc.org/github.com/datasektionen/taitan  
http://godoc.org/github.com/datasektionen/taitan/parse

## Public domain

I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
