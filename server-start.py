import subprocess
import time

while True:
    # Launch Script
    subprocess.Popen(["python3", "server.py"])

    # Wait 1 hour
    time.sleep(3600)

    # Kill it
    subprocess.call(["pkill", "-f", "server.py"])
