package service

import (
	"context"
	"server/model/elasticsearch"
	"server/model/other"
	"server/model/request"
	"server/utils"

	"github.com/elastic/go-elasticsearch/v8/typedapi/eql/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
)

type ArticleService struct {
}

func (articleService *ArticleService) ArticleInfoByID(id string) (elasticsearch.Article, error) {
	go func() {
		articleViews := articleService.NewArticleView()
		articleViews.Set(id)
	}()
	return articleService.Get(id)
}
func (articleService *ArticleService) ArticleSearch(info request.ArticleSearch) (interface{}, int64, error) {
	req := &search.Request{
		Query: &types.Query{},
	}

	boolQuery := &types.BoolQuery{}

	if info.Query != "" {
		boolQuery.Should = []types.Query{
			{Match: map[string]types.MatchQuery{"title": {Query: info.Query}}},
			{Match: map[string]types.MatchQuery{"keyword": {Query: info.Query}}},
			{Match: map[string]types.MatchQuery{"abstract": {Query: info.Query}}},
			{Match: map[string]types.MatchQuery{"content": {Query: info.Query}}},
		}
	}

	if info.Tag != "" {
		boolQuery.Must = []types.Query{
			{Match: map[string]types.MatchQuery{"tags": {Query: info.Tag}}},
		}
	}

	if info.Category != "" {
		boolQuery.Filter = []types.Query{
			{Term: map[string]types.TermQuery{"category": {Value: info.Category}}},
		}
	}

	if boolQuery.Should != nil || boolQuery.Must != nil || boolQuery.Filter != nil {
		req.Query.Bool = boolQuery
	} else {
		req.Query.MatchAll = &types.MatchAllQuery{}
	}

	if info.Sort != "" {
		var sortField string
		switch info.Sort {
		case "time":
			sortField = "created_at"
		case "view":
			sortField = "views"
		case "comment":
			sortField = "comments"
		case "like":
			sortField = "likes"
		default:
			sortField = "created_at"
		}

		var order sortorder.SortOrder
		if info.Order != "asc" {
			order = sortorder.Desc
		} else {
			order = sortorder.Asc
		}

		req.Sort = []types.SortCombinations{
			types.SortOptions{
				SortOptions: map[string]types.FieldSort{
					sortField: {Order: &order},
				},
			},
		}
	}

	option := other.EsOption{
		PageInfo:       info.PageInfo,
		Index:          elasticsearch.ArticleIndex(),
		Request:        req,
		SourceIncludes: []string{"created_at", "cover", "title", "abstract", "category", "tags", "views", "comments", "likes"},
	}

	return utils.EsPagination(context.TODO(), option)
}
func (articleService *ArticleService) ArticleCategory() {

}
func (articleService *ArticleService) ArticleTags() {

}
func (articleService *ArticleService) ArticleLike() {

}
func (articleService *ArticleService) ArticleIsLike() {

}
func (articleService *ArticleService) ArticleLikesList() {

}
func (articleService *ArticleService) ArticleCreate() {

}
func (articleService *ArticleService) ArticleDelete() {

}
func (articleService *ArticleService) ArticleUpdate() {

}
func (articleService *ArticleService) ArticleList() {

}
