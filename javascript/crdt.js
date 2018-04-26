// This is quickly hacked together js version of https://github.com/kycklingar/ipfs-crdt
// Expect heavy refactoring

var ipfsAPI = "http://localhost:5001/api/v0/"
subscribe("test", compareData)

function subscribe(channel, callback)
{
    var req = new XMLHttpRequest()
    var lastResponse = ""
    //var count = 0

    req.onreadystatechange = function(){
        var resp = this.responseText.replace(lastResponse, "").trim()
        lastResponse = this.responseText
        //console.log(resp)
        if(resp.length > 10)
            {
                //console.log(++count)
                //resp = this.responseText
                callback(function(){
                    // K
                    //console.log(resp)
                    // try
                    // {
                    //     var obj = JSON.parse(resp.substring(resp.indexOf("{"), resp.lastIndexOf("}")))
                    //     obj = obj[obj.length-1]
                    //     obj.data = atob(obj.data)
                    //     console.log(obj.data)
                    //     return obj.data
                    // }
                    // catch(e)
                    // {
                    //     console.log(resp, e)
                    //     return ""
                    // }
                    var l = resp.substring(resp.indexOf("\"data\":"))
                    l = l.substring("'data':'".length, l.indexOf(",") - 1)
                    return atob(l)
                }())
        }
    }
    req.open("GET", ipfsAPI + "pubsub/sub?arg=" + channel, true)
    req.send()
}

function ipfsCat(hash, callback)
{
    var req = new XMLHttpRequest
    req.onreadystatechange = function(){
        if(this.readyState == 4 && this.status == 200)
        {
            callback(this.responseText)
        }
    }
    req.open("GET", ipfsAPI + "cat?arg=" + hash, true)
    req.send()
}

function ipfsAdd(string, callback)
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
    req.open("POST", ipfsAPI + "add?cid-version=1&pin=false&quieter=1", true)
    req.send(fd)
}

function ipfsPublish(arg, channel)
{
    var req = new XMLHttpRequest
    req.open("GET", ipfsAPI + "pubsub/pub?arg=" + channel + "&arg=" + arg, true)
    req.send()
}

setTimeout(ask, 1000)
function ask()
{
    ipfsPublish('ASK', 'test')
    return false
}

var currHash = ""

function compareData(hash)
{
    console.log(hash)
    if(hash.length <= 30)
    {
        if(hash == "ASK" && currHash.length > 30)
        {
            ipfsPublish(currHash, "test")
        }
        return
    }
    if(currHash != hash)
    {
        ipfsCat(hash, function(response){
            var spl = response.split("/")
            var tb
            for(var i = 0; i < spl.length; i++)
            {
                var n = spl[i].indexOf("[")
                tb = cmp(newcdata(spl[i].substring(0, n), spl[i].substring(n+1, spl[i].length - 1).split(",")))
            }
            var differ = false
            if(tb.length == objects.length)
            {
                for(var i = 0; i < tb.length; i++)
                {
                    if(!query(objects, tb[i]))
                    {
                        differ = true
                        break
                    }
                }
            }
            else
            {
                differ = true
            }
            if(differ)
            {
                for(var i = 0; i < objects.length; i++)
                {
                    var tags = []
                    for(var j = 1; j < objects[i].data.length; j++)
                    {
                        tags.push(objects[i].data[j])
                    }
                    makePost({"Hash":objects[i].data[0], "Tags":tags})
                }

                ipfsAdd(tostring(objects), function(resp){
                    if(currHash != resp && resp.length > 40 && resp.length < 60)
                    {
                        ipfsPublish(resp, "test")
                        currHash = resp
                    }
                    else{
                        console.log("Current: ", currHash, "New: ", resp)
                    }
                })
            }
        })
    }
}

function newcdata(type, data)
{
    return obj = {
        "type":type,
        "data":data
    }
}

function query(obj, val)
{
    for(var i = 0; i < obj.length; i++)
    {
        if(!ccdata(obj[i], val))
        {
            return false
        }
    }
}

function add(obj, val)
{
    var data = []
    for(var i = 0; i < val.data.length; i++)
    {
        if(typeof val.data[i] != "string" || val.data[i] == "")
        {
            continue
        }
        data.push(val.data[i].trim())
    }
    val.data = data
    var found = false
    for(var i = 0; i < obj.length; i++)
    {
        if(obj[i].data[0] == val.data[0])
        {
            found = true
            obj[i].data = smash(obj[i].data, val.data)
        }
    }
    if(!found)
    {
        obj.push(val)
        //obj.push(newcdata(val.type, val.data))
    }
}

function ccdata(left, right)
{
    if(left.type != right.type)
    {
       return false
    }

    if(left.data.length != right.data.length)
    {
        return false
    }

    if(left.data[0] != right.data[0])
    {
        return false
    }

    for(var i = 0; i < left.data.length; i++)
    {
        var found = false
        for(var j = 0; j < right.data.length; j++)
        {
            if(left.data[i] == right.data[j])
            {
                found = true
                break
            }
        }
        if(!found)
        {
            return false
        }
    }
    return true
}

function tostring(obj)
{
    var str = ""
    for(var i = 0; i < obj.length; i++)
    {
        if(obj[i].length <= 0)
        {
            continue
        }
        //console.log(obj[i])
        str += obj[i].type + "[" + obj[i].data.toString() + "]" + "/"
    }
    return str
}

var objects = []

function cmp(obj)
{
    if(obj.type == "")
    {
        return objects
    }
    var newObj = objects

    var found = false
    for(var i = 0; i < objects.length; i++)
    {
        if(obj.type == objects[i].type)
        {
            if(obj.data[0] == objects[i].data[0])
            {
                found = true
                newObj[i].data = smash(objects[i].data, obj.data)
                break
            }
        }
    }
    if(!found)
    {
        newObj.push(obj)
    }
    return newObj
}

function smash(left, right)
{
    var retObj = left
    for(var i = 1; i < right.length; i++)
    {
        var found = false
        for(var j = 1; j < left.length; j++)
        {
            if(left[j] == right[i])
            {
                found = true
                break
            }
        }
        if(!found)
        {
            retObj.push(right[i])
        }
    }
    return retObj
}

var pb = null
setTimeout(function(){pb = document.getElementById("posts")}, 500)

function makePost(post)
{
    //console.log(post)
    var d = document.getElementById(post.Hash)
    if(d == null)
    {
        n = document.createElement("div")
        n.id = post.Hash
        n.appendChild(image("/ipfs/" + post.Hash))
        n.appendChild(tags(post.Tags))
        pb.appendChild(n)
    }
    else
    {
        d.removeChild(d.getElementsByTagName("ul")[0])
        d.appendChild(tags(post.Tags))
    }
}

function tags(tags)
{
    var ul = document.createElement("ul")
    for(var i = 0; i < tags.length; i++)
    {
        var li = document.createElement("li")
        li.innerText = tags[i]
        ul.appendChild(li)
    }
    return ul
}

function image(src)
{
    var i = new Image()
    i.src = src
    return i
}

function submitNew()
{
    var hash = document.getElementById("formHash").value
    var tags = document.getElementById("formTags").value.split(",")

    var a = []
    a.push(hash)
    a = a.concat(tags)

    add(objects, newcdata("CPOST", a))

    publish()

    return false
}

function publish()
{
    for(var i = 0; i < objects.length; i++)
    {
        var tags = []
        for(var j = 1; j < objects[i].data.length; j++)
        {
            tags.push(objects[i].data[j])
        }
        makePost({"Hash":objects[i].data[0], "Tags":tags})
    
        ipfsAdd(tostring(objects), function(resp){
            if(currHash != resp && resp.length > 40 && resp.length < 60)
            {
                ipfsPublish(resp, "test")
                currHash = resp
            }
            else{
                console.log("Current: ", currHash, "New: ", resp)
            }
        })
    }
}
