all: build

build:
	rm -rf storagehelper
	go build -o storagehelper ./cmd/*.go

install:	
	install -C storagehelper /usr/local/bin

clear:
	rm -rf storagehelper