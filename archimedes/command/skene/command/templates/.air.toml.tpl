root = "/app"
tmp_dir = "tmp"

[build]
post_cmd = ["echo 'building complete - now running in delve'"]
cmd = 'go build -gcflags "all=-N -l" github.com/odysseia-greek/olympia/{{.Name}} .'
bin = "/app/{{.Name}}"
full_bin = "dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec /app/{{.Name}}"
watch = ["./..."]
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_regex = ["_test\\.go"]
log = "air.log"

[color]
main = "magenta"
watcher = "cyan"
build = "yellow"
runner = "green"

[misc]
clean_on_exit = true

[log]
time = true
main_only = false
