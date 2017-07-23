package crdtdb

import (
	"database/sql"
	"errors"

	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "db.sqlite3")
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
			hash VARCHAR(49) UNIQUE NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS tags(
			id INTEGER PRIMARY KEY,
			tag VARCHA(64) UNIQUE NOT NULL
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
	db.QueryRow("SELECT hash FROM hash_history ORDER BY timestamp").Scan(&hash)
	return
}

func PostCount(db *sql.DB) int {
	var count int
	db.QueryRow("SELECT count(1) FROM posts").Scan(&count)
	return count
}

func MappingsCount(db *sql.DB) int {
	var count int
	db.QueryRow("SELECT count(1) FROM post_tag_mapping").Scan(&count)
	return count
}

func InsertPost(db *sql.DB, hash string) error {
	tx, err := db.Begin()
	if err != nil {
		return txError(tx, err)
	}
	_, err = tx.Exec("INSERT OR IGNORE INTO posts(hash) VALUES($1)", hash)
	if err != nil {
		return txError(tx, err)
	}

	return tx.Commit()
}

type Post struct {
	ID   int
	Hash string
	Tags []string
}

func GetPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query("SELECT id, hash FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Hash)
		if err != nil {
			return nil, err
		}
		rws, err := db.Query("SELECT tag_id FROM post_tag_mapping WHERE post_id=$1", p.ID)
		if err != nil {
			return nil, err
		}
		defer rws.Close()
		for rws.Next() {
			var tagID int
			err = rws.Scan(&tagID)
			if err != nil {
				return nil, err
			}
			var tag string
			err = db.QueryRow("SELECT tag FROM tags WHERE id=$1", tagID).Scan(&tag)
			if err != nil {
				return nil, err
			}
			p.Tags = append(p.Tags, tag)
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func AppendTagToPost(db *sql.DB, tag, postHash string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var postID int
	err = tx.QueryRow("SELECT id FROM posts WHERE hash=$1", postHash).Scan(&postID)
	if err != nil {
		return txError(tx, err)
	}

	err = mapTagToPost(tx, postID, tag)
	if err != nil {
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
	if err != nil {
		log.Fatal(er, err)
	}
	return err
}
