package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/smtp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	yaml "gopkg.in/yaml.v2"
)

type conf struct {
	Database string `yaml:"database"`
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}
type email struct {
	to      []string
	subject string
	body    string
}

func (c *conf) getConf() error {
	yamlFile, err := ioutil.ReadFile("config.yaml")

	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return err
}

func newEmail(to []string, subject string) *email {
	return &email{to: to, subject: subject}
}

func (e *email) parseTemplate(fileName string, data interface{}) error {
	t, err := template.ParseFiles(fileName)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, data); err != nil {
		return err
	}
	e.body = buffer.String()
	return nil
}

func (e *email) sendMail() bool {

	var MIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := "From:" + c.Email + "\r\nTo: " + e.to[0] + "\r\nSubject: " + e.subject + "\r\n" + MIME + "\r\n" + e.body
	SMTP := fmt.Sprintf("%s:%d", c.Server, c.Port)

	if err := smtp.SendMail(SMTP, smtp.PlainAuth("", c.Email, c.Password, c.Server), c.Email, e.to, []byte(body)); err != nil {
		return false
	}

	return true
}

func (e *email) Send(templateName string, items interface{}) {
	err := e.parseTemplate(templateName, items)
	if err != nil {
		log.Fatal(err)
	}
	if ok := e.sendMail(); ok {
		log.Printf("Email has been sent to %s\n", e.to)
	} else {
		log.Printf("Failed to send the email to %s\n", e.to)
	}
}

func split(email string) string {
	i := strings.LastIndexByte(email, '@')
	return email[i+1:]
}

var c conf

func init() {
	c.getConf()
}

func main() {

	var (
		id     int
		email  string
		host   string
		ids    []int
		email2 []string
	)

	db, err := sql.Open("sqlite3", c.Database)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT id, email FROM email where valid = 1")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err = rows.Scan(&id, &email)
		if err != nil {
			log.Fatal(err)
		}

		host = split(email)

		if _, err := net.LookupMX(host); err != nil {
			if _, err := net.LookupIP(host); err != nil {

				ids = append(ids, id)
				log.Printf("Failed Update err   #%v ", err)
			}
		} else {

			email2 = append(email2, email)

		}
	}

	rows.Close()
	db.Close()

	db, err = sql.Open("sqlite3", c.Database)
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("update email set valid = 0 where id=?")
	if err != nil {
		log.Printf("Failed Update err   #%v ", err)
	} else {

		for i := range ids {

			_, err = stmt.Exec(ids[i])
			if err != nil {
				log.Println(err)
			}
		}

	}
	stmt.Close()
	db.Close()

	subject := "Custodia de documentos scanchile"

	e := newEmail(email2, subject)
	e.Send("newsletter.html", map[string]string{"username": "usted estimado cliente."})

}
