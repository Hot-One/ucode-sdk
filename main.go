package ucodesdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UcodeApis interface {
	/*
		GetList is function that get list of objects from specific table using filter.
		This method works slower because it gets all the information
		about the table, fields and view.
		default_value:
			page = 1
			limit = 10
	*/
	GetList(arg *ArgumentWithPegination) (GetListClientApiResponse, Response, error)
	/*
		GetSingle is function that get one object with all the information of fields, formulas, views and relations.
		It is better to use GetSlim for better performance

		guid="your_guid"
	*/
	GetSingle(arg *Argument) (ClientApiResponse, Response, error)
	/*
		GetListSlim is function that get list of objects from specific table using filter.
		This method works much lighter than GetList because it doesn't get all information about the table, fields and view.
		default_value:
			page = 1
			limit = 10
	*/
	GetListSlim(arg *ArgumentWithPegination) (GetListClientApiResponse, Response, error)
	/*
		GetSingleSlim is function that get one object with its fields.
		It is light and fast to use.

		guid="your_guid"
	*/
	GetSingleSlim(arg *Argument) (ClientApiResponse, Response, error)
	/*
		------------------------WRITE--------------------------------------------------------
	*/
	GetListAggregate(arg *ArgumentWithPegination) (GetListClientApiResponse, Response, error)
	/*
		------------------------WRITE--------------------------------------------------------
	*/
	GetListAggregation(arg *Argument) (GetListAggregationClientApiResponse, Response, error)
	/*
		CreateObject is a function that creates new object.

	*/
	CreateObject(arg *Argument) (Datas, Response, error)
	/*
		UpdateObject is a function that updates specific object
	*/
	UpdateObject(arg *Argument) (ClientApiUpdateResponse, Response, error)
	/*
		MultipleUpdate is a function that updates multiple objects at once
	*/
	MultipleUpdate(arg *Argument) (ClientApiMultipleUpdateResponse, Response, error)
	/*
		Delete is a function that is used to delete one object
		map[guid]="actual_guid"
	*/
	Delete(arg *Argument) (Response, error)
	/*
		MultipleDelete is a function that is used to delete multiple objects
		map[ids]=[list of ids]
	*/
	MultipleDelete(arg *Argument) (Response, error)
	/*
		Send is a function that is used to Send logs to telegram bots
	*/
	Config() *Config
}

type object struct {
	config *Config
}

func New(cfg *Config) UcodeApis {
	return &object{
		config: cfg,
	}
}

func (o *object) CreateObject(arg *Argument) (Datas, Response, error) {
	var (
		response = Response{
			Status: "done",
		}
		createdObject Datas
		url           = fmt.Sprintf("%s/v2/items/%s?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.DisableFaas)
	)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	createObjectResponseInByte, err := o.DoRequest(url, "POST", arg.Request, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Can't send request", "error": err.Error()}
		response.Status = "error"
		return Datas{}, response, err
	}

	err = json.Unmarshal(createObjectResponseInByte, &createdObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling create object", "error": err.Error()}
		response.Status = "error"
		return Datas{}, response, err
	}

	return createdObject, response, nil
}

func (o *object) UpdateObject(arg *Argument) (ClientApiUpdateResponse, Response, error) {
	var (
		response = Response{
			Status: "done",
		}
		updateObject ClientApiUpdateResponse
		url          = fmt.Sprintf("%s/v2/items/%s?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.DisableFaas)
	)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	updateObjectResponseInByte, err := o.DoRequest(url, "PUT", arg.RequestUpdate, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while updating object", "error": err.Error()}
		response.Status = "error"
		return ClientApiUpdateResponse{}, response, err
	}

	err = json.Unmarshal(updateObjectResponseInByte, &updateObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling update object", "error": err.Error()}
		response.Status = "error"
		return ClientApiUpdateResponse{}, response, err
	}

	return updateObject, response, nil
}

func (o *object) MultipleUpdate(arg *Argument) (ClientApiMultipleUpdateResponse, Response, error) {
	var (
		response = Response{
			Status: "done",
		}
		multipleUpdateObject ClientApiMultipleUpdateResponse
		url                  = fmt.Sprintf("%s/v2/items/multiple-update/%s?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.DisableFaas)
	)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	multipleUpdateObjectsResponseInByte, err := o.DoRequest(url, "PUT", arg.Request, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while multiple updating objects", "error": err.Error()}
		response.Status = "error"
		return ClientApiMultipleUpdateResponse{}, response, errors.New("error")
	}

	err = json.Unmarshal(multipleUpdateObjectsResponseInByte, &multipleUpdateObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling multiple update objects", "error": err.Error()}
		response.Status = "error"
		return ClientApiMultipleUpdateResponse{}, response, errors.New("error")
	}

	return ClientApiMultipleUpdateResponse{}, response, nil
}

func (o *object) GetList(arg *ArgumentWithPegination) (GetListClientApiResponse, Response, error) {
	var (
		response      Response
		getListObject GetListClientApiResponse
		url           = fmt.Sprintf("%s/v2/items/get-list/%s?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.DisableFaas)
		page          int
		limit         int
	)

	if arg.Page > 0 {
		page = arg.Page
	}

	if arg.Limit > 0 {
		limit = arg.Limit
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	arg.Request.Data.ObjectData["offset"] = (page - 1) * limit
	arg.Request.Data.ObjectData["limit"] = limit

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	getListResponseInByte, err := o.DoRequest(url, "POST", arg.Request, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Can't sent request", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, err
	}

	err = json.Unmarshal(getListResponseInByte, &getListObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, errors.New("invalid response")
	}

	return getListObject, response, nil
}

func (o *object) GetListSlim(arg *ArgumentWithPegination) (GetListClientApiResponse, Response, error) {
	var (
		response    Response
		listSlim    GetListClientApiResponse
		url         = fmt.Sprintf("%s/v1/object-slim/get-list/%s?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.DisableFaas)
		page, limit int
	)

	reqObject, err := json.Marshal(arg.Request.Data.ObjectData)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while marshalling request getting list slim object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, err
	}

	if arg.Page > 0 {
		page = arg.Page
	}

	if arg.Limit > 0 {
		limit = arg.Limit
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	url = fmt.Sprintf("%s&data=%s&offset=%d&limit=%d", url, string(reqObject), (page-1)*limit, limit)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	getListResponseInByte, err := o.DoRequest(url, "GET", nil, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Can't sent request", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, err
	}

	err = json.Unmarshal(getListResponseInByte, &listSlim)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, errors.New("invalid response")
	}

	return listSlim, response, nil
}

func (o *object) GetSingle(arg *Argument) (ClientApiResponse, Response, error) {
	var (
		response  Response
		getObject ClientApiResponse
		url       = fmt.Sprintf("%s/v2/items/%s/%v?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.Request.Data.ObjectData["guid"], arg.DisableFaas)
	)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	resByte, err := o.DoRequest(url, "GET", nil, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Can't sent request", "error": err.Error()}
		response.Status = "error"
		return ClientApiResponse{}, response, err
	}

	err = json.Unmarshal(resByte, &getObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list object", "error": err.Error()}
		response.Status = "error"
		return ClientApiResponse{}, response, err
	}

	return getObject, response, nil
}

func (o *object) GetSingleSlim(arg *Argument) (ClientApiResponse, Response, error) {
	var (
		response  Response
		getObject ClientApiResponse
		url       = fmt.Sprintf("%s/v1/object-slim/%s/%v?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.Request.Data.ObjectData["guid"], arg.DisableFaas)
	)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	resByte, err := o.DoRequest(url, "GET", nil, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Can't sent request", "error": err.Error()}
		response.Status = "error"
		return ClientApiResponse{}, response, err
	}

	err = json.Unmarshal(resByte, &getObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling to object", "error": err.Error()}
		response.Status = "error"
		return ClientApiResponse{}, response, err
	}

	return getObject, response, nil
}

func (o *object) GetListAggregate(arg *ArgumentWithPegination) (GetListClientApiResponse, Response, error) {
	var (
		response         Response
		getListAggregate GetListClientApiResponse
		url              = fmt.Sprintf("%s/v2/items/get-list-aggregate/%s?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.DisableFaas)
		page, limit      int
	)

	if arg.Limit > 0 {
		limit = arg.Limit
		url = fmt.Sprintf("%s&limit=%d", url, limit)
	}

	if arg.Page > 0 {
		page = arg.Page
		url = fmt.Sprintf("%s&offset=%d", url, (page-1)*limit)
	}

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	getListAggregateResponseInByte, err := o.DoRequest(url, "POST", arg.Request, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"description": string(getListAggregateResponseInByte), "message": "Can't sent request", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, err
	}

	err = json.Unmarshal(getListAggregateResponseInByte, &getListAggregate)
	if err != nil {
		response.Data = map[string]interface{}{"description": string(getListAggregateResponseInByte), "message": "Error while unmarshalling get list object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, err
	}

	return getListAggregate, response, nil
}

func (o *object) GetListAggregation(arg *Argument) (GetListAggregationClientApiResponse, Response, error) {
	var (
		response           Response
		getListAggregation GetListAggregationClientApiResponse
		url                = fmt.Sprintf("%s/v2/items/%s/aggregation", o.config.BaseURL, arg.TableSlug)
	)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	getListAggregationResponseInByte, err := o.DoRequest(url, "POST", arg.Request, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"description": string(getListAggregationResponseInByte), "message": "Can't sent request", "error": err.Error()}
		response.Status = "error"
		return GetListAggregationClientApiResponse{}, response, err
	}

	err = json.Unmarshal(getListAggregationResponseInByte, &getListAggregation)
	if err != nil {
		response.Data = map[string]interface{}{"description": string(getListAggregationResponseInByte), "message": "Error while unmarshalling get list object", "error": err.Error()}
		response.Status = "error"
		return GetListAggregationClientApiResponse{}, response, err
	}

	return getListAggregation, response, nil
}

func (o *object) Delete(arg *Argument) (Response, error) {
	var (
		response = Response{
			Status: "done",
		}
		url = fmt.Sprintf("%s/v2/items/%s/%v?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.Request.Data.ObjectData["guid"], arg.DisableFaas)
	)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	_, err := o.DoRequest(url, "DELETE", nil, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while deleting object", "error": err.Error()}
		response.Status = "error"
		return response, err
	}

	return response, nil
}

func (o *object) MultipleDelete(arg *Argument) (Response, error) {
	var (
		response = Response{
			Status: "done",
		}
		url = fmt.Sprintf("%s/v2/items/%s/?from-ofs=%t", o.config.BaseURL, arg.TableSlug, arg.DisableFaas)
	)

	var appId = o.config.appId
	if arg.AppId != "" {
		appId = arg.AppId
	}

	_, err := o.DoRequest(url, "DELETE", arg.Request.Data.ObjectData, appId, nil)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while deleting objects", "error": err.Error()}
		response.Status = "error"
		return response, err
	}

	return response, nil
}

/*
DoRequest is a function to send http request easily
It gets url, method, body, app_id(for ucode purpose) as paramters

Returns body of the response as array of bytes and error
*/
func (o *object) DoRequest(url string, method string, body interface{}, appId string, headers map[string]string) ([]byte, error) {
	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	if o.config.RequestTimeout > 0 {
		client.Timeout = time.Duration(5 * time.Second)
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	request.Header.Add("authorization", "API-KEY")
	request.Header.Add("X-API-KEY", appId)

	// Add headers from the map
	for key, value := range headers {
		request.Header.Add(key, value)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	return respByte, err
}

func (o *object) Config() *Config {
	return o.config
}
