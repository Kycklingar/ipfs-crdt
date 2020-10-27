function makePost(post)
{
    if(pb == null)
    {
        pb = document.getElementById("posts")
    }
    //console.log(post)

    var d = document.getElementById(post.Hash)
    if(d == null)
    {
        var n = document.createElement("div")
        n.className = "post"
        n.id = post.Hash

        var a = document.createElement("a")
        a.href = "/ipfs/" + post.Hash
        a.target = "_blank"
        var h4 = document.createElement("h4")
        h4.innerText = post.Hash
        a.appendChild(h4)
        n.appendChild(a)

        n.appendChild(tags(post.Tags))
        pb.appendChild(n)
        
        var req = new XMLHttpRequest
        req.onreadystatechange = function(){
            if(this.readyState == 4 && this.status == 200)
            {
                var ct = this.getResponseHeader("Content-Type")
                if(ct.split("/")[0] == "image")
                {
                    b = image("/ipfs/" + post.Hash)
                    n.children[0].replaceChild(b, n.children[0].children[0])
                }
            }
        }
        req.open("HEAD", "/ipfs/" + post.Hash, true)
        req.send()
    }
    else
    {
        d.removeChild(d.getElementsByTagName("ul")[0])
        d.appendChild(tags(post.Tags))
    }
}

function makeResults(posts)
{
    let sr = document.getElementById("searchResults")
    if(sr == null)
    {
        sr = document.createElement("div")
        sr.id = "searchResults"
        document.body.appendChild(sr)
    }

    removeChilds(sr)

    let title = document.createElement("h3")
    title.innerText = "Results (" + posts.length + ")"
    
    sr.appendChild(title)

    let x = document.createElement("div")
    x.innerText = "X"
    x.style.position = "absolute"
    x.style.right = 0
    x.style.top = 0
    x.onclick = function()
    {
        removeChilds(sr)
        sr.style.display = "none"
    }

    sr.appendChild(x)

    sr.style.display = null

    for(let i = 0; i < posts.length; i++)
    {
        let post = {"Hash":posts[i].data[0], "Tags":posts[i].data.slice(1)}
        var n = document.createElement("div")
        n.className = "post"
        // n.id = post.Hash

        var a = document.createElement("a")
        a.href = "/ipfs/" + post.Hash
        a.target = "_blank"
        var h4 = document.createElement("h4")
        h4.innerText = post.Hash
        a.appendChild(h4)
        n.appendChild(a)

        n.appendChild(tags(post.Tags))
        sr.appendChild(n)
        
        function determine(postDiv, postData)
        {
            var req = new XMLHttpRequest
            req.onreadystatechange = function(){
                if(this.readyState == 4 && this.status == 200)
                {
                    var ct = this.getResponseHeader("Content-Type")
                    if(ct.split("/")[0] == "image")
                    {
                        b = image("/ipfs/" + postData.Hash)
                        postDiv.children[0].replaceChild(b, postDiv.children[0].children[0])
                    }
                }
            }
            req.open("HEAD", "/ipfs/" + postData.Hash, true)
            req.send()
        }
        determine(n, post)
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
    if(mngr == null)
    {
        console.log("Initialize a manager")
        return
    }

    var hash = document.getElementById("formHash").value
    var tags = document.getElementById("formTags").value.split(",")

    var a = []
    a.push(hash)
    a = a.concat(tags)

    var c = new CPOST()
    c.Set(a)

    mngr.Object.Add(c)
    mngr.Publish()

    return false
}

function removeChilds(parent)
{
    if(parent == null){
        return
    }

    while(parent.children.length > 0)
    {
        parent.removeChild(parent.children[0])
    }
}

function clearPosts()
{
    if(pb == null)
    {
        pb = document.getElementById("posts")
    }

    removeChilds(pb)
}
