job "taitan" {
  type = "service"

  group "bawang" {
    network {
      port "http" { }
    }

    service {
      name     = "taitan-bawang"
      port     = "http"
      provider = "nomad"
      tags = [
        "traefik-external.enable=true",
        "traefik-external.http.routers.taitan-bawang.rule=Host(`taitan.datasektionen.se`)",
        "traefik-external.http.routers.taitan-bawang.entrypoints=websecure",
        "traefik-external.http.routers.taitan-bawang.tls.certresolver=default",

        "traefik-internal.enable=true",
        "traefik-internal.http.routers.taitan-bawang.rule=Host(`taitan.nomad.dsekt.internal`)",
      ]
    }

    task "taitan" {
      driver = "docker"

      config {
        image = var.bawang_image_tag
        ports = ["http"]
      }

      template {
        data        = <<ENV
PORT={{ env "NOMAD_PORT_http" }}
DARKMODE_URL=https://darkmode.datasektionen.se/
CONTENT_URL=https://github.com/datasektionen/bawang-content.git
TOKEN={{ with nomadVar "nomad/jobs/taitan" }}{{ .bawang_content_token }}{{ end }}
DEFAULT_LANG=sv
ENV
        destination = "local/.env"
        env         = true
      }

      resources {
        memory = 30
      }
    }
  }

  group "styrdokument" {
    network {
      port "http" { }
    }

    service {
      name     = "taitan-styrdokument"
      port     = "http"
      provider = "nomad"
      tags = [
        "traefik-external.enable=true",
        "traefik-external.http.routers.taitan-styrdokument.rule=Host(`taitan-styrdokument.datasektionen.se`)",
        "traefik-external.http.routers.taitan-styrdokument.entrypoints=websecure",
        "traefik-external.http.routers.taitan-styrdokument.tls.certresolver=default",

        "traefik-internal.enable=true",
        "traefik-internal.http.routers.taitan-styrdokument.rule=Host(`taitan-styrdokument.nomad.dsekt.internal`)",
      ]
    }

    task "taitan" {
      driver = "docker"

      config {
        image = var.styrdokument_image_tag
        ports = ["http"]
      }

      template {
        data        = <<ENV
PORT={{ env "NOMAD_PORT_http" }}
DARKMODE_URL=https://darkmode.datasektionen.se/
CONTENT_URL=https://github.com/datasektionen/styrdokument.git
TOKEN={{ with nomadVar "nomad/jobs/taitan" }}{{ .styrdokument_token }}{{ end }}
ENV
        destination = "local/.env"
        env         = true
      }

      resources {
        memory = 30
      }
    }
  }
}

variable "bawang_image_tag" {
  type = string
  default = "ghcr.io/datasektionen/taitan:latest"
}

variable "styrdokument_image_tag" {
  type = string
  default = "ghcr.io/datasektionen/taitan:latest"
}
