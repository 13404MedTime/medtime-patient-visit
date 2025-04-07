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

if cast.ToString(request.Data["method"]) == "CREATE" {
	doctorData, err, _ := GetSingleObject(urlConst, "doctor", cast.ToString(naznacheniyaData.Data.Data.Response["doctor_id"]))
	if err != nil {
		return Handler("error 4", err.Error())
	}

	getListObjectRequest := Request{
		Data: map[string]interface{}{
			"doctor_id":  cast.ToString(naznacheniyaData.Data.Data.Response["doctor_id"]),
			"cleints_id": cast.ToString(naznacheniyaData.Data.Data.Response["cleints_id"]),
		},
	}
	patientVisits, err, _ := GetListObject(urlConst, "patient_visits", getListObjectRequest)
	if err != nil {
		return Handler("error 5", err.Error())
	}

	if len(patientVisits.Data.Data.Response) < 1 {
		id := cast.ToString(requestData["guid"])
		naznacheniyaIDs := []string{id}
		createtObjectRequest := Request{
			Data: map[string]interface{}{
				"doctor_id":       cast.ToString(naznacheniyaData.Data.Data.Response["doctor_id"]),
				"cleints_id":      cast.ToString(naznacheniyaData.Data.Data.Response["cleints_id"]),
				"date":            cast.ToString(naznacheniyaData.Data.Data.Response["created_time"]),
				"doctor_name":     cast.ToString(doctorData.Data.Data.Response["doctor_name"]),
				"naznachenie_ids": naznacheniyaIDs,
			},
		}
		_, err, response = CreateObject(urlConst, "patient_visits", createtObjectRequest)
		if err != nil {
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
	} else {
		patientVisitsId := cast.ToString(patientVisits.Data.Data.Response[0]["guid"])
		naznacheniyaIDs := cast.ToSlice(patientVisits.Data.Data.Response[0]["naznachenie_ids"])
		naznacheniyaIDs = append(naznacheniyaIDs, cast.ToString(requestData["guid"]))

		updateRequest := Request{
			Data: map[string]interface{}{
				"guid":            patientVisitsId,
				"date":            cast.ToString(naznacheniyaData.Data.Data.Response["created_time"]),
				"naznachenie_ids": naznacheniyaIDs,
			},
		}
		err, response = UpdateObject(urlConst, "patient_visits", updateRequest)
		if err != nil {
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
	}
}
