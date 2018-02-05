#!/usr/bin/env python
import asyncio
import websockets
import io
import uuid

SERVER_KEY = "xyzzy"
#-----------------------------------------------------------------------
class ServiceProvider:
    def init(self, ws, require_uid):
        self.requests = {}
        self.pub_handlers = {}
        self.req_handlers = {}
        self.ws = ws
        self.public_pub_events = []
        self.require_uid = require_uid
    #------------------------------------------------------------------------ 
    async def subscribe_to_events(self): 
        prefix = self.__class__.__name__
        for name in dir(self):
            value = getattr(self, name)
            if callable(value):
                ev = name[4:]
                if name.startswith("PUB_"): 
                    print("Subscribing to External event: " + ev)
                    self.pub_handlers[ev] = value
                if name.startswith("REQ_"): 
                    print("Responding to External request: " + ev)
                    self.req_handlers[ev] = value
                if name.startswith("TASK_"): 
                    print("Starting Task:", name[5:])
                    asyncio.ensure_future(value())
        await self.epub(prefix)
        await self.ereq(prefix)
        await self.sub(prefix)
    #------------------------------------------------------------------------ 
    async def ereq(self, name):
        await self.ws.outbox.put("EREQ "+name+"\n\n")
    #------------------------------------------------------------------------ 
    async def epub(self, name):
        await self.ws.outbox.put("EPUB "+name+"\n\n")
    #------------------------------------------------------------------------ 
    async def esub(self, name):
        await self.ws.outbox.put("ESUB "+name+"\n\n")
    #------------------------------------------------------------------------ 
    async def internal(self, name, headers):
        await self.ws.outbox.put("INT "+name+"\n"+self.headerText(headers)+"\n")
    #------------------------------------------------------------------------ 
    async def pub(self, name, headers, msg):
        await self.ws.outbox.put("PUB "+name+"\n"+self.headerText(headers)+"\n"+self.encode(msg))
    #------------------------------------------------------------------------ 
    async def sub(self, name, headers=None):
        await self.ws.outbox.put("SUB "+name+"\n"+self.headerText(headers)+"\n")
    #------------------------------------------------------------------------ 
    async def uns(self, name, headers=None):
        await self.ws.outbox.put("UNS "+name+"\n"+self.headerText(headers)+"\n")
    #------------------------------------------------------------------------ 
    async def req(self, name, headers, msg):
        if headers is None: headers = {}
        request_id = headers["rid"] = uuid.uuid1().hex
        future = self.requests[request_id] = asyncio.Future()
        await self.ws.outbox.put("REQ "+name+"\n"+self.headerText(headers)+"\n"+self.encode(msg))
        return await future
        return future.result()
    #------------------------------------------------------------------------ 
    async def res(self, name, headers, msg):
        await self.ws.outbox.put("RES "+name+"\n"+self.headerText(headers)+"\n"+self.encode(msg))
    #------------------------------------------------------------------------ 
    async def recv_pub(self, name, headers, stream):
        method = headers["cmd"]
        await self.pub_handlers[method](headers, self.decode(stream))
    #------------------------------------------------------------------------ 
    async def recv_req(self, name, headers, stream):
        method = headers["cmd"]
        return await self.req_handlers[method](headers, self.decode(stream))
    #------------------------------------------------------------------------ 
    async def recv_res(self, name, headers, stream):
        future = self.requests.pop(headers["rid"])
        future.set_result(self.decode(stream))
    #------------------------------------------------------------------------ 
    def headerText(self, headers):
        return "" if headers is None else "".join("{}:{}\n".format(*i) for i in headers.items())
    #------------------------------------------------------------------------ 
    async def handle_incoming(self, stream):
        try:
            header = stream.readline().strip().split(" ")
            cmd = header[0].strip()
            name = header[1].strip()
            headers = {}
            while True:
                header = stream.readline().strip()
                if header == "" or header is None: break
                parts = header.split(":")
                headers[parts[0].strip().lower()] = ":".join(parts[1:]).strip()
            uid = headers.get("uid", "")
            if uid == "": uid = None
            try:
                if self.require_uid and uid is None: 
                    raise Exception("No UID")
                if cmd == "REQ":
                    result = await self.recv_req(name, headers, stream)
                    await self.res(name, headers, result)
                elif cmd == "PUB":
                    await self.recv_pub(name, headers, stream)
                elif cmd == "RES":
                    await self.recv_res(name, headers, stream)
            except Exception as e:
                await self.pub("MSG:"+headers["cid"], dict(cmd="error", type=e.__class__.__name__, msg=str(e)), None)
        except Exception as e:
            print(type(e), e)
            raise
#-----------------------------------------------------------------------
async def handle_outgoing_queue(ws):
    while ws.open:
        msg = await ws.outbox.get()
        await ws.send(msg + "\r\n.\r\n")
#-----------------------------------------------------------------------
async def service(service_provider_class, require_uid=True):
    ws = None
    while ws is None:
        try:
            ws = await websockets.connect('ws://127.0.0.1:8765/')
        except ConnectionRefusedError as e:
            print(e)
            await asyncio.sleep(5)
            ws = None 
    await ws.send(SERVER_KEY)
    ok = await ws.recv()
    if ok != "MM OK":
        raise Exception("Bad Key")
    myConnID = await ws.recv()
    print("My Conn ID: " + myConnID)
    ws.outbox = asyncio.Queue()
    send_task = asyncio.ensure_future(handle_outgoing_queue(ws))
    sp = service_provider_class()
    sp.init(ws, require_uid)
    await sp.subscribe_to_events()
    while True:
        msg = await ws.recv()
        if msg is None: break
        if isinstance(msg, bytes): msg = msg.decode()
        stream = io.StringIO(msg)
        asyncio.ensure_future(sp.handle_incoming(stream))
    send_task.cancel()
    await ws.close()
#-----------------------------------------------------------------------
def run(service_provider_class, *tasks, require_uid=True):
    for t in tasks:
        asyncio.ensure_future(t())
    asyncio.get_event_loop().run_until_complete(service(service_provider_class, require_uid=require_uid))
