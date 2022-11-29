all: main.go
	go build -ldflag="-s -w -H=windowsgui"
