# Whatsapp-GPT
A simple piece of software which connects your whatsapp with chatGPT!

## Install
This installation guide assumes that you already have ***python and Go environment setup in your machine***. 

### Python 
First install following libraries, 

```
$ pip install Flask
```

```
$ pip install playwright
```

once you have installed `playwright` you also need to set it up. 

Open your terminal screen and install the required browsers write the following command, 

```
playwright install
```

this will install the required libraries for your python server. 

### Go
In order to install Go and it's required packages visit [this](https://go.dev/doc/install) official Go installation documentation. 

If you have already installed Go, you can check the installation from terminal, 

```
$ go version
```

## Server
In order to run this app you must start Go and Python server on separate terminals. 
 
### Go
Run the Go server using the following commands, 

```
$ go run main.go
```

the user will be prompted with a QR code similar to what web.whatsapp shows at the time of first authenticating device. The user has to authenticate this server as well using whatsapp application QR code scanner. 
Once authentication is done, the server will start!

### Python 
Run the python server using following command, 

```
$ python server.py
```

A normal Flask server will start in no time. 
