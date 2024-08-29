package filmapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"geico.visualstudio.com/Billing/plutus/api"
	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
)

var log = logging.GetLogger("filmapi-client")

var filmApiUrl string
var filmApiAuthorization string
var filmApiTimeout time.Duration
var filmApiRetryCount int

func Init(url string, authorization string, timeout int, retryCount int, configuration any) {
	filmApiUrl = url
	filmApiAuthorization = authorization
	filmApiTimeout = time.Duration(timeout * int(time.Second))
	filmApiRetryCount = retryCount
}

func UploadFile(path string, content string) error {

	apiRequest := commonAppModels.APIRequest{
		Type:          enums.FilmAPIUploadFile,
		Url:           fmt.Sprintf("%sapi/Files/SaveFileForBilling", filmApiUrl),
		Request:       []byte(content),
		Authorization: fmt.Sprintf("Basic %s", filmApiAuthorization),
		Timeout:       filmApiTimeout,
		FilePath:      path,
	}

	_, err := api.RetryApiCall(api.PostMIMEApiCall, apiRequest, filmApiRetryCount)
	if err != nil {
		log.Error(context.Background(), err, "film api upload call is not succesful despite '%d' attempts", filmApiRetryCount)
		return err
	}

	return nil
}

func DeleteFile(path string) error {

	apiRequest := commonAppModels.APIRequest{
		Type:          enums.FilmAPIDeleteFile,
		Url:           fmt.Sprintf("%sapi/Files/Deletefile?FileName=%s", filmApiUrl, path),
		Request:       []byte{},
		Authorization: fmt.Sprintf("Basic %s", filmApiAuthorization),
		Timeout:       filmApiTimeout,
		FilePath:      path,
	}

	_, err := api.RetryApiCall(api.PostRestApiCall, apiRequest, filmApiRetryCount)
	if err != nil {
		log.Error(context.Background(), err, "film api delete call is not succesful despite '%d' attempts", filmApiRetryCount)
		return err
	}

	return nil
}

func GetFileLines(path string) ([]string, error) {

	apiRequest := commonAppModels.APIRequest{
		Type:          enums.FilmAPIGetFileLines,
		Url:           fmt.Sprintf("%sapi/Files/Retrieve?FileName=%s&key=%s", filmApiUrl, path, ""),
		Request:       []byte{},
		Authorization: fmt.Sprintf("Basic %s", filmApiAuthorization),
		Timeout:       filmApiTimeout,
		FilePath:      path,
	}

	res, err := api.RetryApiCall(api.GetRestApiCall, apiRequest, filmApiRetryCount)
	if err != nil {
		log.Error(context.Background(), err, "film api get call is not succesful despite '%d' attempts", filmApiRetryCount)
		return []string{}, err
	}
	responseStr := string(res.Response)

	fileLines := strings.Split(responseStr, "\n")

	return fileLines, nil
}

func GetFilesInFolder(folderPath string) ([]string, error) {

	var files []string
	apiRequest := commonAppModels.APIRequest{
		Type:          enums.FilmAPIGetFilesInFolder,
		Url:           fmt.Sprintf("%sapi/Files/FilesInFolder?folder=%s", filmApiUrl, folderPath),
		Request:       []byte{},
		Authorization: fmt.Sprintf("Basic %s", filmApiAuthorization),
		Timeout:       filmApiTimeout,
		FilePath:      folderPath,
	}

	apiResponse, err := api.RetryApiCall(api.GetRestApiCall, apiRequest, filmApiRetryCount)
	if err != nil {
		log.Error(context.Background(), err, "film api get call is not succesful despite '%d' attempts", filmApiRetryCount)
		return []string{}, err
	}

	jsonErr := json.Unmarshal(apiResponse.Response, &files)
	if jsonErr != nil {
		log.Error(context.Background(), err, "error in umarshalling film api response")
		return nil, err
	}

	return files, nil
}

func MoveFile(oldFileName string, newFileName string) error {

	fileNames := map[string]interface{}{
		"OldFileName": oldFileName,
		"NewFileName": newFileName,
	}

	jsonData, err := json.Marshal(fileNames)
	if err != nil {
		log.Error(context.Background(), err, "invalidJson")
	}

	apiRequest := commonAppModels.APIRequest{
		Type:          enums.FilmAPIGetFilesInFolder,
		Url:           fmt.Sprintf("%sapi/Files/Renamefile", filmApiUrl),
		Request:       jsonData,
		Authorization: fmt.Sprintf("Basic %s", filmApiAuthorization),
		Timeout:       filmApiTimeout,
	}

	_, err = api.RetryApiCall(api.PostRestApiCall, apiRequest, filmApiRetryCount)
	if err != nil {
		log.Error(context.Background(), err, "film api RenameFile call is not succesful despite '%d' attempts", filmApiRetryCount)
		return err
	}

	return nil
}
