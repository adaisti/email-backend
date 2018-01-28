package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Email struct {
	Date       string
	Sender     string
	Recipients []string
	Cc         []string
	Bcc        []string
	Subject    string
	Text       string
}

func write_matching_emails(compare_email func(Email, string) bool, form_field string, emails []Email, w http.ResponseWriter, r *http.Request) {
	query, ok := r.Form[form_field]
	if ok {
		querystring := query[0]
		var found_emails []Email

		for _, email := range emails {
			if compare_email(email, querystring) {
				found_emails = append(found_emails, email)
			}
		}

		json.NewEncoder(w).Encode(found_emails)
	}
}

func sender_is(email Email, querystring string) bool {
	return email.Sender == querystring
}

func recipient_is(email Email, querystring string) bool {
	for _, recipient := range email.Recipients {
		if recipient == querystring {
			return true
		}
	}
	return false
}

func contains_substring(email Email, querystring string) bool {
	querystring = strings.ToLower(querystring)

	for _, recipient := range email.Recipients {
		if recipient == querystring {
			return true
		}
	}

	for _, recipient := range email.Cc {
		if recipient == querystring {
			return true
		}
	}

	for _, recipient := range email.Bcc {
		if recipient == querystring {
			return true
		}
	}

	return strings.Contains(strings.ToLower(email.Text), querystring) ||
		strings.Contains(email.Sender, querystring) ||
		strings.Contains(strings.ToLower(email.Subject), querystring)
}

func main() {

	var emails []Email

	filehandle, _ := os.Open("enron/enron.json")

	scanner := bufio.NewScanner(filehandle)

	for scanner.Scan() {

		jsonBlob := scanner.Bytes()

		var email Email

		err := json.Unmarshal(jsonBlob, &email)

		if err != nil {
			fmt.Println("error:", err)
		}

		emails = append(emails, email)

	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		r.ParseForm()

		write_matching_emails(sender_is, "sentby", emails, w, r)
		write_matching_emails(recipient_is, "recvby", emails, w, r)
		write_matching_emails(contains_substring, "fulltext", emails, w, r)

	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
