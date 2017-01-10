
var MiddleMan = {

    ApiKey: "be4e5ef3-af59-4bec-a5a6-245b4594443f", 

    Connect: function(wsUri) {
        var ws = new WebSocket(wsUri);
        ws.onopen = MiddleMan.onopen;
        ws.onclose = MiddleMan.onclose;
        ws.onmessage = MiddleMan.onfirstmessage;
        ws.onerror = MiddleMan.onerror;
        MiddleMan.ws = ws;
    }, 

    Publish: function(name, headers, body) {
        if(headers === null) headers = {};
        var msg = "PUB " + name + "\r\n" + MiddleMan.formatHeaders(headers) + "\r\n" + body + "\r\n.\r\n";
        MiddleMan.ws.send(msg);
    },

    Subscribe: function(name, headers, callback) {
        if(headers === null) headers = {};
        MiddleMan.subscriptions[name] = callback;
        var msg = "SUB " + name + "\r\n" + MiddleMan.formatHeaders(headers) + "\r\n" + "\r\n.\r\n";
        MiddleMan.ws.send(msg);
    },

    Request: function(name, headers, body, callback) {
        if(headers === null) headers = {};
        var uid = MiddleMan.guid();
        MiddleMan.requests[uid] = callback
        headers["ReqID"] = uid; 
        var msg = "REQ " + name + "\r\n" + MiddleMan.formatHeaders(headers) + "\r\n" + body + "\r\n.\r\n";
        MiddleMan.ws.send(msg);
    },

    formatHeaders: function(headers) {
        var txt = "";
        if(headers) {
            for(var k in headers) {
                txt = k + ": " + headers[k] + "\r\n";   
            }
        }
        return txt;
    },

    guid: function() {
        function s4() {
            return Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1);
        }
        return s4() + s4() + '-' + s4() + '-' + s4() + '-' + s4() + '-' + s4() + s4() + s4();
    },

    onopen: function (evt) {
        MiddleMan.ws.send(MiddleMan.ApiKey);
    },

    onclose: function (evt) {
    },

    onfirstmessage: function(evt) {
        if(evt.data == "MM OK") {
            console.log("MiddleMan is ready.")
            MiddleMan.ws.onmessage = MiddleMan.onmessage
        } else {
            console.log("Bad Key")
        }
    }, 

    onmessage: function(evt) {
        var lines = evt.data.split("\n");
        var A = lines.shift().split(" ");
        var headers = {};
        while(true) {
            var H = lines.shift().trim();
            if(H == "" || H == null) break;
            var parts = H.split(":");
            headers[parts.shift().trim()] = parts.join(":").trim();
        }
        var body = ""
        for(var line in lines) {
            if(line == ".") break
            body += line + "\n"
        }
        var cmd = A[0].trim();
        var name = A[1].trim();
        switch(cmd) {
            case "PUB":
                var fn = MiddleMan.subscriptions[name];
                if(fn != null) fn(headers, body);
            break;
            case "RES":
                var fn = MiddleMan.requests[headers["ReqID"]];
                if(fn === null) 
                    console.log("Lost Response: " + name + " " + format(headers))
                else {
                    delete MiddleMan.requests[headers["ReqID"]];
                    fn(headers, body);
                }
            break;
        }
    },

    onerror: function(evt) {
        console.log("Error: " + evt);
    },

    requests: {},

    subscriptions: {},

}


