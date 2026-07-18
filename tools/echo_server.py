from http.server import BaseHTTPRequestHandler, HTTPServer
import json


class EchoHandler(BaseHTTPRequestHandler):
    def _handle(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length).decode("utf-8", errors="replace") if length else ""

        response = {
            "method": self.command,
            "path": self.path,
            "headers": dict(self.headers),
            "body": body,
        }

        payload = json.dumps(response, indent=2).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def do_GET(self):
        self._handle()

    def do_POST(self):
        self._handle()

    def do_PUT(self):
        self._handle()

    def do_DELETE(self):
        self._handle()

    def log_message(self, format, *args):
        print(f"[echo_server] {self.command} {self.path}")


if __name__ == "__main__":
    port = 9000
    server = HTTPServer(("localhost", port), EchoHandler)
    print(f"Local echo server running on http://localhost:{port}")
    server.serve_forever()