package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ishihaya/crawling-web-site-list/crawl"
)

func main() {
	var word = flag.String("w", " ", "検索ワードを入力して下さい")
	var start = flag.String("n", " ", "開始を入力してください")
	flag.Parse()
	// log.Println("検索ワード:", *word)
	*word = strings.Replace(*word, " ", "+", -1)
	firstURL := fmt.Sprintf("https://www.google.com/search?rlz=1C5CHFA_enJP962JP964&start=%s&q=%s", *start, string(*word))
	// log.Println("検索URL:", firstURL)
	m := crawl.NewMessage()
	go m.Execute()
	m.Req <- &crawl.Request{
		URL:   firstURL,
		Depth: 2,
	}
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndSearver:", err)
	}
}
