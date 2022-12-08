"""Get ChatGPT to talk to itself."""

import os
import sys
import subprocess
import requests

metaprompt = "Now make that funnier."
chat1 = requests.get("http://localhost:5001/chat?q=%s" % "Teach me about quantum mechanics in a 140 characters or less.")
while True:
    chat2 = requests.get("http://localhost:5002/chat?q=%s" % (chat1.text.replace(metaprompt, "") + " " + metaprompt))
    chat1 = requests.get("http://localhost:5001/chat?q=%s" % (chat2.text.replace(metaprompt, "") + " " + metaprompt))

    