package v1

import (
	"github.com/EDDYCJY/go-gin-example/models"
	"github.com/EDDYCJY/go-gin-example/pkg/app"
	"github.com/EDDYCJY/go-gin-example/pkg/e"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetLogs(c *gin.Context) {
	appG := app.NewGin(c)
	logs, err := models.GetLogs()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_LOGS_FAIL, nil)
	}
	data := make(map[string]interface{})
	data["logs"] = logs
	appG.Response(http.StatusOK, e.SUCCESS, data)
}
