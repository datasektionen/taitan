## WIP

This project is a *work in progress*. The implementation is *incomplete* and
subject to change. The documentation can be inaccurate.

# taitan

[![Build Status](https://travis-ci.org/datasektionen/taitan.svg?branch=master)](https://travis-ci.org/datasektionen/taitan)
[![Coverage Status](https://coveralls.io/repos/datasektionen/taitan/badge.svg?branch=master&service=github)](https://coveralls.io/github/datasektionen/taitan?branch=master)
[![GoDoc](https://godoc.org/github.com/datasektionen/taitan?status.svg)](https://godoc.org/github.com/datasektionen/taitan)

*Taitan* is a RESTful markdown to HTML supplier of pages for [datasektionen/bawang](http://github.com/datasektionen/bawang).

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

## Installation & usage

```bash
$ go get -u github.com/datasektionen/taitan
$ taitan -v site/
INFO[0000] Our root directory                            Root=site/
INFO[0000] Starting server.                             
INFO[0000] Listening on port: 4000
...
```

## API documentation

http://godoc.org/github.com/datasektionen/taitan  
http://godoc.org/github.com/datasektionen/taitan/parse

## Public domain

I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
