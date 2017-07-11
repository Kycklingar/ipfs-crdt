package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var idM *idManager
var db *sql.DB

func main() {
	//DB.InitDB()

	port := flag.Int("port", 80, "Local webserver port")
	channel := flag.String("ch", "test", "Channel to listen on")
	help := *flag.Bool("help", false, "Prints help text")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	idM = newManager()
	go idM.listen(*channel)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post", postHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `
		<html>
		<head>
			<title>CRDT</title>
		</head>
		<body>
			<a href="/post">Post Hash</a>
			<h1>CRDT Content</h1>
			<h3>Channel {{.Channel}}</h3>
			<h4>Current Hash: {{.Hash}}</h4>
			<ul>
				{{range .Content}}
					<li>
						<div>
							<span><img width="250px" src="http://localhost:8080/ipfs/{{.Post.Hash}}"></span>
							<span>
								<ul>
								{{range .Tags}}
									<li>{{.}}</li>
								{{end}}
								</ul>
							</span>
						</div>
					</li>
				{{end}}
			</ul>
		</body>
		</html>
	`
	tmpl, err := template.New("index").Parse(html)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type post struct {
		Post postData
		Tags []string
	}

	type p struct {
		Hash    string
		Channel string
		Content []post
	}

	var c []post

	cd := idM.d.data
	for _, d := range cd {
		if _, ok := d.(*postData); ok {
			pd := *d.(*postData)
			var p post
			p.Post = pd
			c = append(c, p)
		} else if _, ok := d.(*tagData); ok {
			td := *d.(*tagData)
			for i, cu := range c {
				if cu.Post.Hash == td.PostHash {
					c[i].Tags = append(cu.Tags, td.Tag)
					break
				}
			}
		}
	}

	tmpl.Execute(w, p{Hash: idM.currentHash, Channel: idM.ipfs.subject, Content: c})
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Fprint(w, `
			<html>
				<head>
					<title>CRDT ADD</title>
				</head>
				<body>
					<form method="POST">
						<p>hash</p>
						<input type="text" name="hash">
						<p>tags</p>
						<input type="text" name="tags">
						<input type="submit">
					</form>
				</body>
			</html>
		`)
		return
	}

	post := r.FormValue("hash")
	tags := r.FormValue("tags")
	if post == "" || tags == "" {
		http.Error(w, "value is empty", http.StatusInternalServerError)
		return
	}

	var pd postData
	err := pd.set(post, 1000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cd []crdtData
	cd = append(cd, &pd)

	for _, t := range strings.Split(tags, ",") {
		td := tagData{}
		err = td.set(post, t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cd = append(cd, &td)
	}

	idM.add(cd...)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
