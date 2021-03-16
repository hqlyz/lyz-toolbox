package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	urlProtocol = "http://"
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
	rootDir := u.Host
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

	var buf = bufio.NewReader(resp.Body)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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

	log.Println("Need to be downloaded...")
	for _, s := range staticFiles {
		log.Println(s.fileURL, s.filePath)
	}

	buf.Reset(resp.Body)
	contentBytes, err := ioutil.ReadAll(buf)
	log.Println("content length: ", len(contentBytes))
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(path.Join(rootDir, "index.html"), contentBytes, 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func getStaticFilePath(protocol string, u *url.URL, src string) string {
	if len(src) == 0 {
		return ""
	}
	if src[0] == '/' {
		return path.Join(protocol, u.Host, src)
	}
	return path.Join(protocol, u.Host, u.Path, src)
}

func downloadStaticFile(fileURL string, filePath string) {

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