#!/bin/sh

# exit on error
set -e

echo "Installing python dependencies..."
pip install -r requirements.txt

echo "Downloading chromium..."
playwright install

echo "Done!"
echo "You can now run the script with 'python server.py'"