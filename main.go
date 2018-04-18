package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	DB "github.com/kycklingar/ipfs-crdt/database"
)

var idM *idManager
var db *sql.DB

var templates *template.Template

func loadTemplates(path string) ([]string, error) {
	paths, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, p := range paths {
		if p.IsDir() {
			// fmt.Println(p.Name() + "/")
			n, err := loadTemplates(path + "/" + p.Name())
			if err != nil {
				return nil, err
			}
			names = append(names, n...)
		} else {
			// fmt.Println(p.Name())
			names = append(names, path+"/"+p.Name())
		}
	}
	return names, nil
}

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

	names, err := loadTemplates("templates")
	if err != nil {
		log.Fatal(err)
	}
	templates, err = templates.ParseFiles(names...)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/db/populate", populateDBHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
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
	templates.ExecuteTemplate(w, "index.html", p{Hash: idM.currentHash, Channel: idM.ipfs.subject, Content: c})
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "post.html", nil)
		return
	}

	post := r.FormValue("hash")
	tags := r.FormValue("tags")
	if post == "" || tags == "" {
		http.Error(w, "value is empty", http.StatusInternalServerError)
		return
	}

	// stat, err := idM.ipfs.ObjectStat(post)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	var pd compactPost
	var data = []interface{}{post}

	for _, t := range strings.Split(tags, ",") {
		data = append(data, t)
	}
	err := pd.set(data...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cd []crdtData
	cd = append(cd, &pd)

	// for _, t := range strings.Split(tags, ",") {
	// 	td := tagData{}
	// 	err = td.set(post, t)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	cd = append(cd, &td)
	// }

	idM.add(cd...)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "tag.html", nil)
}

func populateDBHandler(w http.ResponseWriter, r *http.Request) {
	idM.d.mutex.Lock()
	defer idM.d.mutex.Unlock()

	tagCount := DB.MappingsCount(db)
	postCount := DB.PostCount(db)

	for _, d := range idM.d.data {
		switch v := d.(type) {
		// case *postData:
		// 	err := DB.InsertPost(db, v.Hash, v.Size)
		// 	if err != nil {
		// 		fmt.Fprintln(w, err)
		// 	}
		// case *tagData:
		// 	err := DB.AppendTagToPost(db, v.Tag, v.PostHash)
		// 	if err != nil {
		// 		fmt.Fprintln(w, err)
		// 	}
		case *compactPost:
			err := DB.InsertPost(db, v.Hash, 0)
			if err != nil {
				fmt.Fprintln(w, err)
			}
			for _, tag := range v.Tags {
				err = DB.AppendTagToPost(db, tag, v.Hash)
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
