# taitan

*Taitan* is a RESTful markdown to HTML supplier of pages for [bawang](http://github.com/datasektionen/bawang) and [styrdokument-bawang](https://github.com/datasektionen/styrdokument-bawang).

*Taitan(タイタン) is romaji for Titan.*

## API

Retrieve a markdown document.

GET /:path

### Response

An example response to `GET /om-oss` could look like this:
```json
{
  "title": "Om Oss",
  "slug": "om-oss",
  "url": "/om-oss",
  "updated_at": "2022-11-06T02:04:58Z",
  "image": "https://example.com/static/hej.png",
  "message": "hej",
  "body": "<h1>...",
  "sidebar": "<ul>...",
  "sort": 1,
  "expanded": false,
  "anchors": [
    {"id": "id", "value": "asdf", "level": 1},
    {"id": "foo-baz", "value": "foo baz", "level": 2}
  ],
  "nav": [
    {
      "slug": "/om-oss",
      "title": "Om Oss",
      "sort": 1,
      "active": true, 
    },
    {
      "slug": "/faq",
      "title": "FAQ",
      "sort": 2
    }
    {
      "slug": "/foo",
      "title": "joke",
      "sort": 3,
      "expanded": true,
      "nav": [{
        "slug": "/foo/bar",
        "title": "baz",
        "sort": null
      }]
    }
  ]
}
```

Some notable behaviours:

* A `nav` item with `expanded` set to true is equivalent to that item containing a nested `nav`.
* If the main `url` parameter is a nested path, that path will always appear in the `nav`-tree with `active` set to `true`, and with all its ancestor `nav`-nodes having `expanded` set to `true`.


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

If you have docker installed, you can also run the repo using `docker compose up --build`

Make sure to copy `.env.example` to `.env` first, and populate `TOKEN` with you personal github token if needed (if your content repo is private).

Note that `CONTENT_DIR` will not work with the current `compose.yml` file.

### Not Docker

**Requires**: `golang`, `git` 

Minimal setup:

1. Run `go mod download`
2. Set relevant env-variables
3. run `go run .`

## Webhooks

`taitan` has two webhooks intended to keep it's content updated.

* Any request with the header `X-Github-Event` set to `push` will cause `taitan` to refetch the content-repo. Meant to be called from a workflow in the content repo that is run on new commits.
* Any request with the header `X-Darkmode-Event` set to `updated` will cause `taitan` to refetch the darkmode status from `DARKMODE_URL`.

## Content repo structure

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
| Sensitive | string    | no        | Weather the whole page should be hidden when `DARMODE_URL` returns true.                                              |

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
  nothing to see here :)
{{- else -}}
  secret text
{{- end }}

more text

```
When `DARMODE_URL` returns `true` (during reception), this will render as:

```
some text

nothing to see here :)

more text
```
If it returns `false` (the rest of the year), it will render as:
```
some text

secret text

more text
```

## Public domain

I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).

