// Code generated by "requestgen -method GET -url /sapi/v1/margin/repay -type GetMarginRepayHistoryRequest -responseType .RowsResponse -responseDataField Rows -responseDataType []MarginRepayRecord"; DO NOT EDIT.

package binanceapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

func (g *GetMarginRepayHistoryRequest) Asset(asset string) *GetMarginRepayHistoryRequest {
	g.asset = asset
	return g
}

func (g *GetMarginRepayHistoryRequest) StartTime(startTime time.Time) *GetMarginRepayHistoryRequest {
	g.startTime = &startTime
	return g
}

func (g *GetMarginRepayHistoryRequest) EndTime(endTime time.Time) *GetMarginRepayHistoryRequest {
	g.endTime = &endTime
	return g
}

func (g *GetMarginRepayHistoryRequest) IsolatedSymbol(isolatedSymbol string) *GetMarginRepayHistoryRequest {
	g.isolatedSymbol = &isolatedSymbol
	return g
}

func (g *GetMarginRepayHistoryRequest) Archived(archived bool) *GetMarginRepayHistoryRequest {
	g.archived = &archived
	return g
}

func (g *GetMarginRepayHistoryRequest) Size(size int) *GetMarginRepayHistoryRequest {
	g.size = &size
	return g
}

func (g *GetMarginRepayHistoryRequest) Current(current int) *GetMarginRepayHistoryRequest {
	g.current = &current
	return g
}

// GetQueryParameters builds and checks the query parameters and returns url.Values
func (g *GetMarginRepayHistoryRequest) GetQueryParameters() (url.Values, error) {
	var params = map[string]interface{}{}

	query := url.Values{}
	for _k, _v := range params {
		query.Add(_k, fmt.Sprintf("%v", _v))
	}

	return query, nil
}

// GetParameters builds and checks the parameters and return the result in a map object
func (g *GetMarginRepayHistoryRequest) GetParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}
	// check asset field -> json key asset
	asset := g.asset

	// assign parameter of asset
	params["asset"] = asset
	// check startTime field -> json key startTime
	if g.startTime != nil {
		startTime := *g.startTime

		// assign parameter of startTime
		// convert time.Time to milliseconds time stamp
		params["startTime"] = strconv.FormatInt(startTime.UnixNano()/int64(time.Millisecond), 10)
	} else {
	}
	// check endTime field -> json key endTime
	if g.endTime != nil {
		endTime := *g.endTime

		// assign parameter of endTime
		// convert time.Time to milliseconds time stamp
		params["endTime"] = strconv.FormatInt(endTime.UnixNano()/int64(time.Millisecond), 10)
	} else {
	}
	// check isolatedSymbol field -> json key isolatedSymbol
	if g.isolatedSymbol != nil {
		isolatedSymbol := *g.isolatedSymbol

		// assign parameter of isolatedSymbol
		params["isolatedSymbol"] = isolatedSymbol
	} else {
	}
	// check archived field -> json key archived
	if g.archived != nil {
		archived := *g.archived

		// assign parameter of archived
		params["archived"] = archived
	} else {
	}
	// check size field -> json key size
	if g.size != nil {
		size := *g.size

		// assign parameter of size
		params["size"] = size
	} else {
	}
	// check current field -> json key current
	if g.current != nil {
		current := *g.current

		// assign parameter of current
		params["current"] = current
	} else {
	}

	return params, nil
}

// GetParametersQuery converts the parameters from GetParameters into the url.Values format
func (g *GetMarginRepayHistoryRequest) GetParametersQuery() (url.Values, error) {
	query := url.Values{}

	params, err := g.GetParameters()
	if err != nil {
		return query, err
	}

	for _k, _v := range params {
		if g.isVarSlice(_v) {
			g.iterateSlice(_v, func(it interface{}) {
				query.Add(_k+"[]", fmt.Sprintf("%v", it))
			})
		} else {
			query.Add(_k, fmt.Sprintf("%v", _v))
		}
	}

	return query, nil
}

// GetParametersJSON converts the parameters from GetParameters into the JSON format
func (g *GetMarginRepayHistoryRequest) GetParametersJSON() ([]byte, error) {
	params, err := g.GetParameters()
	if err != nil {
		return nil, err
	}

	return json.Marshal(params)
}

// GetSlugParameters builds and checks the slug parameters and return the result in a map object
func (g *GetMarginRepayHistoryRequest) GetSlugParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}

	return params, nil
}

func (g *GetMarginRepayHistoryRequest) applySlugsToUrl(url string, slugs map[string]string) string {
	for _k, _v := range slugs {
		needleRE := regexp.MustCompile(":" + _k + "\\b")
		url = needleRE.ReplaceAllString(url, _v)
	}

	return url
}

func (g *GetMarginRepayHistoryRequest) iterateSlice(slice interface{}, _f func(it interface{})) {
	sliceValue := reflect.ValueOf(slice)
	for _i := 0; _i < sliceValue.Len(); _i++ {
		it := sliceValue.Index(_i).Interface()
		_f(it)
	}
}

func (g *GetMarginRepayHistoryRequest) isVarSlice(_v interface{}) bool {
	rt := reflect.TypeOf(_v)
	switch rt.Kind() {
	case reflect.Slice:
		return true
	}
	return false
}

func (g *GetMarginRepayHistoryRequest) GetSlugsMap() (map[string]string, error) {
	slugs := map[string]string{}
	params, err := g.GetSlugParameters()
	if err != nil {
		return slugs, nil
	}

	for _k, _v := range params {
		slugs[_k] = fmt.Sprintf("%v", _v)
	}

	return slugs, nil
}

func (g *GetMarginRepayHistoryRequest) Do(ctx context.Context) ([]MarginRepayRecord, error) {

	// empty params for GET operation
	var params interface{}
	query, err := g.GetParametersQuery()
	if err != nil {
		return nil, err
	}

	apiURL := "/sapi/v1/margin/repay"

	req, err := g.client.NewAuthenticatedRequest(ctx, "GET", apiURL, query, params)
	if err != nil {
		return nil, err
	}

	response, err := g.client.SendRequest(req)
	if err != nil {
		return nil, err
	}

	var apiResponse RowsResponse
	if err := response.DecodeJSON(&apiResponse); err != nil {
		return nil, err
	}
	var data []MarginRepayRecord
	if err := json.Unmarshal(apiResponse.Rows, &data); err != nil {
		return nil, err
	}
	return data, nil
}
