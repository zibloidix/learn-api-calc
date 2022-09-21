package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type AuthRequestType struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

type AuthResponseType struct {
	Token string `json:"token"`
}

type ErrorType struct {
	ErrorID      string `json:"errorId"`
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    string `json:"errorCode"`
}

type ErrorsResponseType struct {
	Errors *map[string]ErrorType `json:"errors"`
}

type TokensResponseType struct {
	Tokens []string `json:"tags"`
}

type OperationRequestType struct {
	A int `json:"a"`
	B int `json:"b"`
}

type OperationResponseType struct {
	A         int    `json:"a"`
	B         int    `json:"b"`
	Result    int    `json:"result"`
	Operation string `json:"operation"`
}

var ErrorsMap = make(map[string]ErrorType)
var Tokens []string

func main() {
	port := os.Getenv("PORT")

	ErrorsMap["AUTH_EMPTY_FIELDS"] = ErrorType{
		ErrorMessage: "Одно из полей запроса не передено или содержит пустую строку в качестве значения",
		ErrorCode:    "AUTH_EMPTY_FIELDS",
	}
	ErrorsMap["AUTH_WRONG_VALUES"] = ErrorType{
		ErrorMessage: "Значения полей login и password указаны неверно (необходимо указать user:secret)",
		ErrorCode:    "AUTH_WRONG_VALUES",
	}
	ErrorsMap["TOKEN_EMPTY"] = ErrorType{
		ErrorMessage: "Заголовок X-Auth-Token не содержит значения или не указан в запросе",
		ErrorCode:    "TOKEN_EMPTY",
	}
	ErrorsMap["TOKEN_IS_NOT_VALID"] = ErrorType{
		ErrorMessage: "Значение заголовока 'X-Auth-Token' не соответствует формату UUID",
		ErrorCode:    "TOKEN_IS_NOT_VALID",
	}
	ErrorsMap["TOKEN_NOT_FOUND"] = ErrorType{
		ErrorMessage: "Значение заголовока X-Auth-Token не найдено в системе",
		ErrorCode:    "TOKEN_NOT_FOUND",
	}
	ErrorsMap["ATTR_EMPTY"] = ErrorType{
		ErrorMessage: "Один из атрибутов операции не передан или содержит null в качестве значения",
		ErrorCode:    "ATTR_EMPTY",
	}
	ErrorsMap["ATTR_ZERO"] = ErrorType{
		ErrorMessage: "Один из атрибутов операции деления равен нулю",
		ErrorCode:    "ATTR_ZERO",
	}
	ErrorsMap["WRONG_REQUEST"] = ErrorType{
		ErrorMessage: "Запрос содержит ошибки и не может быть обработан",
		ErrorCode:    "WRONG_REQUEST",
	}

	http.HandleFunc("/auth", authHandleFunc)
	http.HandleFunc("/add", methodHandleFunc)
	http.HandleFunc("/sub", methodHandleFunc)
	http.HandleFunc("/mul", methodHandleFunc)
	http.HandleFunc("/div", methodHandleFunc)
	http.HandleFunc("/errors", errorsHandleFunc)
	http.HandleFunc("/tokens", tokensHandleFunc)

	http.ListenAndServe(":"+port, nil)
}

func authHandleFunc(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var authreq AuthRequestType
	err := json.NewDecoder(req.Body).Decode(&authreq)
	if err != nil {
		sendErrorResponce(w, "WRONG_REQUEST", http.StatusBadRequest)
	} else {
		if !isAuthWrongValues(&authreq) {
			token := uuid.New().String()
			addToken(token)
			authresp := AuthResponseType{
				Token: token,
			}
			json.NewEncoder(w).Encode(authresp)
		} else if isAuthEmptyFields(&authreq) {
			sendErrorResponce(w, "AUTH_EMPTY_FIELDS", http.StatusUnauthorized)
		} else if isAuthWrongValues(&authreq) {
			sendErrorResponce(w, "AUTH_WRONG_VALUES", http.StatusUnauthorized)
		}
	}
}

func isAuthEmptyFields(authreq *AuthRequestType) bool {
	return len(authreq.Login) == 0 || len(authreq.Password) == 0
}
func isAuthWrongValues(authreq *AuthRequestType) bool {
	return authreq.Login != "user" || authreq.Password != "secret"
}

func addToken(token string) {
	if len(Tokens) >= 50 {
		Tokens = Tokens[1:]
	}
	Tokens = append(Tokens, token)
}

func errorsHandleFunc(w http.ResponseWriter, req *http.Request) {
	var e ErrorsResponseType
	e.Errors = &ErrorsMap
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&e)
}

func tokensHandleFunc(w http.ResponseWriter, req *http.Request) {
	var t TokensResponseType
	t.Tokens = Tokens[:]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&t)
}

func methodHandleFunc(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token := req.Header.Get("X-Auth-Token")
	if isAuthOK(token, w) {
		operation := req.URL.Path
		var values OperationRequestType
		values.A = -1000001
		values.B = 1000001

		err := json.NewDecoder(req.Body).Decode(&values)
		if err != nil {
			sendErrorResponce(w, "WRONG_REQUEST", http.StatusBadRequest)
		} else {
			if isAttrsEmpty(values.A, values.B) {
				sendErrorResponce(w, "ATTR_EMPTY", http.StatusUnprocessableEntity)
			} else if isDivOperation(operation) && isAttrNotZero(values.A, values.B) {
				sendErrorResponce(w, "ATTR_ZERO", http.StatusUnprocessableEntity)
			} else {
				var result OperationResponseType
				result.A = values.A
				result.B = values.B
				result.Operation = operation[1:]
				switch operation {
				case "/add":
					result.Result = addition(values.A, values.B)
				case "/sub":
					result.Result = subtraction(values.A, values.B)
				case "/mul":
					result.Result = multiplication(values.A, values.B)
				case "/div":
					result.Result = division(values.A, values.B)
				}
				json.NewEncoder(w).Encode(&result)
			}
		}
	}
}

func isDivOperation(operation string) bool {
	return operation == "/div"
}

func isAttrNotZero(a, b int) bool {
	return a == 0 || b == 0
}

func isAuthOK(token string, w http.ResponseWriter) bool {
	if isTokenEmpty(token) {
		sendErrorResponce(w, "TOKEN_EMPTY", http.StatusForbidden)
		return false
	} else if isTokenNotValid(token) {
		sendErrorResponce(w, "TOKEN_IS_NOT_VALID", http.StatusUnauthorized)
		return false
	} else if isTokenNotExists(token) {
		sendErrorResponce(w, "TOKEN_NOT_FOUND", http.StatusForbidden)
		return false
	}
	return true
}

func isAttrsEmpty(a, b int) bool {
	return a == -1000001 || b == 1000001
}

func isTokenNotExists(token string) bool {
	for _, v := range Tokens {
		if v == token {
			return false
		}
	}
	return true
}

func isTokenEmpty(token string) bool {
	return len(token) == 0
}

func isTokenNotValid(token string) bool {
	_, err := uuid.Parse(token)
	return err != nil
}

func addition(a int, b int) int {
	return a + b
}

func subtraction(a int, b int) int {
	return a - b
}

func multiplication(a int, b int) int {
	return a * b
}

func division(a int, b int) int {
	return a / b
}

func sendErrorResponce(w http.ResponseWriter, errorCode string, status int) {
	err := ErrorsMap[errorCode]
	err.ErrorID = uuid.New().String()
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(err)
}
