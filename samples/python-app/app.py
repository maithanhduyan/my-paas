import os
import time
from flask import Flask, jsonify

app = Flask(__name__)
start_time = time.time()
visits = 0


@app.route("/")
def index():
    global visits
    visits += 1
    return jsonify(
        message="Hello from My PaaS!",
        service="sample-python-app",
        visits=visits,
        uptime=round(time.time() - start_time, 1),
        timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
    )


@app.route("/health")
def health():
    return jsonify(status="ok")


if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8000))
    app.run(host="0.0.0.0", port=port)
