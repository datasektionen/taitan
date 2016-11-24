## WIP

This project is a *work in progress*. The implementation is *incomplete* and
subject to change. The documentation can be inaccurate.

# taitan

[![Build Status](https://travis-ci.org/datasektionen/taitan.svg?branch=master)](https://travis-ci.org/datasektionen/taitan)
[![Coverage Status](https://coveralls.io/repos/datasektionen/taitan/badge.svg?branch=master&service=github)](https://coveralls.io/github/datasektionen/taitan?branch=master)
[![GoDoc](https://godoc.org/github.com/datasektionen/taitan?status.svg)](https://godoc.org/github.com/datasektionen/taitan)
[![Go Report Card](http://goreportcard.com/badge/datasektionen/taitan)](http://goreportcard.com/report/datasektionen/taitan)

*Taitan* is a RESTful markdown to HTML supplier of pages for [datasektionen/bawang](http://github.com/datasektionen/bawang).

*Taitan(タイタン) is romaji for Titan.*

## API

Retrieve a markdown document.

GET /:path

## Response

```json
{
    "title": "Konglig Datasektionen",
    "slug": "/",
    "url": "/",
    "updated_at": "2016-09-06T09:36:58Z",
    "image": "https://stacken.är.den.verkligen.uppe/a.png",
    "body": "<h3>\nDatasektionen är en ideell studentsektion under Tekniska Högskolans Studentkår\nsom finns till för att alla studenter som läser datateknik på KTH ska få\nen så bra och givande studietid som möjligt.\n</h3>\n\n<p>På Konglig Datasektionen finns det många sätt att roa sig.\nFörutom studier i intressanta ämnen och episka fester anordnas det även qulturella\ntillställningar, hackerkvällar, sektionsmöten och mycket mer.</p>\n",
    "sidebar": "<h2 id=\"studier\">Studier</h2>\n\n<p>Datateknikprogrammet på KTH är bland de främsta datateknikutbildningarna i världen. Efter examen har du breda karriärmöjligheter inom branscher där datasystem är viktiga för verksamheten, exempelvis kultur, finans, handel, vård samt industri. Arbeta med design och produktutveckling, undervisning eller konsultverksamhet.</p>\n\n<p><a href=\"/studier\" class=\"action\">Mer om utbildningen &raquo;</a></p>\n\n<hr>\n\n<h2 id=\"socialt\">Socialt</h2>\n\n<p>Att studera behöver inte bara vara långa kvällar med tunga böcker. Datasektionen anordnar pubar, fester, spelkvällar och andra roliga aktiviteter som ger dig en chans att koppla av mellan studierna och lära känna andra studerande. Aktiviteterna arrangeras av våra medlemmar och som medlem är du självklart välkommen.</p>\n\n<p><a href=\"/sektionen\" class=\"action\">Mer om sektionen &raquo;</a></p>\n\n<hr>\n\n<h2 id=\"näringsliv\">Näringsliv</h2>\n\n<p>Datasektionens näringslivsgrupp arbetar aktivt för ett nära samarbete mellan sektionens medlemmar och aktörer i näringslivet, som i många fall kan bli framtida arbetsgivare. Berätta om ert företag på en lunchföreläsning, eller få personlig kontakt med studenter från <a href=\"http://www.topuniversities.com/university-rankings/university-subject-rankings/2015/computer-science-information-systems#sorting=rank+region=+country=203+faculty=+stars=false+search=\" target=\"_blank\">Sveriges högst rankade datautbildning</a> på vår årliga arbetsmarknadsdag.</p>\n\n<p><a href=\"/naringsliv\" class=\"action\">Mer om samarbete &raquo;</a></p>\n",
    "anchors": [],
    "nav": [{
        "slug": "/kontakt",
        "title": "Kontakt"
    }, {
        "slug": "/namnder",
        "title": "Nämnder"
    }, {
        "slug": "/naringsliv",
        "title": "Näringsliv"
    }, {
        "slug": "/nyheter",
        "title": "Nyheter/Event"
    }, {
        "slug": "/organisation",
        "title": "Organisation"
    }, {
        "slug": "/sektionen",
        "title": "Sektionen"
    }, {
        "slug": "/studier",
        "title": "Studier"
    }]
}

```

## Installation & usage

```bash
$ go get -u github.com/datasektionen/taitan
$ taitan -v 
INFO[0000] Our root directory                            Root=dummy-data/
INFO[0000] Starting server.                             
INFO[0000] Listening on port: 4000
...
```

## API documentation

http://godoc.org/github.com/datasektionen/taitan  
http://godoc.org/github.com/datasektionen/taitan/parse

## Public domain

I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
