project_name='archimedes'

build:
	go fmt ./...
	GOOS=linux CGO_ENABLED=0 go build cmd/archimedes.go;
	sudo mv archimedes /usr/local/bin/$(project_name)

build-mac:
	go fmt ./...
	go build -ldflags "-X main.version=1.0.0" cmd/archimedes.go;
	mv archimedes ~/bin/$(project_name)