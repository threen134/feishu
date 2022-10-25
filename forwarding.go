package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type SysdigWebhookBody struct {
	Alert      Alert      `json:"alert"`
	Event      Event      `json:"event"`
	Condition  string     `json:"condition"`
	Source     string     `json:"source"`
	State      string     `json:"state"`
	Timestamp  int64      `json:"timestamp"`
	Timespan   int        `json:"timespan"`
	Entities   []Entities `json:"entities"`
	CustomData CustomData `json:"customData"`
}
type Alert struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Scope         string `json:"scope"`
	Severity      int    `json:"severity"`
	SeverityLabel string `json:"severityLabel"`
	EditURL       string `json:"editUrl"`
	Subject       string `json:"subject"`
	Body          string `json:"body"`
}
type Event struct {
	ID       int    `json:"id"`
	URL      string `json:"url"`
	Username string `json:"username"`
}
type MetricValues struct {
	Metric           string  `json:"metric"`
	Aggregation      string  `json:"aggregation"`
	GroupAggregation string  `json:"groupAggregation"`
	Value            float64 `json:"value"`
}
type Entities struct {
	Entity       string         `json:"entity"`
	MetricValues []MetricValues `json:"metricValues"`
}
type CustomData struct {
	Webhook string `json:"webhook"`
}

type FeishuBody struct {
	ReceiveID string `json:"receive_id"`
	MsgType   string `json:"msg_type"`
	Card      Card   `json:"card"`
}
type I18N struct {
	EnUs string `json:"en_us"`
}
type Title struct {
	Tag  string `json:"tag"`
	I18N I18N   `json:"i18n"`
}
type Header struct {
	Title Title `json:"title"`
}
type EnUs struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}
type I18NElements struct {
	EnUs []EnUs `json:"en_us"`
}
type Card struct {
	Header       Header       `json:"header"`
	I18NElements I18NElements `json:"i18n_elements"`
}

var feishuTemplate = `
	{
		"receive_id": "all",
		"msg_type": "interactive",
		"card":{
			"header": {
				"title": {
				"tag": "lark_md",
				"i18n": {
					"en_us": "cloud moniter alert"
				}
				}
			},
			"i18n_elements": {
				"en_us": [
					{
					"tag": "markdown",
					"content": ""
					}
					]
				}
			}
	}
`

var key = GetEnvDefault("BASE_KEY", "")

func forwarding(ctx *gin.Context) {
	if key != "" && ctx.GetHeader("BASE_KEY") != key {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"result": "failed",
			"code":   "Unauthorized",
		})
		return
	}

	requestBody := &SysdigWebhookBody{}
	if err := ctx.BindJSON(&requestBody); err != nil {
		log.WithError(err).Error("fail to read json body")
		ctx.AbortWithError(400, err)
	}

	if requestBody.CustomData.Webhook == "" {
		err := errors.New("can not find webhook endpoint")
		log.Error(err)
		ctx.AbortWithError(400, err)
	}
	restBody := convert(requestBody)

	client := resty.New()

	resp, err := client.R().
		SetBody(restBody).
		Post(requestBody.CustomData.Webhook)

	if err != nil {
		ctx.AbortWithError(500, err)
	}
	if resp.StatusCode() != http.StatusOK {
		log.Info(string(resp.Body()))
		ctx.AbortWithStatus(resp.StatusCode())
	}
	log.Info(string(resp.Body()))
	ctx.JSON(http.StatusOK, gin.H{"result": "success"})
}

// 获取环境变量信息
func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

func convert(src *SysdigWebhookBody) *FeishuBody {
	feishu := &FeishuBody{}
	json.Unmarshal([]byte(feishuTemplate), feishu)
	feishu.Card.I18NElements.EnUs[0].Content = src.Alert.Body
	return feishu
}
