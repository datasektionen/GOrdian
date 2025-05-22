job "gordian" {
  namespace = "money"

  type = "service"

  group "gordian" {
    network {
      port "http" { }
    }

    service {
      name     = "gordian"
      port     = "http"
      provider = "nomad"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.gordian.rule=Host(`budget.datasektionen.se`)",
        "traefik.http.routers.gordian.tls.certresolver=default",
      ]
    }

    task "gordian" {
      driver = "docker"

      config {
        image = var.image_tag
        ports = ["http"]
      }

      template {
        data        = <<ENV
SERVER_PORT={{ env "NOMAD_PORT_http" }}
SERVER_URL=https://budget.datasektionen.se
{{ with nomadVar "nomad/jobs/gordian" }}
GO_CONN=postgres://gordian:{{ .db_password }}@postgres.dsekt.internal:5432/gordian?sslmode=disable
CF_CONN=postgres://gordian:{{ .db_password }}@postgres.dsekt.internal:5432/cashflow?sslmode=disable # cursed, should use API
LOGIN_TOKEN={{ .login_token }}
{{ end }}
LOGIN_API_URL=http://sso.nomad.dsekt.internal/legacyapi
LOGIN_FRONTEND_URL=https://sso.datasektionen.se/legacyapi
PLS_URL=https://pls.datasektionen.se
PLS_SYSTEM=gordian
ENV
        destination = "local/.env"
        env         = true
      }
    }
  }
}

variable "image_tag" {
  type = string
  default = "ghcr.io/datasektionen/gordian:latest"
}
