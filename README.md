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

An example response could look like this:
```json
{
  "title": "Om Oss",
  "slug": "om-oss",
  "updated_at": "2022-11-06T02:04:58Z",
  "image": "https://example.com/static/hej.png",
  "message": "hej",
  "body": "<h1>...",
  "sidebar": "<ul>...",
  "anchors": [{"id":"id", "value":"asdf"}],
  "nav": [
    {
      "slug": "/om-oss",
      "title": "Om Oss",
      "sort": 1,
      "active": true, 
      "expanded": false
    },
    {
      "slug": "/faq",
      "title": "FAQ",
      "sort": 2
    }
  ]
}
```

## Running 

### Environment variables

| Name         | Description                                                                                                                              |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| PORT         | The port to listen to requests on                                                                                                        |
| TOKEN        | GitHub Personal Access Token used for authorization when pulling the content repository. (Only needed if the content repo is private)    |
| CONTENT_URL  | The repository to get content from                                                                                                       |
| CONTENT_DIR  | Directory to serve contents from. Setting this disables the automatic fetching using git and makes the `TOKEN` and `CONTENT_URL` unused. |
| DARKMODE_URL | URL to the darkmode system, or `true` or `false` to use that value instead of sending an http request.                                   |

### Flags

| Name | Description                                  |
| ---- | -------------------------------------------- |
| -v   | Print info messages                          |
| -vv  | Print more info messages                     |
| -w   | Reload the contents when they change on disk |

### Docker

If you have docker installed, you can also run the repo using `docker-compose up --build`

Make sure to copy `.env.example` to `.env` first, and populate `TOKEN` with you personal github token if needed.

## API documentation

http://godoc.org/github.com/datasektionen/taitan

## Content repo structure and features

Taitan uses path-based routing. I.e. if your content repo contains a directory `foo` with a subdirectory `bar`, you will get the content of that subdir by navigating to `<taitan url>/foo/bar`.

Every directory has to contain a `meta.toml`, a `sidebar.md` and a `body.md` file, the content of which are described below.

Directories starting with a `.` is ignored by taitan. (eg. `.github`).

### meta.toml

The purpose of this file is to provide meta-data that the frontend might or might not need to render a page. 

The `meta.toml` files can contain the following fields:

| Name      | Data type | Mandatory | Description                                                                                                           |
| --------- | --------- | --------- | --------------------------------------------------------------------------------------------------------------------- |
| Title     | string    | yes       | The title of the page                                                                                                 |
| Image     | string    | no        | Link to an image that can be used by the frontend in any way it wants                                                 |
| Message   | string    | no        | Specifies a string that is sent to the frontend to use as it wants                                                    |
| Sort      | int       | no        | A key appearing in the `nav` attribute intended for the frontend to use for the page when sorting navigation menues.  |
| Expanded  | boolean   | no        | Specifies whether all the children of of a an directory should be always be expanded when it is included in the `nav` |
| Sensitive | string    | no        | Weather the whole page should be hidden during reception times.                                                       |

### sidebar.md

A markdown file that will contain content intended to render as a sidebar for a route.

### body.md

This is the file that will contain the html content that will be served for a route. It is written in markdown, and the generated page will be very similar to how the markdown is rendered.

#### Darkmode (hiding info during reception)

`taitan` has support for hiding some content during reception mode by surrounding text with `{{ if .reception -}} {{- else -}} {{- end }} `.

##### Example:

```
some text

{{ if .reception -}}
  nothing here :)
{{- else -}}
  secret text
{{- end}}

more text

```
when `DARMODE_URL` returns `true` (during reception), this will render as:

```
some text

nothing here

more text
```
and if it returns `false` (the rest of the year), it will return
```
some text

secret text

more text
```



## Public domain

I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).

