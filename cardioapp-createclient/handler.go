
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
                    "name":           objectData["cleint_name"],
                    "phone":          objectData["phone_number"],
                    "project_id":     "a4dc1f1c-d20f-4c1a-abf5-b819076604bc",
                    "resource_type":  0,
                    "type":           "phone",
                },
            }
            _, err := DoRequest1(createUserUrl, "POST", createUser, appId)
            if err != nil {
                Handler("error", "error 4")
                return err
            }

            // Send2("START 11")
        }

        _, err = DoRequest("https://api.admin.u-code.io/v1/object/cleints", "PUT", userUpdateBody)
        if err != nil {
            Handler("error", "error 7")
            return err
        }

        _, err = DoRequest("https://api.admin.u-code.io/v1/object/user", "PUT", userUpdateBody)
        if err != nil {
            Handler("error", "error 7")
            return err
        }

        // Send2(fmt.Sprintf("userUpdateBody: %v", userUpdateBody) + "7")
    }

    return nil
}



func Send(text string) {
    bot, _ := tgbotapi.NewBotAPI("6041044802:AAEDdr0uD4SkxnnGctOOsA2Ua3Ovy-7Sy0A")

    msg := tgbotapi.NewMessage(266798451, text)

    bot.Send(msg)
}

func Send2(text string) {
    bot, _ := tgbotapi.NewBotAPI("6443522083:AAHGM7zwf93W1f2Z_C1Mj8sxRRARnBtROvs")

    msg := tgbotapi.NewMessage(1546926238, text)

    bot.Send(msg)
}

// ! MAKE MESSAGE FOR SENDING
func Handler(status, message string) string {

    var (
        response Response
        Message  = make(map[string]interface{})
    )

    // sendMessage("cardio user-to-user", status, message)
    response.Status = status
    Message["message"] = message
    respByte, _ := json.Marshal(response)
    return string(respByte)
}

func DoRequest1(url string, method string, body interface{}, appId string) ([]byte, error) {
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
    request.Header.Add("Authorization", "API-KEY")
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

func DoRequest(url string, method string, body interface{}) ([]byte, error) {
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

    if method == "PUT" || method == "DELETE" {
        request.Header.Add("authorization", "API-KEY")
        request.Header.Add("X-API-KEY", apiKey)
    }
    if method == "POST" {
        request.Header.Add("Resource-Id", "a97e8954-5d8e-4469-a241-9a9af2ea2978")
        request.Header.Add("Environment-Id", "dcd76a3d-c71b-4998-9e5c-ab1e783264d0")
    }
    resp, err := client.Do(request)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    respByte, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return respByte, nil
}
func generateSevenDigitNumber() int {
    rand.Seed(time.Now().UnixNano())
    num := rand.Intn(9000000) + 1000000
    return num
}

func GetListObject(url, method, appId string, body interface{}) (GetListClientApiResponse, error, Response) {
    response := Response{}
    getListResponseInByte, err := DoRequest1(url, method, body, appId)
    if err != nil {
        response.Data = map[string]interface{}{"message": "Error while getting single object"}
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

type ResponseUserModel struct {
    Status      string `json:"status"`
    Description string `json:"description"`
    Data        Data   `json:"data"`
}
type Data struct {
    ID        string `json:"id"`
    Login     string `json:"login"`
    Password  string `json:"password"`
    Phone     string `json:"phone"`
    CompanyID string `json:"company_id"`
}

// This is response struct from create
type Datas struct {
    Data struct {
        Data struct {
            Data map[string]interface{} `json:"data"`
        } `json:"data"`
    } `json:"data"`
}

// This is get single api response
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

type NewRequestBody struct {
    Data map[string]interface{} `json:"data"`
}
type Request struct {
    Data map[string]interface{} `json:"data"`
}

// This is get list api response
type GetListClientApiResponse struct {
    Data GetListClientApiData `json:"data"`
}

type GetListClientApiData struct {
    Data GetListClientApiResp `json:"data"`
}

type GetListClientApiResp struct {
    Response []map[string]interface{} `json:"response"`
}

