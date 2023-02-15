# whatsapp-gpt
* You'll need to run WhatsApp from a phone number using the golang library I'm using.
* You'll run a dedicated browser in another window that's controlling ChatGPT.
* Two terminals: `go run main.go`, and `python server.py`. I am extremely doubtful they will work for you on the first run.
* You can also try `multichat.py` if you want to watch two ChatGPTs talk to each other.
* This marks the end of the readme file; it is a bit sparse; thankfully the code is too! Just tuck in if you can... and I will try to add more here later.

# Windows
* Download Golang of the https://go.dev
* Download mingw64 of the https://github.com/niXman/mingw-builds-binaries/releases and ADD it to your Windows System environment PATH
* Download Python of the https://www.python.org/
* Run `pip install playwright` and `playwright install` after installed playwright library
* Run `$ENV:CGO_ENABLED=1` in your powershell
* Run `go run main.go`, and `python server.py`
* Use your Whatsapp scan the QR Code in your golang teminals 
