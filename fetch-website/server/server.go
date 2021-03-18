package server

import (
	"context"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	maxServerProcess = 10
)

var (
	urlProtocol = "http://"
	rootDir     = ""
)

type staticFile struct {
	fileURL  string
	filePath string
	slash    bool
}

// Server object is going to download website
type Server struct {
	taskList []string
	Ctx context.Context
}

// New function construct a new Server instance
func New(ctx context.Context) *Server {
	return &Server{
		taskList: []string{},
		Ctx: ctx,
	}
}

// Enqueue function appends newly url to download
func (s *Server) Enqueue(url string) {
	s.taskList = append(s.taskList, url)
}

func (s *Server) dequeue() string {
	if len(s.taskList) == 0 {
		return ""
	}
	ret := s.taskList[0]
	s.taskList = s.taskList[1:]
	return ret
}

// Run function means server is going to handle processes
func (s *Server) Run() {
	go s.handleProcess()
}

func (s *Server) handleProcess() {
	for {
		webURL := s.dequeue()
		if webURL == "" {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		go s.downloadWebsite(webURL)
	}
}

func (s *Server) downloadWebsite(webURL string) {
	if strings.Contains(webURL, "https") {
		urlProtocol = "https://"
	}

	// Make sure the website url is valid
	u, err := url.Parse(webURL)
	if err != nil {
		log.Printf("%s is not a valid URL: %v\n", webURL, err)
		return
	}

	// Create a folder to save whole website
	rootDir = u.Host
	err = createFolder(rootDir, 0666)
	if err != nil {
		return
	}

	log.Println("Start fetching website...")
	cli := &http.Client{
		Timeout: time.Second * 5,
	}
	// construct a http request
	req, err := http.NewRequest("GET", webURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Fetch website failed with status: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	staticFiles := []staticFile{}

	doc.Find("link, script").Each(func(i int, s *goquery.Selection) {
		href, exist := s.Attr("href")
		if exist {
			fileURL, slash := getStaticFileURL(urlProtocol, u, href)
			if fileURL != "" {
				staticFiles = append(staticFiles, staticFile{fileURL: fileURL, filePath: href, slash: slash})
				if slash {
					s.SetAttr("href", href[1:])
				}
			}
		} else {
			src, exist := s.Attr("src")
			if exist {
				fileURL, slash := getStaticFileURL(urlProtocol, u, src)
				if fileURL != "" {
					staticFiles = append(staticFiles, staticFile{fileURL: fileURL, filePath: src, slash: slash})
					if slash {
						s.SetAttr("src", src[1:])
					}
				}
			}
		}
	})

	finalDoc, err := doc.Html()
	if err != nil {
		log.Print(err)
	}
	finalDoc = html.UnescapeString(finalDoc)

	err = ioutil.WriteFile(path.Join(rootDir, "index.html"), []byte(finalDoc), 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Need to be downloaded...")
	var wg sync.WaitGroup
	wg.Add(len(staticFiles))
	for _, sf := range staticFiles {
		log.Println(sf.fileURL, path.Join(rootDir, sf.filePath))
		go s.downloadStaticFile(sf.fileURL, path.Join(rootDir, sf.filePath), &wg)
	}
	wg.Wait()
}

func (s *Server) downloadStaticFile(fileURL string, filePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(filePath) == 0 {
		return
	}
	resp, err := http.Get(fileURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	fp := filePath
	dir := path.Dir(fp)
	err = createFolder(dir, 0666)
	if err != nil {
		return
	}

	file, err := os.Create(fp)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	io.Copy(file, resp.Body)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func getStaticFileURL(protocol string, u *url.URL, src string) (fileURL string, slash bool) {
	if len(src) == 0 || (len(src) >= 4 && src[:4] == "http") {
		fileURL = ""
		slash = false
		return
	}
	if src[0] == '/' {
		fileURL = protocol + path.Join(u.Host, src)
		slash = true
		return
	}
	fileURL = protocol + path.Join(u.Host, u.Path, src)
	slash = false
	return
}

func createFolder(dir string, perm os.FileMode) error {
	if !fileExists(dir) {
		return os.MkdirAll(dir, 0666)
	}
	return nil
}