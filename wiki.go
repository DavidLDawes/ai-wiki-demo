// Copyright 2024 David L. Dawes. All rights reserved.

package main

import (
	"context"
	"fmt"
	"github.com/madebywelch/anthropic-go/v3/pkg/anthropic"
	"github.com/madebywelch/anthropic-go/v3/pkg/anthropic/client/native"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save(log *log.Logger) error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string, log *log.Logger) (*Page, error) {
	fmt.Println("loadPage")
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {

		return nil, err

	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	logger.Println("viewHandler entered")
	p, err := loadPage(title, logger)
	if err != nil {
		// no previous page found, get one from Anthropic via API
		// Prepare a message request
		logger.Println("viewHandler: No previous page found")
		request := anthropic.NewMessageRequest(
			[]anthropic.MessagePartRequest{
				{Role: "user",
					Content: []anthropic.ContentBlock{anthropic.NewTextContentBlock("define " + title +
						" with a short sentence, no introduction")}}},
			anthropic.WithModel[anthropic.MessageRequest](anthropic.ClaudeV2_1),
			anthropic.WithMaxTokens[anthropic.MessageRequest](60),
		)

		// Call the Message method
		logger.Println("viewHandler: call the Anthropic message client")
		response, err := client.Message(ctx, request)
		if err != nil {
			logger.Println("viewHandler: Exception attempting to use Anthropic for AI definition.")
			logger.Println("editHandler: handle this as a new record, so edit it")

			http.Redirect(w, r, "/edit/"+title, http.StatusFound)
			return
		} else {
			if response != nil {
				logger.Println("viewHandler: got content ==>", response.Content)
				p = &Page{Title: title, Body: []byte(response.Content[0].Text)}
			} else {
				logger.Println("viewHandler: Empty result using Anthropic for AI definition.")
				logger.Println("editHandler: handle this as a new record, so edit it")

				http.Redirect(w, r, "/edit/"+title, http.StatusFound)
				return
			}
		}
	}
	logger.Println("viewHandler: rendering template with result")
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	logger.Println("editHandler")
	p, err := loadPage(title, logger)
	if err != nil {
		logger.Println("editHandler: loadpage error, using empty definition")
		p = &Page{Title: title}
	}
	logger.Println("editHandler: rendering template with result")
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	logger.Println("saveHandler")
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save(logger)
	if err != nil {
		logger.Println("saveHandler: failed to save, returning HTTP internal service error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)

	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

var ctx context.Context
var client *native.Client

func main() {
	var err error
	var apiKey = ""
	ctx = context.Background()

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == "ANTHROPICAPIKEY" {
			apiKey = pair[1]
		}
	}

	client, err = native.MakeClient(native.Config{
		APIKey: apiKey,
	})
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
