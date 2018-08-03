package api

import (
	"errors"
	"fmt"
	"time"

	model "github.com/wangjun861205/dalianmodel"
	"github.com/wangjun861205/nbmysql"

	context "golang.org/x/net/context"
)

type Server struct {
	crawler *Crawler
}

func NewServer(headers map[string]string) (*Server, error) {
	c, err := NewCrawler(headers)
	if err != nil {
		return nil, err
	}
	return &Server{c}, nil
}

func (s *Server) Get(ctx context.Context, req *GetRequest) (*BookInfo, error) {
	bookInfos, err := model.QueryBookInfo("@ISBN = " + req.ISBN)
	if err != nil {
		return nil, err
	}
	if len(bookInfos) == 0 {
		info, err := s.crawler.Crawl(req.ISBN)
		if err != nil {
			return nil, err
		}
		return info, nil
	}
	bookInfo, err := toBookInfo(bookInfos[0])
	if err != nil {
		return nil, err
	}
	return bookInfo, nil
}

func (s *Server) Put(ctx context.Context, bookInfo *BookInfo) (*PutResponse, error) {
	bookInfoModel, err := toBookInfoModel(bookInfo)
	if err != nil {
		return nil, err
	}
	err = bookInfoModel.InsertOrUpdate()
	if err != nil {
		return nil, err
	}
	bookModels, err := model.QueryBook(fmt.Sprintf("@UniqueCode='%s'", bookInfo.UniqueCode))
	if err != nil {
		return nil, err
	}
	if len(bookModels) > 0 {
		return nil, fmt.Errorf("%s unique code already exists", bookInfo.UniqueCode)
	}
	bookModel, err := toBookModel(bookInfo)
	if err != nil {
		return nil, err
	}
	err = bookModel.Insert()
	if err != nil {
		return nil, err
	}
	tagList, err := bookInfoModel.TagByIsbn().All()
	if err != nil {
		return nil, err
	}
	for _, tagModel := range tagList {
		err := bookInfoModel.TagByIsbn().Remove(tagModel)
		if err != nil {
			return nil, err
		}
	}
	tagList = toTagModels(bookInfo.Tags)
	for _, tagModel := range tagList {
		err := tagModel.Insert()
		if err != nil && err != nbmysql.ErrDupKey {
			return nil, err
		}
		err = bookInfoModel.TagByIsbn().Add(tagModel)
		if err != nil {
			return nil, err
		}
	}
	return &PutResponse{}, nil
}

func toBookInfo(bookInfoModel *model.BookInfo) (*BookInfo, error) {
	var title string
	if bookInfoModel.Title != nil {
		title = *(bookInfoModel.Title)
	}
	var price int64
	if bookInfoModel.Price != nil {
		price = *(bookInfoModel.Price)
	}
	var author string
	if bookInfoModel.Author != nil {
		author = *(bookInfoModel.Author)
	}
	var publisher string
	if bookInfoModel.Publisher != nil {
		publisher = *(bookInfoModel.Publisher)
	}
	var series string
	if bookInfoModel.Series != nil {
		series = *(bookInfoModel.Series)
	}
	tags := make([]string, 0, 8)
	_tags, err := bookInfoModel.TagByIsbn().All()
	if err != nil {
		return nil, err
	}
	for _, tag := range _tags {
		tags = append(tags, *(tag.Tag))
	}
	var isbn string
	if bookInfoModel.Isbn != nil {
		isbn = *(bookInfoModel.Isbn)
	}
	var publishDate string
	if bookInfoModel.PublishDate != nil {
		publishDate = bookInfoModel.PublishDate.Format("2006-01-02")
	}
	var binding string
	if bookInfoModel.Binding != nil {
		binding = *(bookInfoModel.Binding)
	}
	var format string
	if bookInfoModel.Format != nil {
		format = *(bookInfoModel.Format)
	}
	var pages int64
	if bookInfoModel.Pages != nil {
		pages = *(bookInfoModel.Pages)
	}
	var wordCount int64
	if bookInfoModel.WordCount != nil {
		wordCount = *(bookInfoModel.WordCount)
	}
	var contentIntro string
	if bookInfoModel.ContentIntro != nil {
		contentIntro = *(bookInfoModel.ContentIntro)
	}
	var authorIntro string
	if bookInfoModel.AuthorIntro != nil {
		authorIntro = *(bookInfoModel.AuthorIntro)
	}

	var menu string
	if bookInfoModel.Menu != nil {
		menu = *(bookInfoModel.Menu)
	}

	return &BookInfo{
		Title:        title,
		Price:        price,
		Author:       author,
		Publisher:    publisher,
		Series:       series,
		Tags:         tags,
		ISBN:         isbn,
		PublishDate:  publishDate,
		Binding:      binding,
		Format:       format,
		Pages:        pages,
		WordCount:    wordCount,
		ContentIntro: contentIntro,
		AuthorIntro:  authorIntro,
		Menu:         menu,
	}, nil
}

func toBookInfoModel(bookInfo *BookInfo) (*model.BookInfo, error) {
	_publishDate, err := time.Parse("2006-01-02", bookInfo.PublishDate)
	if err != nil {
		return nil, err
	}
	return model.NewBookInfo(
		nbmysql.NewInt(0, true),
		nbmysql.NewString(bookInfo.Title, false),
		nbmysql.NewInt(bookInfo.Price, false),
		nbmysql.NewString(bookInfo.Author, false),
		nbmysql.NewString(bookInfo.Publisher, false),
		nbmysql.NewString(bookInfo.Series, false),
		nbmysql.NewString(bookInfo.ISBN, false),
		nbmysql.NewTime(_publishDate, false),
		nbmysql.NewString(bookInfo.Binding, false),
		nbmysql.NewString(bookInfo.Format, false),
		nbmysql.NewInt(bookInfo.Pages, false),
		nbmysql.NewInt(bookInfo.WordCount, false),
		nbmysql.NewString(bookInfo.ContentIntro, false),
		nbmysql.NewString(bookInfo.AuthorIntro, false),
		nbmysql.NewString(bookInfo.Menu, false),
	), nil
}

func toTagModels(tags []string) []*model.Tag {
	list := make([]*model.Tag, len(tags))
	for i, tag := range tags {
		list[i] = model.NewTag(nbmysql.NewInt(0, true), nbmysql.NewString(tag, false))
	}
	return list
}

func toBookModel(bookInfo *BookInfo) (*model.Book, error) {
	if bookInfo.UniqueCode == "" {
		return nil, errors.New("unique code can not be empty")
	}
	return model.NewBook(
		nbmysql.NewInt(0, true),
		nbmysql.NewString(bookInfo.ISBN, false),
		nbmysql.NewInt(bookInfo.Volume, false),
		nbmysql.NewString(bookInfo.UniqueCode, false)), nil
}
