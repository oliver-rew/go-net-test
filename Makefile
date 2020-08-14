all: amd64 armhf

amd64:
	go build -o nettest main.go

armhf:
		GOOS=linux GOARCH=arm CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc go build -o nettest-armhf main.go

clean:
	rm -rf nettest nettest-armhf
