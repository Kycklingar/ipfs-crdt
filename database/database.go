package crdtdb

import (
	"database/sql"
	"errors"
	"fmt"

	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(channel string) *sql.DB {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s.sqlite3", channel))
	if err != nil {
		log.Fatal(err)
	}
	update(db)

	return db
}

func Close(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func update(db *sql.DB) {
	str := `
		CREATE TABLE IF NOT EXISTS posts(
			id INTEGER PRIMARY KEY,
			hash VARCHAR(49) UNIQUE NOT NULL,
			size int NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS tags(
			id INTEGER PRIMARY KEY,
			tag VARCHAR(64) UNIQUE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS post_tag_mapping(
			post_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			FOREIGN KEY(post_id) REFERENCES posts(id)
			FOREIGN KEY(tag_id) REFERENCES tags(id)
			CONSTRAINT post_tags_unique UNIQUE(post_id, tag_id)
		);

		CREATE TABLE IF NOT EXISTS hash_history(
			id INTEGER PRIMARY KEY,
			hash VARCHAR(49) UNIQUE NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := db.Exec(str)
	if err != nil {
		log.Fatal(err)
	}
}

func NewHash(db *sql.DB, hash string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO hash_history(hash) VALUES($1)", hash)
	return err
}

func LatestHash(db *sql.DB) (hash string) {
	db.QueryRow("SELECT hash FROM hash_history ORDER BY timestamp DESC LIMIT 1").Scan(&hash)
	return
}

func PostCount(db *sql.DB) int {
	var count int
	db.QueryRow("SELECT count() FROM posts").Scan(&count)
	return count
}

func MappingsCount(db *sql.DB) int {
	var count int
	db.QueryRow("SELECT count() FROM post_tag_mapping").Scan(&count)
	return count
}

func InsertPost(db *sql.DB, hash string, size int) error {
	tx, err := db.Begin()
	if err != nil {
		return txError(tx, err)
	}
	_, err = tx.Exec("INSERT OR IGNORE INTO posts(hash, size) VALUES($1, $2)", hash, size)
	if err != nil {
		log.Print(err)
		return txError(tx, err)
	}

	return tx.Commit()
}

type Post struct {
	ID   int
	Hash string
	Size int
	Tags []string
}

func GetPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query("SELECT id, hash FROM posts ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err = rows.Scan(&p.ID, &p.Hash)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		rws, err := db.Query("SELECT tag_id FROM post_tag_mapping WHERE post_id=$1", p.ID)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		defer rws.Close()
		for rws.Next() {
			var tagID int
			err = rws.Scan(&tagID)
			if err != nil {
				log.Print(err)
				return nil, err
			}
			var tag string
			err = db.QueryRow("SELECT tag FROM tags WHERE id=$1", tagID).Scan(&tag)
			if err != nil {
				log.Print(err)
				return nil, err
			}
			p.Tags = append(p.Tags, tag)
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func Search(tag string, db *sql.DB) ([]Post, error) {
	query := `SELECT id, hash FROM posts WHERE id IN
			(SELECT post_id FROM post_tag_mapping WHERE tag_id =
			(SELECT id FROM tags WHERE tag = $1)) ORDER BY id DESC`
	rows, err := db.Query(query, tag)
	if err != nil {
		return nil, err
	}

	var posts []Post
	for rows.Next() {
		var p Post
		if err = rows.Scan(&p.ID, &p.Hash); err != nil {
			rows.Close()
			return nil, err
		}
		posts = append(posts, p)
	}
	rows.Close()

	for i := range posts {
		rows, err := db.Query("SELECT tag FROM tags WHERE id IN(SELECT tag_id FROM post_tag_mapping WHERE post_id = $1) ORDER BY tag DESC", posts[i].ID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var t string
			err = rows.Scan(&t)
			if err != nil {
				rows.Close()
				return nil, err
			}
			posts[i].Tags = append(posts[i].Tags, t)
		}
		rows.Close()
	}
	return posts, nil
}

func AppendTagToPost(db *sql.DB, tag, postHash string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Print(err)
		return err
	}
	var postID int
	err = tx.QueryRow("SELECT id FROM posts WHERE hash=$1", postHash).Scan(&postID)
	if err != nil {
		log.Print(err)
		return txError(tx, err)
	}

	err = mapTagToPost(tx, postID, tag)
	if err != nil {
		log.Print(err)
		return txError(tx, err)
	}

	return tx.Commit()
}

func mapTagToPost(tx *sql.Tx, postID int, tag string) error {
	var tagID int
	err := tx.QueryRow("SELECT id FROM tags WHERE tag=$1", tag).Scan(&tagID)
	if err != nil {
		r, err := tx.Exec("INSERT INTO tags(tag) VALUES($1)", tag)
		if err != nil {
			log.Print(err)
			return err
		}
		i64, _ := r.LastInsertId()
		tagID = int(i64)
	}

	if postID <= 0 || tagID <= 0 {
		return errors.New("post or tag is zero")
	}
	_, err = tx.Exec("INSERT OR IGNORE INTO post_tag_mapping(post_id, tag_id) VALUES($1, $2)", postID, tagID)

	return err
}

func txError(tx *sql.Tx, err error) error {
	er := tx.Rollback()
	if er != nil {
		log.Fatal(er, err)
	}
	return err
}
