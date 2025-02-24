project_name='archimedes'

build:
	go fmt ./...
	go build -ldflags "-X main.version=1.0.6" cmd/archimedes.go;
	mv archimedes ~/bin/$(project_name)