package api

import (
	"compress/gzip"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/wangjun861205/nbsoup"
	"golang.org/x/net/publicsuffix"
)

const RootURL = "https://www.dushu.com/"
const SearchURL = "https://www.dushu.com/search.aspx?wd=%s"

type Crawler struct {
	*http.Client
	headers map[string]string
	baseURL *url.URL
}

func NewCrawler(headers map[string]string) (*Crawler, error) {
	defaultHeaders := map[string]string{
		"Accept":          "image/webp,image/apng,image/*,*/*;q=0.8",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Cache-Control":   "no-cache",
		"Connection":      "keep-alive",
		"Pragma":          "no-cache",
		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Safari/537.36",
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}
	for k, v := range headers {
		defaultHeaders[k] = v
	}
	baseURL, err := url.Parse(RootURL)
	if err != nil {
		return nil, err
	}
	return &Crawler{client, defaultHeaders, baseURL}, nil
}

func (c *Crawler) search(isbn string) (string, error) {
	reqURL := fmt.Sprintf(SearchURL, isbn)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()
	b, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return "", err
	}
	root, err := nbsoup.Parse(b)
	if err != nil {
		return "", err
	}
	books, err := root.FindAll(`div[class="book-info"].h3.a`)
	if err != nil {
		return "", err
	}
	if len(books) == 0 {
		return "", fmt.Errorf("no book for isbn %s", isbn)
	}
	relURL, err := url.Parse(books[0].AttrMap["href"])
	if err != nil {
		return "", err
	}
	return c.baseURL.ResolveReference(relURL).String(), nil
}

// type BookInfo struct {
// 	Price        int64     `json:"price"`
// 	Author       string    `json:"author"`
// 	Publisher    string    `json:"publisher"`
// 	Series       string    `json:"series"`
// 	Tags         string    `json:"tags`
// 	ISBN         string    `json:"isbn"`
// 	PublishDate  time.Time `json:"publish_date"`
// 	Binding      string    `json:"binding"`
// 	Format       string    `json:"format"`
// 	Pages        int64     `json:"pages`
// 	WordCount    int64     `json:"word_count`
// 	ContentIntro string    `json:"content_intro`
// 	AuthorIntro  string    `json:"authro_intro"`
// 	Menu         string    `json:"menu"`
// }

func findTitle(node *nbsoup.Node) (string, error) {
	hs, err := node.FindAll(`div[class="book-title"].h1`)
	if err != nil {
		return "", err
	}
	if len(hs) == 0 {
		return "", nil
	}
	return escapeAndTrim(hs[0].Content), nil

}

func findPrice(node *nbsoup.Node) (int64, error) {
	prices, err := node.FindAll(`p[class="price"].span[class="num"]`)
	if err != nil {
		return 0, err
	}
	if len(prices) == 0 {
		return 0, nil
	}
	priceStr := escapeAndTrim(prices[0].Content)
	if priceStr == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(strings.Trim(priceStr, "¥"), 64)
	if err != nil {
		return 0, err
	}
	return int64(f * 100), nil
}

func findAuthor(node *nbsoup.Node) (string, error) {
	titleTds, err := node.FindAll(`td[@content="作　者："]`)
	if err != nil {
		return "", err
	}
	if len(titleTds) == 0 {
		return "", nil
	}
	return escapeAndTrim(titleTds[0].Next.Content), nil
}

func findPublisher(node *nbsoup.Node) (string, error) {
	titleTds, err := node.FindAll(`td[@content="出版社："]`)
	if err != nil {
		return "", err
	}
	if len(titleTds) == 0 {
		return "", nil
	}
	publisherAs, err := titleTds[0].Next.FindAll(`a`)
	if err != nil {
		return "", err
	}
	if len(publisherAs) == 0 {
		return "", nil
	}
	return escapeAndTrim(publisherAs[0].Content), nil

}

func findSeries(node *nbsoup.Node) (string, error) {
	titleTds, err := node.FindAll(`td[@content="丛编项："]`)
	if err != nil {
		return "", err
	}
	if len(titleTds) == 0 {
		return "", nil
	}
	return escapeAndTrim(titleTds[0].Next.Content), nil
}

func findTags(node *nbsoup.Node) ([]string, error) {
	titleTds, err := node.FindAll(`td[@content="标　签："]`)
	if err != nil {
		return nil, err
	}
	if len(titleTds) == 0 {
		return nil, nil
	}
	return strings.Split(escapeAndTrim(titleTds[0].Next.Content), ","), nil
}

func findISBN(node *nbsoup.Node) (string, error) {
	titleTds, err := node.FindAll(`td[@content="ISBN："]`)
	if err != nil {
		return "", err
	}
	if len(titleTds) == 0 {
		return "", nil
	}
	return escapeAndTrim(titleTds[0].Next.Content), nil
}

func findPublishDate(node *nbsoup.Node) (string, error) {
	titleTds, err := node.FindAll(`td[@content="出版时间："]`)
	if err != nil {
		return "", err
	}
	if len(titleTds) == 0 {
		return "", nil
	}
	return escapeAndTrim(titleTds[0].Next.Content), nil
}

func findBinding(node *nbsoup.Node) (string, error) {
	titleTds, err := node.FindAll(`td[@content="包装："]`)
	if err != nil {
		return "", err
	}
	if len(titleTds) == 0 {
		return "", nil
	}
	return escapeAndTrim(titleTds[0].Next.Content), nil
}

func findFormat(node *nbsoup.Node) (string, error) {
	titleTds, err := node.FindAll(`td[@content="开本："]`)
	if err != nil {
		return "", err
	}
	if len(titleTds) == 0 {
		return "", nil
	}
	return escapeAndTrim(titleTds[0].Next.Content), nil
}

func findPages(node *nbsoup.Node) (int64, error) {
	titleTds, err := node.FindAll(`td[@content="页数："]`)
	if err != nil {
		return 0, err
	}
	if len(titleTds) == 0 {
		return 0, nil
	}
	pagesStr := escapeAndTrim(titleTds[0].Next.Content)
	if pagesStr == "" {
		return 0, nil
	}
	return strconv.ParseInt(pagesStr, 10, 64)
}

func findWordCount(node *nbsoup.Node) (int64, error) {
	titleTds, err := node.FindAll(`td[@content="字数："]`)
	if err != nil {
		return 0, err
	}
	if len(titleTds) == 0 {
		return 0, nil
	}
	wcStr := escapeAndTrim(titleTds[0].Next.Content)
	if wcStr == "" {
		return 0, nil
	}
	return strconv.ParseInt(wcStr, 10, 64)
}

func findContentIntro(node *nbsoup.Node) (string, error) {
	hs, err := node.FindAll(`h4[@content="内容简介"]`)
	if err != nil {
		return "", err
	}
	if len(hs) == 0 {
		return "", nil
	}
	return escapeAndTrim(hs[0].Next.Children[0].Content), nil
}

func findAuthorIntro(node *nbsoup.Node) (string, error) {
	hs, err := node.FindAll(`h4[@content="作者简介"]`)
	if err != nil {
		return "", err
	}
	if len(hs) == 0 {
		return "", nil
	}
	return escapeAndTrim(hs[0].Next.Children[0].Content), nil
}

func findMenu(node *nbsoup.Node) (string, error) {
	hs, err := node.FindAll(`h4[@content="图书目录"]`)
	if err != nil {
		return "", err
	}
	if len(hs) == 0 {
		return "", nil
	}
	return escapeAndTrim(hs[0].Next.Children[0].Content), nil
}

func (c *Crawler) getBookInfo(bookURL string) (*BookInfo, error) {
	req, err := http.NewRequest("GET", bookURL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()
	b, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}
	root, err := nbsoup.Parse(b)
	if err != nil {
		return nil, err
	}
	_title, err := findTitle(root)
	if err != nil {
		return nil, err
	}
	_price, err := findPrice(root)
	if err != nil {
		return nil, err
	}
	_author, err := findAuthor(root)

	if err != nil {
		return nil, err
	}
	_publisher, err := findPublisher(root)
	if err != nil {
		return nil, err
	}
	_series, err := findSeries(root)
	if err != nil {
		return nil, err
	}
	_tags, err := findTags(root)
	if err != nil {
		return nil, err
	}
	_isbn, err := findISBN(root)
	if err != nil {
		return nil, err
	}
	_publishDate, err := findPublishDate(root)
	if err != nil {
		return nil, err
	}
	_binding, err := findBinding(root)
	if err != nil {
		return nil, err
	}
	_format, err := findFormat(root)
	if err != nil {
		return nil, err
	}
	_pages, err := findPages(root)
	if err != nil {
		return nil, err
	}
	_wordCount, err := findWordCount(root)
	if err != nil {
		return nil, err
	}
	_contentIntro, err := findContentIntro(root)
	if err != nil {
		return nil, err
	}
	_authorIntro, err := findAuthorIntro(root)
	if err != nil {
		return nil, err
	}
	_menu, err := findMenu(root)
	if err != nil {
		return nil, err
	}
	return &BookInfo{
		Title:        _title,
		Price:        _price,
		Author:       _author,
		Publisher:    _publisher,
		Series:       _series,
		Tags:         _tags,
		ISBN:         _isbn,
		PublishDate:  _publishDate,
		Binding:      _binding,
		Format:       _format,
		Pages:        _pages,
		WordCount:    _wordCount,
		ContentIntro: _contentIntro,
		AuthorIntro:  _authorIntro,
		Menu:         _menu,
	}, nil
}

func (c *Crawler) Crawl(isbn string) (*BookInfo, error) {
	u, err := c.search(isbn)
	if err != nil {
		return nil, err
	}
	return c.getBookInfo(u)
}

func escapeAndTrim(s string) string {
	return strings.Trim(html.EscapeString(s), " \n\r\u00a0")
}
