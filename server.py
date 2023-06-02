"""Make some requests to OpenAI's chatbot"""

import flask
import sys
import openai

from flask import g


PORT = 5001 if '--port' not in sys.argv else int(sys.argv[sys.argv.index('--port') + 1])
APP = flask.Flask(__name__)
openai.api_key = 'Your OpenAI API Key'


@APP.route("/chat", methods=["GET"])
def chat():
    message = flask.request.args.get("q")
    print("Sending message: ", message)
    response = openai.Completion.create(model="text-davinci-003", prompt=""+ message, max_tokens=200,top_p=0.0,frequency_penalty=0.5,presence_penalty=0.0)
    print("Response: ", response)
    return response.choices[0].text

def startAI():
    APP.run(port=PORT, threaded=False)

        
if __name__ == "__main__":
    startAI()