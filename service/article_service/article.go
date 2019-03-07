package article_service

import (
	"encoding/json"

	"github.com/EDDYCJY/go-gin-example/models"
	"github.com/EDDYCJY/go-gin-example/pkg/gredis"
	"github.com/EDDYCJY/go-gin-example/pkg/logging"
	"github.com/EDDYCJY/go-gin-example/service/cache_service"
)

type Article struct {
	ID            int
	TagID         int
	Title         string
	Desc          string
	Content       string
	CoverImageUrl string
	State         int
	CreatedBy     string
	ModifiedBy    string

	PageNum  int
	PageSize int
}

var returnError = func(action func(article *models.Article) error) error {
	model := models.NewArticle(models.Open())
	defer model.Dispose()
	return action(model)
}

var returnValue = func(action func(article *models.Article) (interface{}, error)) (interface{}, error) {
	model := models.NewArticle(models.Open())
	defer model.Dispose()
	return action(model)
}

func (a *Article) Add() error {
	return returnError(func(model *models.Article) error {
		article := map[string]interface{}{
			"tag_id":          a.TagID,
			"title":           a.Title,
			"desc":            a.Desc,
			"content":         a.Content,
			"created_by":      a.CreatedBy,
			"cover_image_url": a.CoverImageUrl,
			"state":           a.State,
		}
		if err := model.AddArticle(article); err != nil {
			return err
		}

		return nil
	})
}

func (a *Article) Edit() error {
	return returnError(func(model *models.Article) error {
		return model.EditArticle(a.ID, map[string]interface{}{
			"tag_id":          a.TagID,
			"title":           a.Title,
			"desc":            a.Desc,
			"content":         a.Content,
			"cover_image_url": a.CoverImageUrl,
			"state":           a.State,
			"modified_by":     a.ModifiedBy,
		})
	})
}

func (a *Article) Get() (*models.Article, error) {
	value, e := returnValue(func(model *models.Article) (i interface{}, e error) {
		var cacheArticle *models.Article

		cache := cache_service.Article{ID: a.ID}
		key := cache.GetArticleKey()
		if gredis.Exists(key) {
			data, err := gredis.Get(key)
			if err != nil {
				logging.Info(err)
			} else {
				json.Unmarshal(data, &cacheArticle)
				return cacheArticle, nil
			}
		}

		article, err := model.GetArticle(a.ID)
		if err != nil {
			return nil, err
		}

		gredis.Set(key, article, 3600)
		return article, nil
	})
	return value.(*models.Article), e
}

func (a *Article) GetAll() ([]*models.Article, error) {
	value, e := returnValue(func(model *models.Article) (i interface{}, e error) {

		var (
			articles, cacheArticles []*models.Article
		)

		cache := cache_service.Article{
			TagID: a.TagID,
			State: a.State,

			PageNum:  a.PageNum,
			PageSize: a.PageSize,
		}
		key := cache.GetArticlesKey()
		if gredis.Exists(key) {
			data, err := gredis.Get(key)
			if err != nil {
				logging.Info(err)
			} else {
				json.Unmarshal(data, &cacheArticles)
				return cacheArticles, nil
			}
		}

		articles, err := model.GetArticles(a.PageNum, a.PageSize, a.getMaps())
		if err != nil {
			return nil, err
		}

		gredis.Set(key, articles, 3600)
		return articles, nil
	})
	return value.([]*models.Article), e
}

func (a *Article) Delete() error {
	return returnError(func(model *models.Article) error {
		return model.DeleteArticle(a.ID)
	})
}

func (a *Article) ExistByID() (bool, error) {
	value, e := returnValue(func(model *models.Article) (i interface{}, e error) {
		return model.ExistArticleByID(a.ID)
	})
	return value.(bool), e
}

func (a *Article) Count() (int, error) {
	value, e := returnValue(func(model *models.Article) (i interface{}, e error) {
		return model.GetArticleTotal(a.getMaps())
	})
	return value.(int), e
}

func (a *Article) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	maps["deleted_on"] = 0
	if a.State != -1 {
		maps["state"] = a.State
	}
	if a.TagID != -1 {
		maps["tag_id"] = a.TagID
	}

	return maps
}
