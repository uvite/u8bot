// Code generated by "requestgen -method GET -responseType .APIResponse -responseDataField Data -type GetOrderBookLevel2DepthAllRequest -url /api/v3/market/orderbook/level2 -responseDataType .OrderBook"; DO NOT EDIT.

package kucoinapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
)

func (g *GetOrderBookLevel2DepthAllRequest) Symbol(symbol string) *GetOrderBookLevel2DepthAllRequest {
	g.symbol = symbol
	return g
}

// GetQueryParameters builds and checks the query parameters and returns url.Values
func (g *GetOrderBookLevel2DepthAllRequest) GetQueryParameters() (url.Values, error) {
	var params = map[string]interface{}{}
	// check symbol field -> json key symbol
	symbol := g.symbol

	// assign parameter of symbol
	params["symbol"] = symbol

	query := url.Values{}
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}

	return query, nil
}

// GetParameters builds and checks the parameters and return the result in a map object
func (g *GetOrderBookLevel2DepthAllRequest) GetParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}

	return params, nil
}

// GetParametersQuery converts the parameters from GetParameters into the url.Values format
func (g *GetOrderBookLevel2DepthAllRequest) GetParametersQuery() (url.Values, error) {
	query := url.Values{}

	params, err := g.GetParameters()
	if err != nil {
		return query, err
	}

	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}

	return query, nil
}

// GetParametersJSON converts the parameters from GetParameters into the JSON format
func (g *GetOrderBookLevel2DepthAllRequest) GetParametersJSON() ([]byte, error) {
	params, err := g.GetParameters()
	if err != nil {
		return nil, err
	}

	return json.Marshal(params)
}

// GetSlugParameters builds and checks the slug parameters and return the result in a map object
func (g *GetOrderBookLevel2DepthAllRequest) GetSlugParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}

	return params, nil
}

func (g *GetOrderBookLevel2DepthAllRequest) applySlugsToUrl(url string, slugs map[string]string) string {
	for k, v := range slugs {
		needleRE := regexp.MustCompile(":" + k + "\\b")
		url = needleRE.ReplaceAllString(url, v)
	}

	return url
}

func (g *GetOrderBookLevel2DepthAllRequest) GetSlugsMap() (map[string]string, error) {
	slugs := map[string]string{}
	params, err := g.GetSlugParameters()
	if err != nil {
		return slugs, nil
	}

	for k, v := range params {
		slugs[k] = fmt.Sprintf("%v", v)
	}

	return slugs, nil
}

func (g *GetOrderBookLevel2DepthAllRequest) Do(ctx context.Context) (*OrderBook, error) {

	// no body params
	var params interface{}
	query, err := g.GetQueryParameters()
	if err != nil {
		return nil, err
	}

	apiURL := "/api/v3/market/orderbook/level2"

	req, err := g.client.NewAuthenticatedRequest(ctx, "GET", apiURL, query, params)
	if err != nil {
		return nil, err
	}

	response, err := g.client.SendRequest(req)
	if err != nil {
		return nil, err
	}

	var apiResponse APIResponse
	if err := response.DecodeJSON(&apiResponse); err != nil {
		return nil, err
	}
	var data OrderBook
	if err := json.Unmarshal(apiResponse.Data, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
