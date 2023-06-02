package serde_wrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"sonamusica-backend/errs"
	"sonamusica-backend/logging"
	"sonamusica-backend/service/output"

	"github.com/go-chi/chi/v5"
)

type HTTPHandlerSerdeWrapper interface {
	// WrapFunc accepts function with signature:
	//   1. Request: must exist in: [(ctx context.Context), (ctx context.Context, req <any_Go_struct>)]
	//   2. Response: must exist in: [(err error), (res <any_go_struct>, err Error)]
	//
	// We can pass urlParamKeys if the request comes from dynamic URL (URL with parameter), e.g..:
	//   1. /user/1 --> pass ["id"] as urlParamKeys
	//   2. /user/myusername --> pass ["username"] as urlParamKeys
	WrapFunc(fn interface{}, urlParamKeys ...string) http.HandlerFunc

	// parseRequest is a internal helper function for WrapFunc, which also accepts urlParamKeys
	parseRequest(r *http.Request, rType reflect.Type, urlParamKeys ...string) (reflect.Value, errs.HTTPError)
	handleError(r *http.Request, w http.ResponseWriter, err errs.HTTPError)
	handleSuccess(w http.ResponseWriter, response interface{})
}

type JSONSerdeWrapper struct{}

func NewJSONSerdeWrapper() HTTPHandlerSerdeWrapper {
	return &JSONSerdeWrapper{}
}

func (wrapper JSONSerdeWrapper) WrapFunc(fn interface{}, urlParamKeys ...string) http.HandlerFunc {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("input parameter is not a function")
	}
	fnValue := reflect.ValueOf(fn)

	fnName := runtime.FuncForPC(fnValue.Pointer()).Name()
	debugInfo := fmt.Sprintf(" fnName: %s", fnName)

	// Check function inputs
	fnNumIn := fnType.NumIn()
	if fnNumIn < 1 {
		panic(fmt.Sprintf("the function must have at least one input: context.%s", debugInfo))
	} else if fnNumIn > 2 {
		panic(fmt.Sprintf("the function can have at most two inputs: context and request.%s", debugInfo))
	}
	ctxType := fnType.In(0)
	var reqType reflect.Type
	if fnNumIn == 2 {
		reqType = fnType.In(1)
	}
	if !ctxType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		panic(fmt.Sprintf("context type must implement 'context.Context'.%s", debugInfo))
	}
	if reqType != nil && !(reqType.Kind() == reflect.Struct || (reqType.Kind() == reflect.Ptr && reqType.Elem().Kind() == reflect.Struct)) {
		panic(fmt.Sprintf("request type must be a struct or a pointer to struct.%s", debugInfo))
	}

	// Check function outputs
	fnNumOut := fnType.NumOut()
	if fnNumOut < 1 {
		panic(fmt.Sprintf("the function must have at least one output: error.%s", debugInfo))
	} else if fnNumOut > 2 {
		panic(fmt.Sprintf("the function can have at most two outputs: response and error.%s", debugInfo))
	}
	var resType, errType reflect.Type
	if fnNumOut == 1 {
		errType = fnType.Out(0)
	} else {
		resType, errType = fnType.Out(0), fnType.Out(1)
	}

	if resType.Kind() != reflect.Ptr || resType.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("response type must be a pointer to a struct.%s", debugInfo))
	}

	if !errType.Implements(reflect.TypeOf((*errs.HTTPError)(nil)).Elem()) {
		panic(fmt.Sprintf("error type must implement HTTPError type.%s", debugInfo))
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		inputs := make([]reflect.Value, 0, fnNumIn)
		inputs = append(inputs, reflect.ValueOf(request.Context()))

		// Parse the request object if exists
		if reqType != nil {
			req, err := wrapper.parseRequest(request, reqType, urlParamKeys...)
			if err != nil {
				wrapper.handleError(request, writer, err)
				return
			}

			inputs = append(inputs, req)
		}

		// Call the function and get its outputs
		outputs := fnValue.Call(inputs)

		var resValue, errValue reflect.Value
		if resType == nil { // means only error
			errValue = outputs[0]
		} else {
			resValue, errValue = outputs[0], outputs[1]
		}

		if !errValue.IsNil() {
			wrapper.handleError(request, writer, errValue.Interface().(errs.HTTPError))
			return
		}

		// If there's no expectation for response body, we can stop here. The response will have status code 200.
		if resType == nil {
			return
		}

		if resValue.IsNil() {
			panic(fmt.Sprintf("response should not be nil if there is no error, path: %v", request.URL.Path))
		}

		response := resValue.Interface()
		wrapper.handleSuccess(writer, response)
	})
}

func (wrapper JSONSerdeWrapper) parseRequest(r *http.Request, rType reflect.Type, urlParamKeys ...string) (reflect.Value, errs.HTTPError) {
	if rType.Kind() != reflect.Ptr || (rType.Kind() == reflect.Ptr && rType.Elem().Kind() != reflect.Struct) {
		panic("request type must be a pointer to struct")
	}

	elem := reflect.New(rType).Interface() // create a pointer to a zero value request param, e.g. similar to &SignupRequest{}

	// we accept input parameters from:
	//   1. request body (only JSON), or
	//   2. URL query param
	if r.Header.Get("Content-Type") == "application/json" {
		err := json.NewDecoder(r.Body).Decode(elem)
		if err != nil {
			return reflect.ValueOf(nil), errs.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("json.NewDecoder(r.Body).Decode(): %v", err), map[string]string{errs.ClientMessageKey_NonField: "Does the request contain valid JSON, and valid value types?"}, "")
		}
	} else {
		urlQueryInJSON := convertURLQueryToJSONString(r.URL.Query().Encode())
		err := json.Unmarshal(urlQueryInJSON, elem)
		if err != nil {
			return reflect.ValueOf(nil), errs.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("json.Unmarshal(urlQueryInJSON): %v", err), map[string]string{errs.ClientMessageKey_NonField: "The request doesn't contain JSON and has invalid URL query params!"}, "")
		}
	}

	elemValue := reflect.ValueOf(elem).Elem()

	// Set field value from URL params
	for _, urlParamKey := range urlParamKeys {
		urlParamValue := chi.URLParam(r, urlParamKey)
		logging.HTTPServerLogger.Debug("Found urlParam: key = %q, value = %q", urlParamKey, urlParamValue)
		if len(urlParamValue) == 0 {
			panic(fmt.Sprintf("found non-existing urlParamKey = %q. please check your URL pattern", urlParamKey))
		}

		elemField := reflect.Indirect(elemValue).FieldByName(urlParamKey)
		if elemField.IsValid() {
			if !elemField.IsZero() {
				logging.HTTPServerLogger.Warn("Found pre-populated struct field = %q, check the JSON body or URL query param for the duplicated key. Discarding the URL param value.", urlParamKey)
				continue
			}

			switch elemField.Kind() {
			case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
				valueInt, err := strconv.ParseInt(urlParamValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("incompatible urlParam: key = %q, value = %q can't be converted to int", urlParamKey, urlParamValue))
				}
				elemField.SetInt(valueInt)
			case reflect.String:
				elemField.SetString(urlParamValue)
			default:
				panic(fmt.Sprintf("unsupported field kind with name = %q. list of supported kind: [int, string]", urlParamKey))
			}
		} else {
			panic(fmt.Sprintf("found non-existing request field (urlParamKey = %q). please check your request struct: urlParamKey must match struct's field 'name', NOT 'json tag', and is case sensitive", urlParamKey))
		}

	}

	return elemValue, nil
}

func convertURLQueryToJSONString(encodedURLQuery string) []byte {
	jsonStruct := make(map[string]interface{}, 0)

	queries := strings.Split(encodedURLQuery, "&")
	for _, query := range queries {
		splittedQuery := strings.Split(query, "=")
		if len(splittedQuery) == 2 {
			key, value := splittedQuery[0], splittedQuery[1]
			if valueInt, err := strconv.Atoi(value); err == nil {
				jsonStruct[key] = valueInt
			} else if valueFloat, err := strconv.ParseFloat(value, 64); err == nil {
				jsonStruct[key] = valueFloat
			} else {
				jsonStruct[key] = value
			}
		}
	}

	jsonString, err := json.Marshal(jsonStruct)
	if err != nil {
		panic(fmt.Sprintf("error on json.Marshal() while parsing URL Query=%q", encodedURLQuery))
	}

	return jsonString
}

func (wrapper JSONSerdeWrapper) handleError(r *http.Request, w http.ResponseWriter, httpErr errs.HTTPError) {
	logging.HTTPServerLogger.Error("Error: %v", httpErr)

	errResponse := output.ErrorResponse{
		Errors:  httpErr.GetProcessableErrors(),
		Message: httpErr.GetClientMessage(),
	}

	resBytes, err := json.Marshal(errResponse)
	if err != nil {
		logging.HTTPServerLogger.Error("Error on json.Marshal(): %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	errorJSON(w, string(resBytes), httpErr.GetHTTPErrorCode())
}

func (wrapper JSONSerdeWrapper) handleSuccess(w http.ResponseWriter, response interface{}) {
	if response == nil {
		return
	}

	resBytes, err := json.Marshal(response)
	if err != nil {
		logging.HTTPServerLogger.Error("Error on json.Marshal(): %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resBytes)
	if err != nil {
		logging.HTTPServerLogger.Error("Error on http.ResponseWriter.Write(): %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func errorJSON(w http.ResponseWriter, jsonBody string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, jsonBody)
}

// Disabled to be used later by another non-JSON serde_wrapper
// func handleError(r *http.Request, w http.ResponseWriter, err errs.HTTPError) {
// 	logging.HTTPServerLogger.Error("Error: %v", err)
// 	http.Error(w, err.GetClientMessage(), err.GetHTTPErrorCode())
// 	return
// }
