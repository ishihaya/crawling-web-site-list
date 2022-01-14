package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Crawl(url string, depth int, m *message) {
	defer func() { m.quit <- 0 }()

	// WebページからURLを取得
	urls, err := Fetch(url)

	// 結果送信
	m.res <- &respons{
		url: url,
		err: err,
	}

	if err == nil {
		for _, url := range urls {
			// 新しいリクエスト送信
			m.req <- &request{
				url:   url,
				depth: depth - 1,
			}
		}
	}
}

func Fetch(u string) (urls []string, err error) {
	baseUrl, err := url.Parse(u)
	if err != nil {
		return
	}

	resp, err := http.Get(baseUrl.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 取得したhtmlを文字列で確認したい時
	// body, err := ioutil.ReadAll(resp.Body)
	// buf := bytes.NewBuffer(body)
	// html := buf.String()
	// fmt.Println(html)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	urls = make([]string, 0)
	// doc.Find(".r").Each(func(_ int, srg *goquery.Selection) {
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			reqURL, err := baseUrl.Parse(href)
			if err == nil {
				reqURLStr := reqURL.String()
				url := addURL(reqURLStr)
				if url != "" {
					urls = append(urls, url)
				}
			}
		}
	})
	// })

	return
}

// addURL - （条件に応じた）URLを追加する
func addURL(url string) string {
	// 暗号化されていないhttp通信のwebサイトのみ許容する
	if !strings.Contains(url, "https://www.google.com/url?q=http://") {
		return ""
	}
	// httpsも含めて探すとき
	// if !strings.Contains(url, "https://www.google.com/url?q=http") {
	// return ""
	// }
	// accounts.google.comが含まれてしまうのでそれ以外
	if strings.Contains(url, "https://accounts.google.com") {
		return ""
	}

	// 欲しいURL
	// 先頭と末尾を切り取る
	url = strings.Replace(url, "https://www.google.com/url?q=", "", 1)
	url = strings.Split(url, "&")[0]
	return url
}

func main() {
	var word = flag.String("w", " ", "検索ワードを入力して下さい")
	var start = flag.String("n", " ", "開始を入力してください")
	flag.Parse()
	// log.Println("検索ワード:", *word)
	*word = strings.Replace(*word, " ", "+", -1)
	firstURL := fmt.Sprintf("https://www.google.com/search?rlz=1C5CHFA_enJP962JP964&start=%s&q=%s", *start, string(*word))
	// log.Println("検索URL:", firstURL)
	m := newMessage()
	go m.execute()
	m.req <- &request{
		url:   firstURL,
		depth: 2,
	}
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndSearver:", err)
	}
}
