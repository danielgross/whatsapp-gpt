"""Make some requests to OpenAI's chatbot"""

import time
import os 
import flask
import sys

from flask import g

from playwright.sync_api import sync_playwright

PROFILE_DIR = "/tmp/playwright" if '--profile' not in sys.argv else sys.argv[sys.argv.index('--profile') + 1]
PORT = 5001 if '--port' not in sys.argv else int(sys.argv[sys.argv.index('--port') + 1])
APP = flask.Flask(__name__)
PLAY = sync_playwright().start()
BROWSER = PLAY.chromium.launch_persistent_context(
    user_data_dir=PROFILE_DIR,
    headless=False,
)
USERS = []
PAGES = []

def get_input_box(user):
    textareas = PAGES[USERS.index(user)].query_selector_all("textarea")
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

def send_message(user, message):
    box = get_input_box(user)
    box.click()
    box.fill(message)
    box.press("Enter")
    time.sleep(1)
    while PAGES[USERS.index(user)].query_selector('.self-end.visible') is None:
        time.sleep(0.1)

def get_last_message(user):
    page_elements = PAGES[USERS.index(user)].query_selector_all('[data-message-author-role="assistant"]')
    last_element = page_elements[len(page_elements) - 1]
    return last_element.inner_text()

@APP.route("/chat", methods=["GET"])
def chat():
    user = flask.request.args.get("user")
    index = USERS.index(user) if user in USERS else -1
    if index == -1:
        USERS.append(user)
        PAGES.append(BROWSER.new_page())	
        PAGES[USERS.index(user)].goto("https://chat.openai.com/")
        time.sleep(2)

    message = flask.request.args.get("q")
    print("Sending message: ", message)
    send_message(user, message)
    response = get_last_message(user)
    print("Response: ", response)
    return response

def start_browser():
	PAGE = BROWSER.new_page()
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