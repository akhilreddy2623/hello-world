package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
)

var log = logging.GetLogger("pkg-api")

const statusError = "status '%d' return from api call"

func GenerateBearerTokenApiCall(request commonAppModels.BearerTokenRequest) (*commonAppModels.BearerTokenResponse, error) {

	data := url.Values{}
	data.Set("client_id", request.ClientId)
	data.Set("client_secret", request.ClientSecret)
	data.Set("scope", request.Scope)
	data.Set("grant_type", request.GrantType)

	res, err := http.Post(request.Url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		log.Error(context.Background(), err, "error in creating request for GenerateBearerTokenApiCall")
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(context.Background(), err, "error in reading api response from response body")
		return nil, err
	}

	response := commonAppModels.BearerTokenResponse{}
	if err := json.Unmarshal(resBody, &response); err != nil {
		log.Error(context.Background(), err, "unable to unmarshal BearerTokenResponse")
		return nil, err
	}

	defer res.Body.Close()
	return &response, nil
}

func PostRestApiCall(apiRequest commonAppModels.APIRequest) (*commonAppModels.APIResponse, error) {
	req, err := http.NewRequest("POST", apiRequest.Url, bytes.NewBuffer(apiRequest.Request))

	if err != nil {
		log.Error(context.Background(), err, "error in creating request for api type '%s'", apiRequest.Type)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if apiRequest.Type == enums.Forte {
		req.Header.Set("X-Forte-Auth-Organization-Id", apiRequest.Header)
	}
	req.Header.Set("Authorization", apiRequest.Authorization)

	client := http.Client{Timeout: apiRequest.Timeout}
	res, err := client.Do(req)
	if err != nil {
		log.Error(context.Background(), err, "error in making api call")
		return nil, err
	}
	// For forte decline we get 400 back
	if res.StatusCode != 201 && res.StatusCode != 200 && res.StatusCode != 400 {
		log.Error(context.Background(), err, statusError, res.StatusCode)
		return nil, fmt.Errorf(statusError, res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(context.Background(), err, "error in reading api response from response body")
		return nil, err
	}

	apiResponse := commonAppModels.APIResponse{
		Status:   res.Status,
		Response: resBody,
	}
	defer res.Body.Close()
	return &apiResponse, nil
}

func PostMIMEApiCall(apiRequest commonAppModels.APIRequest) (*commonAppModels.APIResponse, error) {
	var buf = new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("File", apiRequest.FilePath)
	if err != nil {
		log.Error(context.Background(), err, "error in creating multipart form file")
		return nil, err
	}
	_, err = part.Write(apiRequest.Request)
	if err != nil {
		log.Error(context.Background(), err, "error in writing multipart form file")
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		log.Error(context.Background(), err, "error in closing multipart form file")
		return nil, err
	}

	req, err := http.NewRequest("POST", apiRequest.Url, buf)
	if err != nil {
		log.Error(context.Background(), err, "error in creating request for api type '%s'", apiRequest.Type)
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", apiRequest.Authorization)

	client := http.Client{Timeout: apiRequest.Timeout}
	res, err := client.Do(req)
	if err != nil {
		log.Error(context.Background(), err, "error in making api call")
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		log.Error(context.Background(), err, statusError, res.StatusCode)
		return nil, fmt.Errorf(statusError, res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(context.Background(), err, "error in reading api response from response body")
		return nil, err
	}

	apiResponse := commonAppModels.APIResponse{
		Status:   res.Status,
		Response: resBody,
	}
	defer res.Body.Close()
	return &apiResponse, nil
}

func GetRestApiCall(apiRequest commonAppModels.APIRequest) (*commonAppModels.APIResponse, error) {
	req, err := http.NewRequest("GET", apiRequest.Url, bytes.NewBuffer(apiRequest.Request))

	if err != nil {
		log.Error(context.Background(), err, "error in creating request for api type '%s'", apiRequest.Type)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiRequest.Authorization)

	client := http.Client{Timeout: apiRequest.Timeout}
	res, err := client.Do(req)
	if err != nil {
		log.Error(context.Background(), err, "error in making api call")
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		log.Error(context.Background(), err, statusError, res.StatusCode)
		return nil, fmt.Errorf(statusError, res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(context.Background(), err, "error in reading api response from response body")
		return nil, err
	}

	apiResponse := commonAppModels.APIResponse{
		Status:   res.Status,
		Response: resBody,
	}
	defer res.Body.Close()
	return &apiResponse, nil
}
