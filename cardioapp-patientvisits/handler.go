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

// func main() {
// 	str := `{
// 		"data": {
// 			"additional_parameters": [],
// 			"app_id": "P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS",
// 			"environment_id": "dcd76a3d-c71b-4998-9e5c-ab1e783264d0",
// 			"method": "CREATE",
// 			"object_data": {
// 				"amount": "0",
// 				"amount_med_taken": "0",
// 				"cleints_id": "e94effc5-0c07-4f1b-a65a-ecbc29fe9205",
// 				"client_name": "Baxrom",
// 				"client_surname": "Umarov",
// 				"company_service_environment_id": "dcd76a3d-c71b-4998-9e5c-ab1e783264d0",
// 				"company_service_project_id": "a4dc1f1c-d20f-4c1a-abf5-b819076604bc",
// 				"consultation_type": "платная консультация",
// 				"created_time": "2023-12-26T19:50:34.569Z",
// 				"default_number": 1,
// 				"doctor_id": "72cc2ea7-88eb-4dfc-838d-b99d6e354be6",
// 				"guid": "c262448b-9c99-46be-8e6c-1063acfbd83f",
// 				"ill_id": "55220df5-e863-4a9d-9abf-c277ee881407",
// 				"ill_name": "A01.2 Паратиф B",
// 				"increment_id": "N-000740",
// 				"invite": false,
// 				"multi": []
// 			},
// 			"object_data_before_update": null,
// 			"object_ids": [
// 				"c262448b-9c99-46be-8e6c-1063acfbd83f"
// 			],
// 			"project_id": "a4dc1f1c-d20f-4c1a-abf5-b819076604bc",
// 			"table_slug": "naznachenie",
// 			"user_id": "72cc2ea7-88eb-4dfc-838d-b99d6e354be6"
// 		}
// 	}`
// 	fmt.Println(Handle([]byte(str)))
// }

// Handle a serverless request
func Handle(req []byte) string {
	var response Response
	var request Request
	const urlConst = "https://api.admin.u-code.io"

	err := json.Unmarshal(req, &request)
	if err != nil {
		return Handler("error 1", err.Error())
	}
	// Send(string(req))
	// var tableSlug = "naznachenie"
	requestData := request.Data["object_data"].(map[string]interface{})
	naznacheniyaData, err, _ := GetSingleObject(urlConst, "naznachenie", cast.ToString(requestData["guid"]))
	if err != nil {
		// fmt.Println(err)
		return Handler("error 3", err.Error())

	}
	// os.Exit(1)
	// create patient visits
	if cast.ToString(request.Data["method"]) == "CREATE" {
		// ----------------------------------------------------------------------------------------------------------------------------------------------

		doctorData, err, _ := GetSingleObject(urlConst, "doctor", cast.ToString(naznacheniyaData.Data.Data.Response["doctor_id"]))
		if err != nil {
			return Handler("error 4", err.Error())

		}

		// ----------------------------------------------------------------------------------------------------------------------------------------------

		getListObjectRequest := Request{
			// some filters
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

			naznacheniyaIDs := []string{}
			naznacheniyaIDs = append(naznacheniyaIDs, id)

			//create objects response example
			createtObjectRequest := Request{
				// some filters
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
				// fmt.Println("error6")

				responseByte, _ := json.Marshal(response)
				return string(responseByte)
			}
		} else {
			// tableSlug = "patient_visits"
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
				// fmt.Println("error7")

				responseByte, _ := json.Marshal(response)
				return string(responseByte)
			}
		}
		// ----------------------------------------------------------------------------------------------------------------------------------------------
		// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// create report for admin
		// get patients
		// tableSlug = "cleints"
		patientData, err, response := GetSingleObject(urlConst, "cleints", cast.ToString(naznacheniyaData.Data.Data.Response["cleints_id"]))
		if err != nil {
			// fmt.Println("error8")

			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
		fullName := patientData.Data.Data.Response["cleint_lastname"].(string) + cast.ToString(patientData.Data.Data.Response["client_name"])
		// tableSlug = "report_for_admin"
		//create objects response example
		createObjectRequest := Request{
			// some filters
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
			// fmt.Println("error9")

			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}

		//Send notification
		var (
			title   string = "У вас новое назначение от врача!"
			body    string = "Вам назначены препараты для лечения. Пожалуйста, ознакомьтесь с расписанием приема препаратов."
			titleUz string = "Sizda shifokor tomonidan yangi tayinlovlar bor!"
			bodyUz  string = "Sizga davolanish uchun dorilar buyurilgan. Iltimos, dori-darmonlarni qabul qilish jadvalini tekshiring."
		)
		// tableSlug = "notifications"
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
			// fmt.Println("error10")

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

		// ----------------------------------------------------------------------------------------------------------------------------------------------
		// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// create report for doctor
		// tableSlug = "puls"
		var (
			puls     float64
			pressure string
		)

		today := time.Now()

		// Get the beginning of the day (00:00:00)
		firstDateTime := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
		firstDateTime = firstDateTime.Add(time.Hour)
		// Get the end of the day (23:59:59)
		lastDateTime := today.Add(time.Hour)

		// Format the first and last date time as strings
		firstDateTimeStr := firstDateTime.Format("2006-01-02T15:04:05Z")
		lastDateTimeStr := lastDateTime.Format("2006-01-02T15:04:05Z")

		// get pressure and puls
		newReq := Request{
			Data: map[string]interface{}{
				"cleints_id": cast.ToString(naznacheniyaData.Data.Data.Response["cleints_id"]),
				"order": map[string]interface{}{
					"createdAt": -1,
				},
				"date": map[string]interface{}{
					"$lte": lastDateTimeStr,
					"$gte": firstDateTimeStr,
				},
			},
		}
		pulsData, err, response := GetListObject(urlConst, "puls", newReq)
		if err != nil {
			// fmt.Println("error11")

			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
		if len(pulsData.Data.Data.Response) > 0 {
			puls = pulsData.Data.Data.Response[0]["puls"].(float64)
			sistolicheskoe := strconv.FormatFloat(cast.ToFloat64(pulsData.Data.Data.Response[0]["sistolicheskoe"]), 'f', 0, 64)
			diastolicheskoe := strconv.FormatFloat(cast.ToFloat64(pulsData.Data.Data.Response[0]["diastolicheskoe"]), 'f', 0, 64)
			pressure = diastolicheskoe + "/" + sistolicheskoe
		}

		// tableSlug = "report_for_doctor"
		//create objects response example
		createObjectRequest = Request{
			// some filters
			Data: map[string]interface{}{
				"date":           cast.ToString(naznacheniyaData.Data.Data.Response["created_time"]),
				"client_id":      cast.ToString(naznacheniyaData.Data.Data.Response["cleints_id"]),
				"naznachenie_id": cast.ToString(naznacheniyaData.Data.Data.Response["guid"]),
				"doctor_id":      cast.ToString(naznacheniyaData.Data.Data.Response["doctor_id"]),
				"id_patient":     patientData.Data.Data.Response["user_number_id"],
				"patient_phone":  patientData.Data.Data.Response["phone_number"],
				"patient_illnes": naznacheniyaData.Data.Data.Response["ill_name"],
				"id_naznachenie": naznacheniyaData.Data.Data.Response["increment_id"],
				"patient_name":   fullName,
				"puls":           puls,
				"blood_pressure": pressure,
			},
		}
		_, err, response = CreateObject(urlConst, "report_for_doctor", createObjectRequest)
		if err != nil {
			// fmt.Println("error12")

			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
		// ----------------------------------------------------------------------------------------------------------------------------------------------

	}

	response.Data = map[string]interface{}{}
	response.Status = "done" //if all will be ok else "error"
	responseByte, _ := json.Marshal(response)

	return string(responseByte)
}

func GetListObject(url, tableSlug string, request Request) (GetListClientApiResponse, error, Response) {
	response := Response{}

	getListResponseInByte, err := DoRequest(url+"/v1/object/get-list/"+tableSlug+"?from-ofs=true&project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "POST", request, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while getting list of objects"}
		response.Status = "error"
		return GetListClientApiResponse{}, errors.New("error"), response
	}
	var getListObject GetListClientApiResponse
	err = json.Unmarshal(getListResponseInByte, &getListObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list object"}
		response.Status = "error"
		return GetListClientApiResponse{}, errors.New("error"), response
	}
	return getListObject, nil, response
}

func SendNotification(notification UserNotification) {
	// // 0 - IOS, 1 - Android
	// msg := &fcm.Message{
	// 	To: notification.Fcm,
	// }
	// if int(notification.Platform) == 1 {
	// 	msg.Data = map[string]interface{}{
	// 		"title": notification.Title,
	// 		"body":  notification.Body,
	// 	}
	// } else if int(notification.Platform) == 0 {
	// 	msg.Notification = &fcm.Notification{
	// 		Title: notification.Title,
	// 		Body:  notification.Body,
	// 	}
	// }
	// // Create a FCM client to send the message.
	// client, _ := fcm.NewClient("AAAAyfojPTI:APA91bGSMDl45M7GWh08DIGLzZR_hI2mHXqcvI84_p0_3LXSeISJJgx7d41YUMc7riyk66IJYoblRz9hGDoq8iWUxQZZIO6sRiUhiBnaqxgoi575zb2fcQgjCh4W7xAEjJjP3UXVod6h")
	// client.Send(msg)

	var (
		title, body string
	)

	if notification.UserLanguage == "ru" {
		title = notification.Title
		body = notification.Body
	} else {
		title = notification.TitleUz
		body = notification.BodyUz
	}

	// serverKey := "AAAAyfojPTI:APA91bGSMDl45M7GWh08DIGLzZR_hI2mHXqcvI84_p0_3LXSeISJJgx7d41YUMc7riyk66IJYoblRz9hGDoq8iWUxQZZIO6sRiUhiBnaqxgoi575zb2fcQgjCh4W7xAEjJjP3UXVod6h"
	// var payload string

	client, err := fcm.NewClient(
		context.Background(),
		fcm.WithCredentialsJSON([]byte(config.FcmJson)),
	)
	if err != nil {
		return
	}

	// if int(notification.Platform) == 0 {
	// 	notificationMsg = messaging.Message{
	// 		Token: notification.Fcm,
	// 		Notification: &messaging.Notification{
	// 			Title: title,
	// 			Body:  body,
	// 		},
	// 		Android: &messaging.AndroidConfig{
	// 			Priority: "high",
	// 		},
	// 		APNS: &messaging.APNSConfig{
	// 			Payload: &messaging.APNSPayload{
	// 				Aps: &messaging.Aps{
	// 					ContentAvailable: true,
	// 				},
	// 			},
	// 			Headers: map[string]string{
	// 				"apns-priority": "10",
	// 			},
	// 		},
	// 	}
	// 	// payload = `
	// 	// 	{
	// 	// 		"to": "%v",
	// 	// 		"notification": {
	// 	// 			"title": "%v",
	// 	// 			"body": "%v"
	// 	// 		}
	// 	// 	}
	// 	// `
	// 	// payload = fmt.Sprintf(payload, notification.Fcm, title, body)
	// } else {

	// 	// payload = `
	// 	// 	{
	// 	// 		"to": "%v",
	// 	// 		"data": {
	// 	// 			"title": "%v",
	// 	// 			"body": "%v"
	// 	// 		}
	// 	// 	}
	// 	// `
	// 	// payload = fmt.Sprintf(payload, notification.Fcm, title, body)
	// }

	_, err = client.Send(context.Background(), &messaging.Message{
		Token: notification.Fcm,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: map[string]string{
			"title": title,
			"body":  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
			Headers: map[string]string{
				"apns-priority": "10",
			},
		},
	})
	if err != nil {
		return
	}

	// req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer([]byte(payload)))
	// if err != nil {
	// 	// Send("This is error" + err.Error())
	// 	return
	// }

	// req.Header.Set("Authorization", "key="+serverKey)
	// req.Header.Set("Content-Type", "application/json")

	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	return
	// }
}

func Send(text string) {
	bot, err := tgbotapi.NewBotAPI("6364121049:AAGqVbhQCxcsHckUzEB28FQpOq289mCsCKo")
	if err != nil {
		log.Panic(err)
	}

	chatID := int64(-4047694896)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("message from madad cardioapp patientvisits function: %s", text))
	_, err = bot.Send(msg)
	if err != nil {
		log.Panic(err)
	}
}

func GetSingleObject(url, tableSlug, guid string) (ClientApiResponse, error, Response) {
	response := Response{}

	var getSingleObject ClientApiResponse
	getSingleResponseInByte, err := DoRequest(url+"/v1/object/"+tableSlug+"/"+guid+"?from-ofs=true&project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "GET", nil, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while getting single object"}
		response.Status = "error"
		return ClientApiResponse{}, errors.New("error"), response
	}
	err = json.Unmarshal(getSingleResponseInByte, &getSingleObject)
	if err != nil {

		response.Data = map[string]interface{}{"message": "Error while unmarshalling single object" + err.Error()}
		response.Status = "error"
		return ClientApiResponse{}, errors.New("error"), response
	}
	return getSingleObject, nil, response
}

func CreateObject(url, tableSlug string, request Request) (Datas, error, Response) {
	response := Response{}

	var createdObject Datas
	createObjectResponseInByte, err := DoRequest(url+"/v1/object/"+tableSlug+"?from-ofs=true&project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "POST", request, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while creating object"}
		response.Status = "error"
		return Datas{}, errors.New("error"), response
	}
	err = json.Unmarshal(createObjectResponseInByte, &createdObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling create object object"}
		response.Status = "error"
		return Datas{}, errors.New("error"), response
	}
	return createdObject, nil, response
}

func UpdateObject(url, tableSlug string, request Request) (error, Response) {
	response := Response{}

	_, err := DoRequest(url+"/v1/object/"+tableSlug+"?from-ofs=true&project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "PUT", request, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while updating object"}
		response.Status = "error"
		return errors.New("error"), response
	}
	return nil, response
}

func DeleteObject(url, tableSlug, guid string) (error, Response) {
	response := Response{}

	_, err := DoRequest(url+"/v1/object/{table_slug}/{guid}?from-ofs=true", "DELETE", Request{}, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while updating object"}
		response.Status = "error"
		return errors.New("error"), response
	}
	return nil, response
}

// func main() {
// 	data := `{
// 		"data": {
// 			"additional_parameters": [],
// 			"app_id": "P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS",
// 			"method": "CREATE",
// 			"object_data": {
// 				"amount": "0",
// 				"amount_med_taken": "0",
// 				"cleints_id": "65f2b5dd-2cbf-4e42-8d7d-2d186f6b2c25",
// 				"client_name": "Mirabbos",
// 				"client_surname": "Botirjonov",
// 				"comment": "\u003cp\u003edasdas\u003c/p\u003e",
// 				"company_service_environment_id": "dcd76a3d-c71b-4998-9e5c-ab1e783264d0",
// 				"company_service_project_id": "a4dc1f1c-d20f-4c1a-abf5-b819076604bc",
// 				"created_time": "2023-10-02T04:53:16.911Z",
// 				"default_number": 1,
// 				"doctor_id": "61ba244d-ee0d-4581-9f54-75d180bbdb05",
// 				"guid": "c8b346da-a27b-4852-be4f-70ce99d391af",
// 				"ill_id": "1e1cac8a-8ae6-494d-a8f3-c8a9cc9e5119",
// 				"ill_name": "A02.0 Сальмонеллезный энтерит",
// 				"increment_id": "N-000477",
// 				"invite": false,
// 				"multi": []
// 			},
// 			"object_data_before_update": null,
// 			"object_ids": [
// 				"c8b346da-a27b-4852-be4f-70ce99d391af"
// 			],
// 			"table_slug": "naznachenie",
// 			"user_id": "0de8f626-388c-4ea8-8213-aa54c1ad4a5d"
// 		}
// 	}`
// 	fmt.Println(Handle([]byte(data)))
// }

// Datas This is response struct from create
type Datas struct {
	Data struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	} `json:"data"`
}

// ClientApiResponse This is get single api response
type ClientApiResponse struct {
	Data ClientApiData `json:"data"`
}

type ClientApiData struct {
	Data ClientApiResp `json:"data"`
}

type ClientApiResp struct {
	Response map[string]interface{} `json:"response"`
}

type Response struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// NewRequestBody's Data (map) field will be in this structure
//.   fields
// objects_ids []string
// table_slug string
// object_data map[string]interface
// method string
// app_id string

// but all field will be an interface, you must do type assertion

type HttpRequest struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
	Params  url.Values  `json:"params"`
	Body    []byte      `json:"body"`
}

type AuthData struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type NewRequestBody struct {
	RequestData HttpRequest            `json:"request_data"`
	Auth        AuthData               `json:"auth"`
	Data        map[string]interface{} `json:"data"`
}
type Request struct {
	Data map[string]interface{} `json:"data"`
}

// GetListClientApiResponse This is get list api response
type GetListClientApiResponse struct {
	Data GetListClientApiData `json:"data"`
}

type GetListClientApiData struct {
	Data GetListClientApiResp `json:"data"`
}

type GetListClientApiResp struct {
	Response []map[string]interface{} `json:"response"`
}

type UserNotification struct {
	Title        string
	Fcm          string
	Body         string
	Platform     float64
	TitleUz      string
	BodyUz       string
	UserLanguage string
}

func DoRequest(url string, method string, body interface{}, appId string) ([]byte, error) {
	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("authorization", "API-KEY")
	request.Header.Add("X-API-KEY", appId)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respByte, nil
}

func Handler(status, message string) string {
	var (
		response Response
		Message  = make(map[string]interface{})
	)

	// Send(status + message)
	response.Status = status
	Message["message"] = message
	response.Data = Message
	respByte, _ := json.Marshal(response)
	return string(respByte)

}
