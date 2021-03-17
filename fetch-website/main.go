package main

import (
	"bytes"
	"fmt"
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
	rootDir = ""
)

type staticFile struct {
	fileURL  string
	filePath string
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

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	doc, err := goquery.NewDocumentFromReader(tee)
	if err != nil {
		log.Fatal(err)
	}

	staticFiles := []staticFile{}

	doc.Find("link, script").Each(func(i int, s *goquery.Selection) {
		href, exist := s.Attr("href")
		if exist {
			fileURL := getStaticFilePath(urlProtocol, u, href)
			if fileURL != "" {
				staticFiles = append(staticFiles, staticFile{fileURL: fileURL, filePath: href})
			}
		} else {
			src, exist := s.Attr("src")
			if exist {
				fileURL := getStaticFilePath(urlProtocol, u, src)
				if fileURL != "" {
					staticFiles = append(staticFiles, staticFile{fileURL: fileURL, filePath: src})
				}
			}
		}
	})	

	contentBytes, err := ioutil.ReadAll(&buf)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(path.Join(rootDir, "index.html"), contentBytes, 0666)
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

func getStaticFilePath(protocol string, u *url.URL, src string) string {
	if len(src) == 0 {
		return ""
	}
	if len(src) >= 4 && src[:4] == "http" {
		return ""
	}
	if src[0] == '/' {
		return protocol + path.Join(u.Host, src)
	}
	return protocol + path.Join(u.Host, u.Path, src)
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
