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
Create a config.json in the same forlder as `main.go`
```
#owlet-golang/config.json
{
  "email": "YOUR EMAIL HERE",
  "password: "YOUR PASSWORD HERE"
}
```
Then simply run.
```
go run main.go
```

## Using with docker
### Build
### This docker image may not be the best solution at the time. It _may_ not save file to local system. I haven't used/touched it in a while.
```
$ docker build -t owlet-golang -f Dockerfile .
```
### Run
```
$ docker run -it owlet-golang -e "owletEmail={YOUR EMAIL HERE} owletPassword={YOUR PASSWORD HERE}"
```
