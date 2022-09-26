package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

var X_AUTH_TOKEN = "e5ab6024-330b-43c9-b5d2-984524a90790"

func TestAuth(t *testing.T) {
	var reqBody = []byte(`{"login":"user","password":"secret"}`)

	req, _ := http.NewRequest(http.MethodPost, "/auth", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	authHandleFunc(w, req)
	res := w.Result()
	defer res.Body.Close()

	var authResp AuthResponseType
	err := json.NewDecoder(res.Body).Decode(&authResp)
	if err != nil {
		fmt.Println(err)
	}
	_, errParse := uuid.Parse(authResp.Token)
	if errParse != nil {
		t.Errorf("Expected uuid token format %v", string(authResp.Token))
	}
}

type CalcOprationTestCaseType struct {
	A              int
	B              int
	Operation      string
	ExpectedResult int
}

var CalcOprationTestCases = []CalcOprationTestCaseType{
	{
		A:              2,
		B:              2,
		Operation:      "/add",
		ExpectedResult: 4,
	},
	{
		A:              6,
		B:              3,
		Operation:      "/sub",
		ExpectedResult: 3,
	},
	{
		A:              3,
		B:              2,
		Operation:      "/mul",
		ExpectedResult: 6,
	},
	{
		A:              4,
		B:              2,
		Operation:      "/div",
		ExpectedResult: 2,
	},
}

func TestCalcOperations(t *testing.T) {

	for _, testCase := range CalcOprationTestCases {
		var reqBody = []byte(fmt.Sprintf(`{"a":%d,"b":%d}`, testCase.A, testCase.B))

		req, _ := http.NewRequest(http.MethodPost, testCase.Operation, bytes.NewBuffer(reqBody))
		req.Header.Add("X-Auth-Token", X_AUTH_TOKEN)

		w := httptest.NewRecorder()
		methodHandleFunc(w, req)
		res := w.Result()
		defer res.Body.Close()

		var opResp OperationResponseType
		err := json.NewDecoder(res.Body).Decode(&opResp)

		if err != nil {
			fmt.Println(err)
		}

		if opResp.Result != testCase.ExpectedResult {
			t.Errorf("Operation %s Expected %v but got %v", testCase.Operation, testCase.ExpectedResult, opResp.Result)
		}
	}
}
