package crawl

import (
	"fmt"
	"os"
)

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
