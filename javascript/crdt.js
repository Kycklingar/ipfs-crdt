// This is quickly hacked together js version of https://github.com/kycklingar/ipfs-crdt

let mngr = new Manager("test")

function Manager(channel, api="http://localhost:5001/api/v0/")
{
    this.CurrHash = ""

    this.ReadMsg = (function(msg)
    {
        console.log(msg)
        if(msg.length <= 30)
        {
            if(msg == "ASK" && this.CurrHash.length > 30)
            {
                this.Ipfs.Publish(this.CurrHash)
            }
            return
        }
        if(this.CurrHash != msg)
        {
            this.Ipfs.Cat(msg, this.buildObject)
        }
    }).bind(this)

    this.buildObject = (function(response)
    {
        var obj = new Object()
        var spl = response.split("/")
        for(var i = 0; i < spl.length; i++)
        {
            obj.AddStr(spl[i])
        }

        this.Object.Merge(obj)
        this.Publish()
    }).bind(this)

    this.Publish = (function()
    {
        this.Object.MakePosts()

        this.Ipfs.Add(this.Object.toString(), (function(resp){
            if(this.CurrHash != resp && resp.length > 40 && resp.length < 60)
            {
                this.Ipfs.Publish(resp)
                this.CurrHash = resp
                
                if(typeof(Storage) !== "undefined")
                {
                    localStorage.setItem("kycklingarCrdtCurrentHash-" + this.Ipfs.channel, this.CurrHash)
                    //localStorage.kycklingarCrdtCurrentHash = this.CurrHash
                }
            }
        }).bind(this))
    }).bind(this)

    this.Ask = (function()
    {
        this.Ipfs.Publish("ASK")
    }).bind(this)

    this.Search = (function(query)
    {
        let tmpData = this.Object.data
        let results = []
        for(let i = 0; i < tmpData.length; i++)
        {
            if(tmpData[i].Query(query))
            {
                results.push(tmpData[i])
            }
        }
        makeResults(results)
    }).bind(this)

    this.Ipfs = new Ipfs(channel, api)
    this.Object = new Object()

    if(typeof(Storage) !== "undefined")
    {
        let hash = localStorage.getItem("kycklingarCrdtCurrentHash-" + this.Ipfs.channel)
        if(typeof(hash) === "string")
        {
            this.CurrHash = hash
            this.Ipfs.Cat(this.CurrHash, this.buildObject)
        }
    }
    else
    {
        console.log("Your browser does not support localStorage")
    }

    document.title = channel
    
    clearPosts()

    this.Ipfs.Subscribe(this.ReadMsg)
    setTimeout(this.Ipfs.Publish("ASK"), 500)
}

function Object()
{
    this.data = []
    this.lock = false

    this.toString = function()
    {
        var str = ""
        while(this.lock){}
        this.lock = true
        for(var i = 0; i < this.data.length; i++)
        {
            if(this.data[i].length <= 0)
            {
                continue
            }
            str += this.data[i].toString() + "/"
            //str += this.data[i].type + "[" + this.data[i].data.toString() + "]" + "/"
        }
        this.lock = false
        return str
    }

    this.AddStr = function(str)
    {
        if(str == "")
        {
            return
        }
        var n = str.indexOf("[")
        var cd = new getCRDTData(str.substring(0, n))
        
        cd.Set(
            str.substring(
                n+1,
                str.length-1
            ).split(",")
        )

        this.Add(cd)
    }

    this.Add = function(cdata)
    {
        while(this.lock){}
        this.lock = true

        if(this.Query(cdata))
        {
            this.Smash(cdata)
        }
        else
        {
            this.data.push(cdata)
        }

        this.lock = false
    }

    this.Query = function(cdata)
    {
        for(var i = 0; i < this.data.length; i++)
        {
            if(this.data[i].Same(cdata))
            {
                return true
            }
        }
        return false
    }

    this.Smash = function(cdata)
    {
        for(var i = 0; i < this.data.length; i++)
        {
            if(this.data[i].Same(cdata))
            {
                this.data[i].Smash(cdata)
                return
            }
        }
    }

    this.Merge = function(obj)
    {
        for(var i = 0; i < obj.data.length; i++)
        {
            var found = false
            for(var j = 0; j < this.data.length; j++)
            {
                if(this.data[j].Same(obj.data[i]))
                {
                    found = true
                    this.data[j].Smash(obj.data[i])
                }
            }
            if(!found)
            {
                this.Add(obj.data[i])
            }
        }
    }

    this.MakePosts = function()
    {
        while(this.lock){}
        this.lock = true

        for(var i = 0; i < this.data.length; i++)
        {
            if(this.data[i].type == "CPOST")
            {
                makePost({"Hash":this.data[i].data[0], "Tags":this.data[i].data.slice(1)})
            }
        }
        this.lock = false
    }
}

function CRDTData()
{
    this.type = ""
    this.data = []

    this.toString = function()
    {
        return this.type + "[" + this.data.toString() + "]"
    }

    this.fromString = function(str)
    {
        // CPOST[Qmabcd...,tag1,tag2]
        var n = str.indexOf("[")
        this.type = str.substring(0, n)
        this.Set(str.substring(n+1, str.length-1).split(","))
    }

    this.Same = function(cdata)
    {
        if(this.type != cdata.type)
        {
            return false
        }
        for(var i = 0; i < cdata.data.length; i++)
        {
            if(!this.Query(cdata[i]))
            {
                return false
            }
        }
        return true
    }

    this.Smash = function(cdata)
    {
        if(
            cdata.type != this.type ||
            cdata.data.length <= 0 ||
            this.data.length <= 0
        )
        {
            return false
        }

        for(var i = 0; i < cdata.data.length; i++)
        {
            var found = false
            for(var j = 0; j < this.data.length; j++)
            {
                if(cdata.data[i] == this.data[j])
                {
                    found = true
                    break
                }
            }
            if(!found)
            {
                this.data.push(cdata.data[i])
            }
        }
        return true
    }

    this.Set = function(data)
    {
        if(data.length < 1)
        {
            return "data length < 1"
        }
        for(var i = 0; i < data.length; i++)
        {
            //console.log(data[i])
            this.data.push(data[i].trim())
        }
    }

    this.Query = function(val)
    {
        for(var i = 0; i < this.data.length; i++)
        {
            if(this.data[i] == val)
            {
                return true
            }
        }
        return false
    }
}

function CPOST()
{
    CPOST.prototype = new CRDTData
    this.type = "CPOST"

    this.Smash = function(cdata)
    {
        if(
            cdata.type != this.type || 
            cdata.data.length <= 0 ||
            this.data.length <= 0 ||
            cdata.data[0] != this.data[0]
        )
        {
            return 1
        }

        for(var i = 1; i < cdata.data.length; i++)
        {
            var found = false
            for(var j = 1; j < this.data.length; j++)
            {
                if(cdata.data[i] == this.data[j])
                {
                    found = true
                    break
                }
            }
            if(!found)
            {
                this.data.push(cdata.data[i])
            }
        }
        return 0
    }

    this.Same = function(cdata)
    {
        if(this.type != cdata.type)
        {
            return false
        }

        return this.data[0] == cdata.data[0]
    }
}
CPOST.prototype = new CRDTData

var CRDTMap = []

registerCRDTData("CPOST", CPOST)

function registerCRDTData(type, func)
{
    CRDTMap.push({"type":type, "func":func})
}

function getCRDTData(type)
{
    for(var i = 0; i < CRDTMap.length; i++)
    {
        if(CRDTMap[i].type == type)
        {
            return new CRDTMap[i].func()
        }
    }
    var r = new CRDTData()
    r.type = type
    return r
}

var pb = null

