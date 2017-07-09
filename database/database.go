package database

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
			hash VARCHAR(49) UNIQUE NOT NULL,
			size INTEGER NOT NULL
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
	`

	_, err := db.Exec(str)
	if err != nil {
		log.Fatal(err)
	}
}

func InsertPost(db *sql.DB, hash string, size int, tags []string) error {
	if size <= 0 {
		return errors.New("size is zero")
	}

	tx, err := db.Begin()
	if err != nil {
		return txError(tx, err)
	}

	r, err := tx.Exec("INSERT INTO posts(hash, size) VALUES($1, $2)", hash, size)
	if err != nil {
		return txError(tx, err)
	}

	i64, err := r.LastInsertId()
	if err != nil {
		return txError(tx, err)
	}

	postID := int(i64)

	for _, tag := range tags {
		err := mapTagToPost(tx, postID, tag)
		if err != nil {
			return txError(tx, err)
		}
	}

	return nil
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
	_, err = tx.Exec("INSERT INTO OR IGNORE post_tag_mapping(post_id, tag_id) VALUES($1, $2)", postID, tagID)

	return nil
}

func txError(tx *sql.Tx, err error) error {
	er := tx.Rollback()
	if err != nil {
		log.Fatal(er, err)
	}
	return err
}
