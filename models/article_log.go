package models

const (
	OperationAdd    = "Add"
	OperationEdit   = "Edit"
	OperationDelete = "Delete"
)

type ArticleLog struct {
	ID           int    `gorm:"primary_key" json:"id"`
	CreatedOn    int    `json:"created_on"`
	CreatedBy    string `json:"created_by"`
	ArticleTitle string `json:"article_title"`
	Operation    string `json:"operation"`
}

func AddArticleLog(log ArticleLog) error {
	return Db.Create(&log).Error()
}

func GetLogs() ([]*ArticleLog, error) {
	var logs []*ArticleLog
	err := Db.Find(&logs).Error()
	return logs, err
}
