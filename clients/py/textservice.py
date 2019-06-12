
class TextService:
    def decode(self, stream):
        text = stream.read()
        if text.endswith("\r\n.\r\n"):
            text = text[:-5]
        return text if text != "" else None

    def encode(self, obj):
        return obj if obj != None else ""
