from collections import defaultdict
import serviceprovider
import jsonservice

class Chat(serviceprovider.ServiceProvider, jsonservice.JsonService):
    rooms = defaultdict(set)
    clients = defaultdict(set)

    async def PUB_Hello(self, headers, msg):
        print("Hello", headers, msg)

    async def REQ_Leave(self, headers, msg):
        conn = headers["cid"]
        room = headers["room"]
        self.clients[conn].remove(room)
        self.rooms[room].remove(conn)

    async def REQ_Join(self, headers, msg):
        print("REQ");
        conn = headers["cid"]
        room = headers["room"]
        self.clients[conn].add(room)
        self.rooms[room].add(conn)

    async def PUB_Say(self, headers, msg):
        print("PUB");
        conn = headers["cid"]
        room = headers["room"]
        headers["command"] = "Say"
        headers["from"] = conn
        print(self.rooms)
        for i in self.rooms[room]:
            await self.pub("MSG:"+i, headers=headers, msg=msg)
        




serviceprovider.run(Chat)