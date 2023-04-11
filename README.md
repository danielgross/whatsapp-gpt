# Whatsapp-GPT
The Whatsapp GPT client is a tool that uses the playwright library instead of the OpenAI API. This allows users to access the normal functionality of GPT-3 without the need for a premium account or additional costs. With this client, users can generate text in natural language by simply sending a message to a Whatsapp number. The client will then use GPT-3 to generate a response based on the message received. This tool is useful for anyone who needs to generate text quickly and efficiently without the hassle of setting up an API account or paying additional fees.

## Install
This installation guide assumes that you already have ***python and Go environment setup in your machine***. 

### Python 
First install following libraries, 

`$ pip install Flask`
`$ pip install playwright`

once you have installed `playwright` you also need to set it up. 

Open your terminal screen and install the required browsers write the following command, 

`playwright install`

this will install the required libraries for your python server. 

### Go
In order to install Go and it's required packages visit [this](https://go.dev/doc/install) official Go installation documentation. 

If you have already installed Go, you can check the installation from terminal, 

`$ go version`

## Server
In order to run this app you must start Go and Python server on separate terminals. 

### Go
Run the Go server using the following commands, 

`$ go run main.go`

the user will be prompted with a QR code similar to what web.whatsapp shows at the time of first authenticating device. The user has to authenticate this server as well using whatsapp application QR code scanner. 
Once authentication is done, the server will start!

### Python 
Run the python server using following command, 

`$ python server.py`

A normal Flask server will start in no time. 

## Troubleshooting
If you encounter any issues during the installation or running of the servers, check the following:

Make sure that your Python and Go environments are set up correctly and are up to date.
Check that all required libraries and packages are installed correctly.
Make sure that both the Go and Python servers are running on separate terminals and are not conflicting with each other.
If you encounter any error messages, refer to the documentation or seek help from online forums or communities.

## Todo
- Automatic multichat support
- Automatic profile creation (behavioral instructions for ChatGPT)
- Check WhatsApp Business Account compatiblity

## Known Issues
Under Windows, it is known that installing Python through the Windows Store can lead to errors. In this case, the PATH files need to be manually created.

Multisession is not Working.
