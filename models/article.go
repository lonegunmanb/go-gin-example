package models

import "github.com/jinzhu/gorm"

type Article struct {
	Model

	TagID int `json:"tag_id" gorm:"index"`
	Tag   Tag `json:"tag"`

	Title         string `json:"title"`
	Desc          string `json:"desc"`
	Content       string `json:"content"`
	CoverImageUrl string `json:"cover_image_url"`
	CreatedBy     string `json:"created_by"`
	ModifiedBy    string `json:"modified_by"`
	State         int    `json:"state"`
	db            IDB
}

func NewArticle(db IDB) *Article {
	if !db.Connected() {
		panic("broken db")
	}
	return &Article{
		db: db,
	}
}

func (a *Article) Dispose() error {
	return a.db.Close()
}

func (a *Article) ExistArticleByID(id int) (bool, error) {
	var article Article
	err := a.db.Select("id").Where("id = ? AND deleted_on = ? ", id, 0).First(&article).Error()
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	if article.ID > 0 {
		return true, nil
	}

	return false, nil
}

func (a *Article) GetArticleTotal(maps interface{}) (int, error) {
	var count int
	if err := a.db.Model(&Article{}).Where(maps).Count(&count).Error(); err != nil {
		return 0, err
	}

	return count, nil
}

func (a *Article) GetArticles(pageNum int, pageSize int, maps interface{}) ([]*Article, error) {
	var articles []*Article
	err := a.db.Preload("Tag").Where(maps).Offset(pageNum).Limit(pageSize).Find(&articles).Error()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return articles, nil
}

func (a *Article) GetArticle(id int) (*Article, error) {
	var article Article
	err := a.db.Where("id = ? AND deleted_on = ? ", id, 0).First(&article).Related(&article.Tag).Error()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &article, nil
}

func (a *Article) EditArticle(id int, data interface{}) error {
	if err := a.db.Model(&Article{}).Where("id = ? AND deleted_on = ? ", id, 0).Updates(data).Error(); err != nil {
		return err
	}

	return nil
}

func (a *Article) AddArticle(data map[string]interface{}) error {
	article := Article{
		TagID:         data["tag_id"].(int),
		Title:         data["title"].(string),
		Desc:          data["desc"].(string),
		Content:       data["content"].(string),
		CreatedBy:     data["created_by"].(string),
		State:         data["state"].(int),
		CoverImageUrl: data["cover_image_url"].(string),
	}
	if err := a.db.Create(&article).Error(); err != nil {
		return err
	}

	return nil
}

func (a *Article) DeleteArticle(id int) error {
	if err := a.db.Where("id = ?", id).Delete(Article{}).Error(); err != nil {
		return err
	}

	return nil
}

func (a *Article) CleanAllArticle() error {
	if err := a.db.Unscoped().Where("deleted_on != ? ", 0).Delete(&Article{}).Error(); err != nil {
		return err
	}

	return nil
}
