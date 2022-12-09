# WhatsappGPT 

Chat with GPT from whatsapp or make it autoreply for your specific contacts

![OpenAI Whatsapp](https://cdn.iconscout.com/icon/free/png-256/openai-1524348-1290723.png)

# Setup

    git clone https://github.com/danielgross/whatsapp-gpt.git

Once you have clone `cd whatsapp-gpt` and now we have to make sure there is GO and Python.

- Install Go you can find the guide below 

- Next Python Requirements
```bash
    pip install flask
```
- Next Playwright
```bash
    pip install playwright
```
- After installation of playwright
```bash
    playwright install
```
- Now you are good to go 😋 


## How to run

- Run the main.go first

```bash
    go run main.go
```
- Run server now

```
    python3 server.py
```    

**note:** Make sure to scan the QR code that is created or else the whole things is joke 🤡

      
## Go Installation - (Linux)

**note:** always check for the latest versions, the version numbers shown here will fall out of date quickly.

#### Installing Go on Ubuntu:

**Step 1.** Grab yourself a binary release from here: https://golang.org/dl/

You'll want to use one from the `Stable versions`, you probably want one which is in bold, for Ubuntu it's `xxx-linux-amd64.tar.gz`

```bash
wget https://storage.googleapis.com/golang/go1.4.linux-amd64.tar.gz;
```

**Step 2.** Install

```bash
sudo tar -C /usr/local -xzf go1.4.linux-amd64.tar.gz;
```

**Step 3.** Decide where your packages will live

**unless your name is also peter you might want to change that bit!**

```bash
mkdir /home/peter/.go;
```

**Step 4.** Configure Environment

On Ubuntu you can edit `~/.bashrc` to set your PATH, at the bottom add:

**change the path below to match the one you created in step 3!**

```bash
# The go binary, so we can actually run it
export PATH=$PATH:/usr/local/go/bin;

# This is where all your go packages live
GOPATH=/home/peter/.go;
export GOPATH;

# Add GOPATH/bin so compiled go libs appear on your PATH
export PATH=$PATH:$GOPATH/bin;
```

**Step 5.** Run that script

Every new terminal will run the above scripts when it starts; to apply it to the current terminal window we just:

```bash
source ~/.bashrc
```

**Step 6.** Test

```bash
$ go version
go version go1.4 linux/amd64
```

Full install instructions can be found here: https://golang.org/doc/install





