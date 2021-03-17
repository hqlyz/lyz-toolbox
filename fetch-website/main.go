package main

import (
	"fmt"
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

var (
	urlProtocol = "http://"
	rootDir     = ""
)

type staticFile struct {
	fileURL  string
	filePath string
	slash    bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: fetch-website some-url")
		return
	}
	webURL := os.Args[1]
	if strings.Contains(webURL, "https") {
		urlProtocol = "https://"
	}
	u, err := url.Parse(webURL)
	// 确保传入的URL是合法的URL
	if err != nil {
		log.Fatal(err)
	}

	// 新建文件夹存储所有文件
	rootDir = u.Host
	if !fileExists(rootDir) {
		err = os.Mkdir(rootDir, 0666)
		if err != nil {
			log.Fatal(err)
		}
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

	// var buf bytes.Buffer
	// tee := io.TeeReader(resp.Body, &buf)

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
	for _, s := range staticFiles {
		log.Println(s.fileURL, path.Join(rootDir, s.filePath))
		go downloadStaticFile(s.fileURL, path.Join(rootDir, s.filePath), &wg)
	}
	wg.Wait()
	log.Println("All done.")
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

func downloadStaticFile(fileURL string, filePath string, wg *sync.WaitGroup) {
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
	if !fileExists(dir) {
		err = os.MkdirAll(dir, 0666)
		if err != nil {
			return
		}
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
