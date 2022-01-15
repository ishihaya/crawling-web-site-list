package crawl

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Crawl - クロール
func Crawl(url string, depth int, m *Message) {
	defer func() { m.Quit <- 0 }()

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

// Fetch - WebページからURLを取得する
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
