## Owlet-Golang
An Owlet specific client for Ayla networks api.

An example client is provided but could be improved upon a lot. It works though.

## Disclaimer
This code is in no way "official" or affiliated with Owlet Babe Care in any way. This is an Unoffical API client and should be used at your own risk.

## Install
```
go get github.com/dgra/owlet-golang
```

## Usage
See example/main.go for an example consumer.

## Run example program(main.go)
Create a config.json in the same folder as `main.go`
```
#owlet-golang/example/config.json
{
  "email": "YOUR EMAIL HERE",
  "password: "YOUR PASSWORD HERE"
}
```
Then simply run.
```
go run main.go # or go build main.go for an executable.
```
