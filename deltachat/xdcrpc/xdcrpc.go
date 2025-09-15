// Package helping with the communication between bots and WebXDC apps via JSON-RPC 1.0
package xdcrpc

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/chatmail/rpc-client-go/deltachat"
)

type ErrorCode int

const (
	// The method does not exist / is not available
	MethodNotFoud ErrorCode = -32601
	// Invalid JSON was received by the server
	ParseError ErrorCode = -32700
	// The JSON sent is not a valid Request object
	InvalidRequest ErrorCode = -32600
	// Invalid method parameter(s)
	InvalidParams ErrorCode = -32602
)

// Request sent by the frontend app
type Request struct {
	Id     string `json:"id,omitempty"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type _Request struct {
	Id     string            `json:"id,omitempty"`
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

// Response sent by the bot
type Response struct {
	Id     string `json:"id,omitempty"`
	Result any    `json:"result"`
	Error  *Error `json:"error"`
}

// Error data sent by the bot in ErrorResponse
type Error struct {
	Code    ErrorCode `json:"code,omitempty"`
	Message string    `json:"message,omitempty"`
	Data    any       `json:"data,omitempty"`
}

type StatusUpdate[T any] struct {
	Info      string `json:"info,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Document  string `json:"document,omitempty"`
	Payload   T      `json:"payload,omitempty"`
	Serial    uint   `json:"serial,omitempty"`
	MaxSerial uint   `json:"max_serial,omitempty"`
}

type SelfMessageErr struct {
}

func (error *SelfMessageErr) Error() string {
	return "RPC message seems to be from self"
}

func HandleMessage(rpc *deltachat.Rpc, accId deltachat.AccountId, msgId deltachat.MsgId, serial uint, api any) error {
	rawUpdate, err := GetUpdate(rpc, accId, msgId, serial)
	if err != nil {
		return err
	}
	if IsFromSelf(rawUpdate) {
		return &SelfMessageErr{}
	}
	if response := GetResponse(api, rawUpdate); response != nil {
		return SendPayload(rpc, accId, msgId, response)
	}
	return nil
}

func GetResponse(api any, rawUpdate []byte) *Response {
	response := &Response{}
	var update StatusUpdate[_Request]
	err := json.Unmarshal(rawUpdate, &update)
	if err != nil {
		response.Error = &Error{Code: ParseError, Message: "Parse error"}
		return response
	}
	request := update.Payload
	response.Id = request.Id

	valOf := reflect.ValueOf(api)
	method := valOf.MethodByName(request.Method)
	if !method.IsValid() || method.IsNil() {
		if request.Id != "" {
			response.Error = &Error{Code: MethodNotFoud, Message: "Method not found"}
			return response
		}
		return nil
	}

	invalidParamsErr := &Error{Code: InvalidParams, Message: "Invalid params"}

	argsCount := method.Type().NumIn()
	if len(request.Params) != argsCount {
		response.Error = invalidParamsErr
		return response
	}

	callArgs := make([]reflect.Value, argsCount)
	if argsCount > 0 {
		for i := 0; i < argsCount; i++ {
			argType := method.Type().In(i)
			val := reflect.New(argType).Interface()
			err = json.Unmarshal(request.Params[i], val)
			if err != nil {
				response.Error = invalidParamsErr
				return response
			}
			callArgs[i] = reflect.ValueOf(val).Elem()
		}
	}

	result := method.Call(callArgs)
	if request.Id == "" {
		return nil
	}

	var returnValues []any
	for _, valOf := range result {
		valAny := valOf.Interface()
		switch val := valAny.(type) {
		case *Error:
			response.Error = val
		default:
			returnValues = append(returnValues, val)
		}
	}
	count := len(returnValues)
	switch {
	case count == 1:
		response.Result = returnValues[0]
	case count > 1:
		response.Result = returnValues
	}
	return response
}

// Return true if the raw status update is from self, false otherwise
func IsFromSelf(rawUpdate []byte) bool {
	var update StatusUpdate[map[string]json.RawMessage]
	if err := json.Unmarshal(rawUpdate, &update); err != nil {
		return false
	}
	if _, ok := update.Payload["result"]; ok {
		return true
	}
	if _, ok := update.Payload["error"]; ok {
		return true
	}
	return false
}

// Get all setatus updates with serial greater than the given serial
func GetUpdates(rpc *deltachat.Rpc, accId deltachat.AccountId, msgId deltachat.MsgId, serial uint) ([]json.RawMessage, error) {
	var rawUpdates []json.RawMessage
	data, err := rpc.GetWebxdcStatusUpdates(accId, msgId, serial)
	if err != nil {
		return rawUpdates, err
	}
	err = json.Unmarshal([]byte(data), &rawUpdates)
	return rawUpdates, err
}

// Get the status update with the given serial
func GetUpdate(rpc *deltachat.Rpc, accId deltachat.AccountId, msgId deltachat.MsgId, serial uint) (json.RawMessage, error) {
	rawUpdates, err := GetUpdates(rpc, accId, msgId, serial-1)
	if err != nil {
		return nil, err
	}
	if len(rawUpdates) > 0 {
		return rawUpdates[0], nil
	}
	return nil, errors.New("no new status update was found")
}

// Send a WebXDC status update
func SendUpdate[T any](rpc *deltachat.Rpc, accId deltachat.AccountId, msgId deltachat.MsgId, update StatusUpdate[T], description string) error {
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}
	return rpc.SendWebxdcStatusUpdate(accId, msgId, string(data), description)
}

// Send a WebXDC status update with the given payload
func SendPayload[T any](rpc *deltachat.Rpc, accId deltachat.AccountId, msgId deltachat.MsgId, payload T) error {
	return SendUpdate(rpc, accId, msgId, StatusUpdate[T]{Payload: payload}, "")
}
