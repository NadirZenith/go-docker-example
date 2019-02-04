package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var tplPath = "./tpl"
var dirPages = "pages"
var dataPath = "./data"

//var templates = template.Must(template.ParseFiles(tplPath+"/layout.html", tplPath+"/index.html"))

var ulLinksTemplate = `
<h2>{{.title}}</h2>
<ul>
	{{range $url, $text := .}}
		<li><a href="{{$url}}">{{$text}}</a></li>
	{{end}}
</ul>`

var validPath = regexp.MustCompile("^/(new|edit|save|view)/([a-zA-Z0-9]+)$")

type Link struct {
	url  string
	text string
}

type Page struct {
	Title    string
	Body     []byte
	BodyHtml template.HTML
}

//func newPage(title string, body []byte, bodyHtml template.HTML) *Page {
//	return &Page{Title: title, Body: body, BodyHtml: bodyHtml}
//}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(dataPath+"/"+filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(dataPath + "/" + filename)
	if err != nil {
		return nil, err
	}
	//return &Page{Title: title, Body: template.HTML(body)}, nil
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	var templates = template.Must(template.New("").Funcs(template.FuncMap{
		"now": time.Now,
	}).ParseFiles(
		filepath.Join(tplPath, "parts/navigation.html"),
		filepath.Join(tplPath, "parts/footer.html"),
		filepath.Join(tplPath, "layout.html"),
		filepath.Join(tplPath, tmpl+".html"),
	))

	err := templates.ExecuteTemplate(w, "layout", p)

	if err != nil {
		// Log the detailed error
		log.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
	}
}
//func newHandler(w http.ResponseWriter, r *http.Request) {
//	//p, err := loadPage(title)
//	//if err != nil {
//		var p = &Page{Title: "New"}
//	//}
//	renderTemplate(w, "wiki/new", p)
//}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	//wtitle := r.FormValue("title")
	p := &Page{title, []byte(body), template.HTML("")}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "wiki/edit", p)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "wiki/view", p)
}

func getNotesList(buffer *bytes.Buffer) {
	// read data files
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	// build links array
	links := make(map[string]string)
	for _, f := range files {
		var file = strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		links["/view/"+file] = file
	}

	// create template from links
	t := template.New("t")
	t, err = t.Parse(ulLinksTemplate)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	err = t.Execute(buffer, links)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Here we will extract the page title from the Request,
		// and call the provided handler 'fn'
		log.Print(r.URL.Path)
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	var pathPage = filepath.Clean(r.URL.Path)
	var buffer bytes.Buffer
	var title = ""

	// for index, fill buffer with notes list
	if pathPage == "/" {
		pathPage = "/index"
		title = "Notes"
		getNotesList(&buffer)
	}
	pathPage = filepath.Join(dirPages, pathPage)

	// Return a 404 if the template doesn't exist
	var templateFilePath = filepath.Join(tplPath, pathPage)
	info, err := os.Stat(templateFilePath + ".html")
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}
	// Return a 404 if the request is for a directory
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	var page = &Page{Title: title, BodyHtml: template.HTML(buffer.Bytes())}

	renderTemplate(w, pathPage, page)
}

func startHttp() {
	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/", pageHandler)
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	//http.HandleFunc("/new", newHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	startHttp()
}
