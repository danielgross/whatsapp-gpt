"""Make some requests to OpenAI's chatbot"""

import time
import os
import flask
import sys
from flask import Flask, jsonify
import json
import re
import random
import string
from gtts import gTTS
import speech_recognition as sr

from flask import g

from playwright.sync_api import sync_playwright

PROFILE_DIR = "/tmp/playwright" if '--profile' not in sys.argv else sys.argv[sys.argv.index('--profile') + 1]
PORT = 5001 if '--port' not in sys.argv else int(sys.argv[sys.argv.index('--port') + 1])
APP = flask.Flask(__name__)
APP.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # 16 MB

PLAY = sync_playwright().start()

#BROWSER = PLAY.chromium.launch_persistent_context(
#    user_data_dir=PROFILE_DIR,
#    headless=False,
#)
BROWSER = PLAY.firefox.launch_persistent_context(
    user_data_dir=PROFILE_DIR,
    headless=False,
    java_script_enabled=True,
)
PAGE = BROWSER.new_page()

def get_input_box():
    """Find the input box by searching for the largest visible one."""
    textareas = PAGE.query_selector_all("textarea")
    candidate = None
    for textarea in textareas:
        if textarea.is_visible():
            if candidate is None:
                candidate = textarea
            elif textarea.bounding_box().width > candidate.bounding_box().width:
                candidate = textarea
    return candidate

def is_logged_in():
    try:
        return get_input_box() is not None
    except AttributeError:
        return False

def send_message(message):
    # Send the message
    box = get_input_box()
    box.click()
    box.fill(message)
    box.press("Enter")
    while PAGE.query_selector(".result-streaming") is not None:
        time.sleep(0.1)

def get_mp3_file(message):
    language = 'en'
    # convert the text response from chatgpt to an audio file
    audio = gTTS(text=message, lang=language, slow=False, )
    # save the audio file
    name = random_file_name() + '.mp3'
    audio.save(name)
    return name

def random_file_name(length=8):
    letters = string.ascii_letters
    return ''.join(random.choice(letters) for i in range(length))

def get_last_message():
    """Get the latest message"""
    page_elements = PAGE.query_selector_all(".flex.flex-col.items-center > div")
    last_element = page_elements[-2]
    return last_element.inner_text()

def get_last_code():
    """Get the latest message"""
    page_elements = PAGE.query_selector_all(".flex.flex-col.items-center > div")
    last_element = page_elements[-2]
    pattern = r"^Copy code$\n(?P<lines>(?P<line>.+$\n)+)"

    match = re.search(pattern, last_element.inner_text(), re.MULTILINE)
    if match:
        lines = match.group("lines")
        return lines
    else:
        return

@APP.route("/chat", methods=["GET"])
def chat():
    message = flask.request.args.get("q")
    print("Sending message: ", message)
    send_message(message)
    response = get_last_message()
    code = get_last_code()
    mp3 = get_mp3_file(response)
    data = {
        "code": code,
        "response": response,
        "mp3": mp3
    }
    return jsonify(data)

@APP.route("/transcribe", methods=["GET"])
def transcribe():
    message = flask.request.args.get("q")
    print("Sending message: ", message)
    r = sr.Recognizer()
    with sr.AudioFile(message) as source:
        audio_text = r.listen(source)
        text = r.recognize_google(audio_text)
        print("The audio file contains: " + text)
    data = {
        "response": text
    }
    return jsonify(data)

def start_browser():
    PAGE.goto("https://chat.openai.com/")
    APP.run(port=PORT, threaded=False)
    if not is_logged_in():
        print("Please log in to OpenAI Chat")
        print("Press enter when you're done")
        input()
    else:
        print("Logged in")

if __name__ == "__main__":
    start_browser()