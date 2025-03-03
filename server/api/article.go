package api

import (
	"server/global"
	"server/model/request"
	"server/model/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ArticleApi struct {
}

func (articleApi *ArticleApi) ArticleInfoByID(c *gin.Context) {
	var req request.ArticleInfoByID
	err := c.ShouldBindUri(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	article, err := articleService.ArticleInfoByID(req.ID)
	if err != nil {
		global.Log.Error("Failed to get article information:", zap.Error(err))
		response.FailWithMessage("Failed to get article information", c)
		return
	}

	response.OkWithData(article, c)
}
func (articleApi *ArticleApi) ArticleSearch(c *gin.Context) {
	var info request.ArticleSearch
	err := c.ShouldBindQuery(&info)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, total, err := articleService.ArticleSearch(info)

	if err != nil {
		global.Log.Error("Failed to get article search result:", zap.Error(err))
		response.FailWithMessage("Failed to get article search result", c)
		return
	}
	response.OkWithData(response.PageResult{
		List:  list,
		Total: total,
	}, c)
}
func (articleApi *ArticleApi) ArticleCategory(c *gin.Context) {

}
func (articleApi *ArticleApi) ArticleTags(c *gin.Context) {

}
func (articleApi *ArticleApi) ArticleLike(c *gin.Context) {

}
func (articleApi *ArticleApi) ArticleIsLike(c *gin.Context) {

}
func (articleApi *ArticleApi) ArticleLikesList(c *gin.Context) {

}
func (articleApi *ArticleApi) ArticleCreate(c *gin.Context) {

}
func (articleApi *ArticleApi) ArticleDelete(c *gin.Context) {

}
func (articleApi *ArticleApi) ArticleUpdate(c *gin.Context) {

}
func (articleApi *ArticleApi) ArticleList(c *gin.Context) {

}
