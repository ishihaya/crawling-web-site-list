package crawl

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Crawl(url string, depth int, m *Message) {
	defer func() { m.Quit <- 0 }()

	// WebページからURLを取得
	urls, err := Fetch(url)

	// 結果送信
	m.Res <- &Response{
		URL: url,
		Err: err,
	}

	if err == nil {
		for _, url := range urls {
			// 新しいリクエスト送信
			m.Req <- &Request{
				URL:   url,
				Depth: depth - 1,
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
				url := AddURL(reqURLStr)
				if url != "" {
					urls = append(urls, url)
				}
			}
		}
	})
	// })

	return
}

// AddURL - （条件に応じた）URLを追加する
func AddURL(url string) string {
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

type Message struct {
	Res  chan *Response
	Req  chan *Request
	Quit chan int
}
type Response struct {
	URL string
	Err interface{}
}
type Request struct {
	URL   string
	Depth int
}

func NewMessage() *Message {
	return &Message{
		Res:  make(chan *Response),
		Req:  make(chan *Request),
		Quit: make(chan int),
	}
}

func (m *Message) Execute() {
	// ワーカーの数
	wc := 0
	urlMap := make(map[string]bool, 100)
	done := false
	for !done {
		select {
		case res := <-m.Res:
			if res.Err == nil {
				fmt.Printf("%s\n", res.URL)
			} else {
				fmt.Fprintf(os.Stderr, "Error %s\n%v\n", res.URL, res.Err)
			}
		case req := <-m.Req:
			if req.Depth == 0 {
				break
			}

			if urlMap[req.URL] {
				// 取得済み
				break
			}
			urlMap[req.URL] = true

			wc++
			go Crawl(req.URL, req.Depth, m)
		case <-m.Quit:
			wc--
			if wc == 0 {
				done = true
			}
		}
	}
	// log.Println("スクレイピング完了")
	os.Exit(0)
}
