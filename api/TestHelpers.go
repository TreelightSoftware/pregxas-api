package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi"
)

type fn func() *chi.Mux

// TestAPICall allows an easy way to test HTTP end points in unit testing
func TestAPICall(method string, endpoint string, data io.Reader, handler http.HandlerFunc, jwt string, secretKey string) (code int, body *bytes.Buffer, err error) {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	req, err := http.NewRequest(method, endpoint, data)
	if err != nil {
		return 500, nil, err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	if jwt != "" {
		req.Header.Add("jwt", jwt)
	}
	if secretKey != "" {
		req.Header.Add("X-API-SECRET", secretKey)
	}
	rr := httptest.NewRecorder()

	chi := SetupApp()
	chi.ServeHTTP(rr, req)

	return rr.Code, rr.Body, nil
}

// UnmarshalTestMap helps to unmarshal the request for the testing calls
func UnmarshalTestMap(body *bytes.Buffer) (PregxasAPIReturn, map[string]interface{}, error) {
	ret := PregxasAPIReturn{}
	retBuf := new(bytes.Buffer)
	retBuf.ReadFrom(body)
	err := json.Unmarshal(retBuf.Bytes(), &ret)
	if err != nil {
		return ret, map[string]interface{}{}, err
	}
	retBody, ok := ret.Data.(map[string]interface{})
	if !ok {
		return ret, map[string]interface{}{}, errors.New("Could not convert")
	}

	return ret, retBody, nil
}

// UnmarshalTestArray unmarshals a response that is an array in the data field
func UnmarshalTestArray(body *bytes.Buffer) (PregxasAPIReturn, []interface{}, error) {
	ret := PregxasAPIReturn{}
	retBuf := new(bytes.Buffer)
	retBuf.ReadFrom(body)
	err := json.Unmarshal(retBuf.Bytes(), &ret)
	if err != nil {
		return ret, []interface{}{}, err
	}
	retBody, ok := ret.Data.([]interface{})
	if !ok {
		return ret, []interface{}{}, errors.New("Could not convert")
	}

	return ret, retBody, nil
}

func convertTestJSONFloatToInt(input interface{}) (int64, error) {
	i, ok := input.(float64)
	if !ok {
		return 0, errors.New("could not convert")
	}
	return int64(i), nil
}
