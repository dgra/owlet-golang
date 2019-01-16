## Owlet-Golang
An Owlet specific client for Ayla networks api.

## Disclaimer
This is code is in no way "official" or affiliated with Owlet Babe Care in any way. This is an Unoffical API client and should be used at your own risk.

## Install
```
go get github.com/dgra/owlet-golang
```

## Usage
See main.go for an example consumer.

## Run example program(main.go)
```
OWLET_EMAIL=YOUR_EMAIL_ADDRESS OWLET_PASSWORD=YOUR_PASSWORD go run main.go
```

## Using with docker
### Build
```
$ docker build -t owlet-golang -f Dockerfile .
```
### Run
```
$ docker run -it owlet-golang -e "owletEmail={YOUR EMAIL HERE} owletPassword={YOUR PASSWORD HERE}"
```
