## WIP

This project is a *work in progress*. The implementation is *incomplete* and
subject to change. The documentation can be inaccurate.

# taitan

*Taitan* is a RESTful markdown to HTML supplier of pages for [datasektionen/bawang](http://github.com/datasektionen/bawang).

*Taitan(タイタン) is romaji for Titan.*

## API

Retrieve a markdown document

GET /:path

## Response

```json
{
  "title": "unimplemented",
  "slug": "unimplemented",
  "updated_at": "2015-11-06T02:04:58Z",

  "image": "unimplemented",

  "body": "<h1>...",
  "sidebar": "<script>...",
  "anchors": [{"id":"id", "value":"asdf"}],
}
```

## API documentation

http://godoc.org/github.com/datasektionen/taitan


## Public domain

I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
