group "default" {
  targets = [
	"build"
	# "package"
  ]
}

variable "TAG" {
  default = "local"
}
variable "REPO" {
  default = "ghcr.io/lesomnus/flob"
}
variable "BUILD_HASH" {
  default = "0000000000000000000000000000000000000000"
}
variable "BUILD_ID" {
  default = "r0"
}

function "tags" {
  params = [name]
  result = [
    "${REPO}/${name}:${TAG}",
    "${REPO}/${name}:${BUILD_ID}",
  ]
}

target "build" {
  context    = "."
  dockerfile = "./Dockerfile"
  args = {
    BUILD_HASH = BUILD_HASH
    BUILD_ID   = BUILD_ID
  }
  
  output = [{ type = "local", dest = "./output" }]
}
target "build-cache" {
  inherits = ["build"]
  cache-from = [{ type = "registry", ref = "${REPO}:cache" }]
  cache-to   = [{ type = "registry", ref = "${REPO}:cache", mode = "max" }]
}

target "package" {
  depends_on = ["build"]
  context    = "./output"

  dockerfile-inline = <<EOT
FROM scratch

ARG TARGETARCH
COPY "./$${TARGETARCH}" /flob

VOLUME ["/.flob"]

ENTRYPOINT ["/flob"]
CMD ["--help"]
EOT

  tags = [
    "${REPO}:${TAG}",
    "${REPO}:${BUILD_ID}",
  ]
}
