import json


class JsonService:
    def decode(self, stream):
        text = stream.read()
        if text.endswith("\r\n.\r\n"):
            text = text[:-5]
        return json.loads(text) if text != "" else None

    def encode(self, obj):
        return json.dumps(obj)
