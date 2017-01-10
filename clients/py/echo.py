import serviceprovider
import jsonservice

class Echo(serviceprovider.ServiceProvider, jsonservice.JsonService):
    async def PUB_Hello(self, headers, msg):
        print("Hello", headers, msg)

serviceprovider.run(Echo)
