package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myfacebook-dialog/internal/internalapi"
	"myfacebook-dialog/internal/repository"
)

type SendDialog struct {
	DialogRepository repository.DialogRepository
}

type sendDialogRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

func (h *SendDialog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	var sendDialogReq sendDialogRequest
	if err := json.NewDecoder(request.Body).Decode(&sendDialogReq); err != nil {
		return internalapi.NewServerError(fmt.Errorf("send dialog handler, cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	err := h.validateSendDialogRequest(sendDialogReq)
	if err != nil {
		return err
	}

	ctx := request.Context()

	dialogMessage := repository.DialogMessage{
		From: sendDialogReq.From,
		To:   sendDialogReq.To,
		Text: sendDialogReq.Text,
	}

	err = h.DialogRepository.Add(ctx, dialogMessage)
	if err != nil {
		return internalapi.NewServerError(fmt.Errorf("send dialog handler, failed to add dialog message to repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	return nil
}

func (h *SendDialog) validateSendDialogRequest(sendDialogReq sendDialogRequest) error {
	if sendDialogReq.Text == "" {
		return internalapi.NewInvalidRequestErrorMissingRequiredParameter("text")
	}

	if sendDialogReq.From == "" {
		return internalapi.NewInvalidRequestErrorMissingRequiredParameter("from")
	}

	if sendDialogReq.From == "" {
		return internalapi.NewInvalidRequestErrorMissingRequiredParameter("to")
	}

	return nil
}
