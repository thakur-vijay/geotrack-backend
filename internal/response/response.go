package response

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type ListResponse[T any] struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       []T    `json:"data"`
}

type ItemResponse[T any] struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       *T     `json:"data"`
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       any    `json:"data,omitempty"`
}

func WriteList[T any](
	w http.ResponseWriter,
	statusCode int,
	message string,
	data []T,
) {
	if data == nil {
		data = []T{}
	}

	write(w, statusCode, message, data)
}

func WriteItem[T any](
	w http.ResponseWriter,
	statusCode int,
	message string,
	data T,
) {
	write(w, statusCode, message, data)
}

func Write(
	w http.ResponseWriter,
	statusCode int,
	message string,
) {
	write(w, statusCode, message, nil)
}

func write(
	w http.ResponseWriter,
	statusCode int,
	message string,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(
		Response{
			StatusCode: statusCode,
			Message:    message,
			Data:       data,
		},
	)
}

func WriteRaw(
	w http.ResponseWriter,
	statusCode int,
	body []byte,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_, _ = w.Write(body)
}

func HandleError(

	w http.ResponseWriter,
	err error,
	errAlreadyProcessing error,
	notFoundMessage string,
	internalMessage string,

) {

	switch {
	case errors.Is(err, errAlreadyProcessing):
		Write(w, http.StatusConflict, err.Error())
	case errors.Is(err, sql.ErrNoRows):
		Write(w, http.StatusNotFound, notFoundMessage)
	default:
		log.Printf("handler error: %v", err)
		Write(w, http.StatusInternalServerError, internalMessage)
	}

}
