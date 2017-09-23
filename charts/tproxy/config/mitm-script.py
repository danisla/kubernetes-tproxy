from mitmproxy import http
import re
import os

ALLOW_BUCKET_NAME = "solutions-public-assets"

RE_BUCKET = re.compile(r'https://storage.googleapis.com/%s/.*' % ALLOW_BUCKET_NAME)

def request(flow: http.HTTPFlow) -> None:
    if not RE_BUCKET.match(flow.request.pretty_url):
        flow.response = http.HTTPResponse.make(
            418,
            b"Access Denied by Administrator",
            {"Content-Type": "text/html"}
        )
