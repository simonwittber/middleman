from collections import defaultdict
import serviceprovider
import textservice
import hashlib
import os


class Auth(serviceprovider.ServiceProvider, textservice.TextService):
    salts = dict()
    users = dict()

    async def REQ_Salt(self, headers, msg):
        conn = headers["cid"]
        headers["salt"] = self.salts[conn] = hashlib.md5(
            os.urandom(16)).hexdigest()

    async def REQ_Verify(self, headers, msg):
        conn = headers["cid"]
        e = headers["email"]
        h = headers["hash"]
        salt = self.salts[conn]
        # lookup pass for user
        if e in self.users:
            p = self.users[e]
            localHash = hashlib.md5((salt + p).encode()).hexdigest()
            if h == localHash:
                headers["auth"] = "Y"
                await self.internal("UID", dict(forcid=conn, setuid=hashlib.md5(e.encode()).hexdigest()))
            else:
                headers["auth"] = "N"
        else:
            headers["auth"] = "N"

    async def REQ_Register(self, headers, msg):
        h = headers["hash"]
        e = headers["email"]
        if e in self.users:
            headers["register"] = "N"
        else:
            headers["register"] = "Y"
            self.users[e] = h


serviceprovider.run(Auth, require_uid=False)
