package main 

import(
	"fmt"
	"net/http"
	"io/ioutil"
    "os"
	"strings"
	"text/template"
)

type Song struct {
	Name string
	Path string
}

type File struct {
	Name string
	Directory string
}

type PlaylistsTemplateData struct {
	Playlists []string
	Ip string
}

type PlaylistTemplateData struct {
	Songs []Song
	Name string
	Ip string
}

func urlencode(input string) string {
	input = strings.Replace(input, "[", "%5B", -1)
	input = strings.Replace(input, "]", "%5D", -1)
	input = strings.Replace(input, " ", "%20", -1)
	input = strings.Replace(input, "(", "%28", -1)
	input = strings.Replace(input, ")", "%29", -1)
	input = strings.Replace(input, "&", "%26", -1)
	return input
}

func readFilesRecursive(dirname string) []File {
	results := make([]File, 0)
	
	files, _ := ioutil.ReadDir(dirname)
	for _, f := range files {
		if f.IsDir() {
			innerDir := dirname + "/" + f.Name()
			inner := readFilesRecursive(innerDir)
			for _, i := range inner {
				results = append(results, File{i.Name, i.Directory})
			}
		} else {
			results = append(results, File{f.Name(), dirname})
		}
	}

	return results
}

var ips = make(map[string]string)
var dir string = os.Getenv("STREAMER_PATH")

func main() {
	ips["TITAN"] = "localhost"
	ips["alx-glesys"] = "109.74.1.81"

	host, _ := os.Hostname()
	ip := ips[host]

	http.Handle("/listen/", http.StripPrefix("/listen/", http.FileServer(http.Dir(dir))))
	
	// Returns a feed of songs in a specific playlist.
	http.HandleFunc("/playlists/", func (w http.ResponseWriter, r *http.Request) {
		name := strings.Replace(r.URL.String(), "/playlists/", "", -1)
		songs := make([]Song, 0)
		files := readFilesRecursive(dir + "/" + name)
		for _, f := range files {
			name := strings.Replace(f.Name, "&", "&amp;", -1)
			genre := strings.Split(f.Directory, "/")
			path := genre[len(genre) - 1] + "/" + urlencode(f.Name)

			songs = append(songs, Song{name, path})
		}

		t := template.Must(template.ParseFiles(
			"templates/playlist.xml",
		))
		t.Execute(w, PlaylistTemplateData{songs, name, ip})
	})

	// Shows a list of all playlists.
	http.HandleFunc("/playlists", func (w http.ResponseWriter, r *http.Request) {
		playlists := make([]string, 0)
		files, _ := ioutil.ReadDir(dir)
		for _, f := range files {
			if f.IsDir() {
				playlists = append(playlists, f.Name())
			}
		}

		t := template.Must(template.ParseFiles(
			"templates/playlists.xml",
		))
		t.Execute(w, PlaylistsTemplateData{playlists, ip})
	})

	err := http.ListenAndServe(":8084", nil)
	if err != nil {
		fmt.Printf("Error: ", err)
	}
}