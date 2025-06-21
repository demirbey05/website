package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Post struct {
	Title    string
	Date     string
	Author   string
	Content  template.HTML
	FileName string
}

type PageData struct {
	PageTitle string
	Title     string
	Date      string
	Author    string
	Posts     []Post
	Body      template.HTML
}

var md goldmark.Markdown

func main() {
	// Configure Markdown parser with syntax highlighting
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
		),
	)

	// Set up HTTP handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post/", postHandler)

	// Start server
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := listPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("base").Parse(baseTemplate))
	tmpl = template.Must(tmpl.Parse(indexTemplate))

	data := PageData{
		PageTitle: "Academic Blog",
		Posts:     posts,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.TrimPrefix(r.URL.Path, "/post/")
	contentBytes, err := os.ReadFile(filepath.Join("posts", fileName))
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	contentStr := string(contentBytes)

	// Parse YAML front matter if present
	var author, date string
	if strings.HasPrefix(contentStr, "---") {
		parts := strings.SplitN(contentStr, "\n", -1)
		end := -1
		for i := 1; i < len(parts); i++ {
			if strings.TrimSpace(parts[i]) == "---" {
				end = i
				break
			}
		}
		if end != -1 {
			for _, l := range parts[1:end] {
				kv := strings.SplitN(l, ":", 2)
				if len(kv) != 2 {
					continue
				}
				key := strings.TrimSpace(kv[0])
				val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
				switch strings.ToLower(key) {
				case "author":
					author = val
				case "date":
					date = val
				case "title":
					// optional: override title
				}
			}
			contentStr = strings.Join(parts[end+1:], "\n")
		}
	}

	if err := md.Convert([]byte(contentStr), &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract date and title from filename (format: YYYY-MM-DD-title.md)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	parts := strings.SplitN(baseName, "-", 4)
	if len(parts) != 4 {
		http.Error(w, "Invalid post filename format", http.StatusInternalServerError)
		return
	}

	date = parts[0] + "-" + parts[1] + "-" + parts[2]
	caser := cases.Title(language.English)
	title := caser.String(strings.ReplaceAll(parts[3], "-", " "))

	post := Post{
		Title:    title,
		Date:     date,
		Author:   author,
		Content:  template.HTML(buf.String()),
		FileName: fileName,
	}

	tmpl := template.Must(template.New("base").Parse(baseTemplate))
	tmpl = template.Must(tmpl.Parse(postTemplate))

	data := PageData{
		PageTitle: "Academic Blog",
		Title:     post.Title,
		Date:      post.Date,
		Author:    post.Author,
		Body:      post.Content,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func listPosts() ([]Post, error) {
	files, err := os.ReadDir("posts")
	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".md" {
			baseName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			parts := strings.SplitN(baseName, "-", 4)
			if len(parts) != 4 {
				log.Printf("Skipping file with invalid format: %s", f.Name())
				continue
			}

			date := parts[0] + "-" + parts[1] + "-" + parts[2]
			caser := cases.Title(language.English)
			title := caser.String(strings.ReplaceAll(parts[3], "-", " "))

			posts = append(posts, Post{
				Title:    title,
				Date:     date,
				FileName: f.Name(),
			})
		}
	}

	return posts, nil
}

const baseTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Title}}</title>
	<style>
		{{template "css"}}
	</style>
	<script>
		MathJax = {
			tex: {
				inlineMath: [['\\(', '\\)']],
				displayMath: [['\\[', '\\]']],
				processEscapes: true,
			},
			options: {
				skipHtmlTags: ['script', 'noscript', 'style', 'textarea', 'pre']
			}
		};
	</script>
	<script src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-chtml.js"></script>
</head>
<body>
	<div class="container">
		<header>
			<h1><a href="/">{{.PageTitle}}</a></h1>
		</header>
		<main>
			{{template "content" .}}
		</main>
		<footer>
			<hr>
			<p>Academic Blog · Built with Go and ❤️</p>
		</footer>
	</div>
</body>
</html>`

const indexTemplate = `{{define "css"}}` + css + `{{end}}
{{define "content"}}
	<h2>Latest Posts</h2>
	<ul class="post-list">
	{{range .Posts}}
		<li>
			<a href="/post/{{.FileName}}">{{.Title}}</a>
			<time datetime="{{.Date}}">{{.Date}}</time>
		</li>
	{{end}}
	</ul>
{{end}}`

const postTemplate = `{{define "css"}}` + css + `{{end}}
{{define "content"}}
	<article>
		<header>
			<p class="meta"><time datetime="{{.Date}}">{{.Date}}</time> · <span class="author">{{.Author}}</span></p>
		</header>
		<div class="content">
			{{.Body}}
		</div>
	</article>
	<div class="back-link">
		<a href="/">← Back to all posts</a>
	</div>
{{end}}`

const css = `
body {
	font-family: "Computer Modern Serif", serif;
	line-height: 1.6;
	margin: 0;
	padding: 0;
	background-color: white;
	color: black;
}

.container {
	max-width: 800px;
	margin: 0 auto;
	padding: 2rem;
}

header {
	text-align: center;
	margin-bottom: 2rem;
}

h1, h2, h3 {
	font-weight: normal;
}

.post-list {
	list-style: none;
	padding: 0;
}

.post-list li {
	margin-bottom: 1rem;
}

.post-list a {
	text-decoration: none;
	color: black;
}

.post-list time {
	color: #666;
	font-size: 0.9em;
	margin-left: 1rem;
}

article header h1 {
	font-size: 2rem;
	margin-bottom: 0.5rem;
}

.content {
	margin-top: 2rem;
}

.back-link {
	margin-top: 2rem;
}

pre {
	background-color: #f5f5f5;
	padding: 1rem;
	overflow-x: auto;
}

code {
	font-family: "Computer Modern Typewriter", monospace;
}

hr {
	border: 0;
	border-top: 1px solid #ccc;
	margin: 2rem 0;
}

footer {
	text-align: center;
	margin-top: 4rem;
	color: #666;
}
.MathJax {
	font-size: 1.1em;
}

.math {
	text-align: center;
	margin: 1.5em 0;
}

/* Add some spacing around equations */
.mjx-chtml {
	padding: 10px 0;
}`
