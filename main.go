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

	DB "github.com/kycklingar/ipfs-crdt/database"
)

var idM *idManager
var db *sql.DB

func main() {
	log.SetFlags(log.Llongfile)

	port := flag.Int("port", 80, "Local webserver port")
	channel := flag.String("ch", "test", "Channel to listen on")
	help := flag.Bool("help", false, "Prints help text")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	db = DB.InitDB(*channel)

	idM = newManager()
	idM.Init()
	go idM.listen(*channel)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/db/populate", populateDBHandler)

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
			<form action="/db/populate">
				<input type="submit" value="Populate DB">
			</form>
			<ul>
				{{range .Content}}
					<li>
						<div>
							<span><img width="250px" src="http://localhost:8080/ipfs/{{.Hash}}"></span>
							<p>{{.Size}}B</p>
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

	type p struct {
		Hash    string
		Channel string
		Content []DB.Post
	}

	c, err := DB.GetPosts(db)
	if err != nil {
		fmt.Fprintln(w, err)
		return
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

	stat, err := idM.ipfs.ObjectStat(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var pd postData
	err = pd.set(post, stat.DataSize)
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

func populateDBHandler(w http.ResponseWriter, r *http.Request) {
	idM.d.mutex.Lock()
	defer idM.d.mutex.Unlock()

	tagCount := DB.MappingsCount(db)
	postCount := DB.PostCount(db)

	for _, d := range idM.d.data {
		switch v := d.(type) {
		case *postData:
			err := DB.InsertPost(db, v.Hash, v.Size)
			if err != nil {
				fmt.Fprintln(w, err)
			}
		case *tagData:
			err := DB.AppendTagToPost(db, v.Tag, v.PostHash)
			if err != nil {
				fmt.Fprintln(w, err)
			}
		default:
			fmt.Println("error")
		}
	}

	newTagCount := DB.MappingsCount(db)
	newPostCount := DB.PostCount(db)

	fmt.Fprintln(w, "New posts: ", newPostCount-postCount)
	fmt.Fprintln(w, "New tags: ", newTagCount-tagCount)
}
