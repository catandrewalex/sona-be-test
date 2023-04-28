package serde_wrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	"sonamusica-backend/errs"
	"sonamusica-backend/logging"
)

type HTTPHandlerSerdeWrapper interface {
	WrapFunc(fn interface{}) http.HandlerFunc

	parseRequest(r *http.Request, t reflect.Type) (reflect.Value, errs.HTTPError)
	handleError(r *http.Request, w http.ResponseWriter, err errs.HTTPError)
	handleSuccess(w http.ResponseWriter, response interface{})
}

type JSONSerdeWrapper struct{}

func NewJSONSerdeWrapper() HTTPHandlerSerdeWrapper {
	return &JSONSerdeWrapper{}
}

func (wrapper JSONSerdeWrapper) WrapFunc(fn interface{}) http.HandlerFunc {
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
			req, err := wrapper.parseRequest(request, reqType)
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

func (wrapper JSONSerdeWrapper) parseRequest(r *http.Request, t reflect.Type) (reflect.Value, errs.HTTPError) {
	if t.Kind() != reflect.Ptr || (t.Kind() == reflect.Ptr && t.Elem().Kind() != reflect.Struct) {
		panic("request type must be a pointer to struct")
	}

	elem := reflect.New(t).Interface() // create a pointer to a zero value request param, e.g. similar to &SignupRequest{}
	err := json.NewDecoder(r.Body).Decode(elem)
	if err != nil {
		return reflect.ValueOf(nil), errs.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("json.NewDecoder(r.Body).Decode(): %v", err), "Does the request contain valid JSON?")
	}

	return reflect.ValueOf(elem).Elem(), nil
}

func (wrapper JSONSerdeWrapper) handleError(r *http.Request, w http.ResponseWriter, err errs.HTTPError) {
	handleError(r, w, err)
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

func handleError(r *http.Request, w http.ResponseWriter, err errs.HTTPError) {
	logging.HTTPServerLogger.Error("Error: %v", err)
	http.Error(w, err.GetClientMessage(), err.GetHTTPErrorCode())
	return
}
