package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cast"
	// "github.com/spf13/cast"
)

const (
	apiKey = "P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS"

	createUrl = "https://api.auth.u-code.io/v2/user"
	deleteUrl = "https://api.auth.u-code.io/v2/user"
	checkUrl  = "https://api.auth.u-code.io/v2/user/check"
)

// Handle a serverless request
func Handle(req []byte) string {
	// Send2("START 0")
	var (
		request NewRequestBody
	)

	if err := json.Unmarshal(req, &request); err != nil {
		return Handler("error", "error 1")
	}

	if cast.ToString(request.Data["method"]) == "CREATE" {
		if err := CreateUser(request); err != nil {
			return Handler("error", "eror from CREATE")
		}
	}

	// Send2("START 2")

	return Handler("OK", "OK")
}

func CreateUser(request NewRequestBody) error {
	// Send2("START 1")

	// ! firsst we must check user: is it existed in auth of the ucode

	var (
		appId             = cast.ToString(request.Data["app_id"])
		objectData        = cast.ToStringMap(request.Data["object_data"])
		userGuid          = cast.ToString(objectData["guid"])
		responseCheckUser ResponseUserModel
		checkRequestBody  = map[string]interface{}{
			"email":         "",
			"login":         "",
			"phone":         objectData["phone_number"],
			"resource_type": 1,
		}
		phoneNumber = cast.ToString(objectData["phone_number"])
	)
	responseBodyCheck, err := DoRequest(checkUrl, "POST", checkRequestBody)
	if err != nil {
		Handler("error", "error 2")
		return err
	}
	// Send2(fmt.Sprintf("responseBodyCheck: %v", string(responseBodyCheck)+"2"))

	if err := json.Unmarshal(responseBodyCheck, &responseCheckUser); err != nil {
		// Send2(fmt.Sprintf("error: %v", err) + "3")
		// Handler("error", " error 3"+string(responseBodyCheck))
		// return err
		var (
			responseUser ResponseUserModel

			requestBody = map[string]interface{}{
				"active":                  1,
				"client_type_id":          "24fd6d7e-c0e7-4029-88cc-2595e9c643d5",
				"role_id":                 "425486de-89dc-48a7-9fa8-47f7b4eeffcb",
				"login":                   "",
				"name":                    objectData["cleint_name"],
				"password":                "",
				"phone":                   objectData["phone_number"],
				"project_id":              "a4dc1f1c-d20f-4c1a-abf5-b819076604bc",
				"resource_type":           0,
				"year_of_birth":           "",
				"base_url":                "",
				"client_platform_id":      "",
				"company_id":              "",
				"email":                   "",
				"expires_at":              "",
				"photo_url":               "",
				"resource_environment_id": "",
				// "balance_id":              generateSevenDigitNumber(),
			}
		)

		responseBody, err := DoRequest(createUrl, "POST", requestBody)
		if err != nil {
			Handler("error", "error 4")
			return err
		}
		// Send2(fmt.Sprintf("responseBody: %v", string(responseBody)) + "4")

		if err := json.Unmarshal(responseBody, &responseUser); err != nil {
			Handler("error", " error 5"+string(responseBody))
			return err
		}

		// Send2(fmt.Sprintf("responseUser: %v", responseUser) + "5")

		userDeleteUrl := fmt.Sprintf("https://api.admin.u-code.io/v1/object/cleints/%s", responseUser.Data.ID)
		_, err = DoRequest(userDeleteUrl, "DELETE", Request{Data: map[string]interface{}{}})
		if err != nil {
			Handler("error", "error 6")
			return err
		}

		// Send2(fmt.Sprintf("userDeleteUrl: %v", userDeleteUrl) + "6")

		var (
			userUpdateBody = Request{
				Data: map[string]interface{}{
					"guid":      userGuid,
					"auth_guid": responseUser.Data.ID,
				},
			}
		)

		_, err = DoRequest("https://api.admin.u-code.io/v1/object/cleints", "PUT", userUpdateBody)
		if err != nil {
			Handler("error", "error 7")
			return err
		}

		// Send2(fmt.Sprintf("userUpdateBody: %v", userUpdateBody) + "7")
	} else {
		// Send2(fmt.Sprintf("responseCheckUser: %v", responseCheckUser) + "8")
		m := map[string]interface{}{}
		err := json.Unmarshal(responseBodyCheck, &m)
		if err != nil {
			return err
		}

		id := m["data"].(map[string]interface{})["user_id"].(string)
		var (
			userUpdateBody = Request{
				Data: map[string]interface{}{
					"guid":      userGuid,
					"auth_guid": id,
					"phone":     phoneNumber,
				},
			}
		)

		// update user, if admin contains user
		phoneNumber = strings.Replace(phoneNumber, "+", "", 1)
		b := map[string]interface{}{
			"data": map[string]interface{}{
				"phone": phoneNumber,
			},
		}

		getUserByPhone := fmt.Sprintf("%v/v1/object/get-list/user", "https://api.admin.u-code.io")
		res, err, _ := GetListObject(getUserByPhone, "POST", appId, b)
		if err != nil {
			return err
		}
		// Send2(fmt.Sprintf("res: %v", res) + "9")
		//fmt.Println(res.Data.Data.Response)

		if len(res.Data.Data.Response) == 0 {
			// Send2("START 10")

			createUserUrl := "https://api.auth.u-code.io/v2/register?project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc"
			createUser := map[string]interface{}{
				"data": map[string]interface{}{
					"active":         1,
					"client_type_id": "24fd6d7e-c0e7-4029-88cc-2595e9c643d5",
					"role_id":        "425486de-89dc-48a7-9fa8-47f7b4eeffcb",
					"login":          "",
					"name":           objectData["cleint_name"],
					"phone":          objectData["phone_number"],
					"project_id":     "a4dc1f1c-d20f-4c1a-abf5-b819076604bc",
					"resource_type":  0,
					"year_of_birth":  "",
					"base_url":       "",
					"client_platform_id": "",
				},
			}

			response, err := DoRequest(createUserUrl, "POST", createUser)
			if err != nil {
				Handler("error", "error 11")
				return err
			}
		}
	}

	return nil
}

