"""Get ChatGPT to talk to itself."""

import requests

# Launch two instances of server.py
# python3 server.py --port 5001 --profile /tmp/chat1
# python3 server.py --port 5002 --profile /tmp/chat2

metaprompt = "Now make that funnier."
chat1 = requests.get("http://localhost:5001/chat?q=%s" % "Teach me about quantum mechanics in a 140 characters or less.")
while True:
    chat2 = requests.get("http://localhost:5002/chat?q=%s" % (chat1.text.replace(metaprompt, "") + " " + metaprompt))
    chat1 = requests.get("http://localhost:5001/chat?q=%s" % (chat2.text.replace(metaprompt, "") + " " + metaprompt))

    