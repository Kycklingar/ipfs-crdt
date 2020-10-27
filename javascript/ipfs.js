function Ipfs(channel, api)
{
    this.channel = channel
    this.API = api

    this.Subscribe = function(callback)
    {
        if(this.channel == "")
        {
            console.log("Channel not specified")
            return 
        }
        var req = new XMLHttpRequest()
        var lastResponse = ""
        //var count = 0
    
        req.onreadystatechange = function(){
            var resp = this.responseText.replace(lastResponse, "").trim()
            lastResponse = this.responseText
            //console.log(resp)
            if(resp.length > 10)
                {
                    var msg = resp.split("\n")
                    //console.log(msg.length)
                    for(var i = 0; i < msg.length; i++)
                    {
                        var l = msg[i].substring(msg[i].indexOf("\"data\":"))
                        l = l.substring("'data':'".length, l.indexOf(",") - 1)
                        var data = atob(l)
                        callback(data)
                    }
            }
        }
        req.open("GET", this.API + "pubsub/sub?discover=true&arg=" + this.channel, true)
        req.send()
    }

    this.Cat = function(hash, callback)
    {
        var req = new XMLHttpRequest
        req.onreadystatechange = function(){
            if(this.readyState == 4 && this.status == 200)
            {
                callback(this.responseText)
            }
        }
        req.open("GET", this.API + "cat?arg=" + hash, true)
        req.send()
    }

    this.Add = function(string, callback)
    {
        var req = new XMLHttpRequest
        req.onreadystatechange = function(){
            if(this.readyState == 4 && this.status == 200)
            {
                var j = JSON.parse(this.responseText)
                callback(j.Hash)
            }
        }
        var fd = new FormData()
        var data = new Blob([string], {type: 'text/plain'});
        fd.append("arg", data)
        req.open("POST", this.API + "add?cid-version=1&pin=false&quieter=1", true)
        req.send(fd)
    }

    this.Publish = function(arg)
    {
        var req = new XMLHttpRequest
        req.open("GET", this.API + "pubsub/pub?arg=" + this.channel + "&arg=" + arg, true)
        req.send()
    }
}
