package function

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"handler/function/config"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/appleboy/go-fcm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/cast"
)

const (
	appId = "P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS"
)

func Handle(req []byte) string {
	var response Response
	var request Request
	const urlConst = "https://api.admin.u-code.io"

	err := json.Unmarshal(req, &request)
	if err != nil {
		return Handler("error 1", err.Error())
	}

	requestData := request.Data["object_data"].(map[string]interface{})
	naznacheniyaData, err, _ := GetSingleObject(urlConst, "naznachenie", cast.ToString(requestData["guid"]))
	if err != nil {
		return Handler("error 3", err.Error())
	}
}
