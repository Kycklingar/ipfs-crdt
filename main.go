package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var idM *idManager

func main() {
	port := "80"
	if len(os.Args) >= 2 {
		port = os.Args[1]
	}

	idM = newManager()
	go idM.listen()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post", postHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `
		<html>
		<head>
			<title>CRDT</title>
		</head>
		<body>
			<h1>CRDT Content</h1>
			<h4>Current Hash: {{.Hash}}</h4>
			<ul>
				{{range .Content}}
					<li>
						<img width="250px" src="http://localhost:8080/ipfs/{{.}}">
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
		Content []string
	}

	tmpl.Execute(w, p{Hash: idM.currentHash, Content: idM.d.multihash})
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	val := r.FormValue("val")
	if val == "" {
		http.Error(w, "value is empty", http.StatusInternalServerError)
		return
	}

	idM.add(val)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
