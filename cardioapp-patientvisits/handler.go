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

patientData, err, response := GetSingleObject(urlConst, "cleints", cast.ToString(naznacheniyaData.Data.Data.Response["cleints_id"]))
if err != nil {
	responseByte, _ := json.Marshal(response)
	return string(responseByte)
}

fullName := patientData.Data.Data.Response["cleint_lastname"].(string) + cast.ToString(patientData.Data.Data.Response["client_name"])
createObjectRequest := Request{
	Data: map[string]interface{}{
		"date":           cast.ToString(naznacheniyaData.Data.Data.Response["created_time"]),
		"doctor_id":      cast.ToString(naznacheniyaData.Data.Data.Response["doctor_id"]),
		"client_id":      cast.ToString(naznacheniyaData.Data.Data.Response["cleints_id"]),
		"naznachenie_id": cast.ToString(naznacheniyaData.Data.Data.Response["guid"]),
		"clinic_name":    doctorData.Data.Data.Response["hospital"],
		"id_doctor":      doctorData.Data.Data.Response["doctor_id"],
		"doctor_name":    doctorData.Data.Data.Response["doctor_name"],
		"doctor_phone":   doctorData.Data.Data.Response["phone_number"],
		"id_patient":     patientData.Data.Data.Response["user_number_id"],
		"patient_name":   fullName,
		"patient_phone":  patientData.Data.Data.Response["phone_number"],
	},
}
_, err, response = CreateObject(urlConst, "report_for_admin", createObjectRequest)
if err != nil {
	responseByte, _ := json.Marshal(response)
	return string(responseByte)
}

var (
	title   string = "У вас новое назначение от врача!"
	body    string = "Вам назначены препараты для лечения. Пожалуйста, ознакомьтесь с расписанием приема препаратов."
	titleUz string = "Sizda shifokor tomonidan yangi tayinlovlar bor!"
	bodyUz  string = "Sizga davolanish uchun dorilar buyurilgan. Iltimos, dori-darmonlarni qabul qilish jadvalini tekshiring."
)
notifRequest := Request{
	Data: map[string]interface{}{
		"client_id":    cast.ToString(naznacheniyaData.Data.Data.Response["cleints_id"]),
		"title":        title,
		"body":         body,
		"title_uz":     titleUz,
		"body_uz":      bodyUz,
		"preparati_id": "",
		"is_read":      false,
	},
}
_, err, response = CreateObject(urlConst, "notifications", notifRequest)
if err != nil {
	responseByte, _ := json.Marshal(response)
	return string(responseByte)
}

userNotif := UserNotification{
	Title:        title,
	Body:         body,
	TitleUz:      titleUz,
	BodyUz:       bodyUz,
	Fcm:          cast.ToString(patientData.Data.Data.Response["fcm_token"]),
	Platform:     cast.ToFloat64(patientData.Data.Data.Response["platform"]),
	UserLanguage: cast.ToString(patientData.Data.Data.Response["user_lang"]),
}

SendNotification(userNotif)
