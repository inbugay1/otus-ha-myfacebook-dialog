package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"myfacebook-dialog/internal/internalapi"
	"myfacebook-dialog/internal/repository"
)

type ListDialog struct {
	DialogRepository repository.DialogRepository
}

type dialogMessage struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

type listDialogRequest struct {
	From string
	To   string
}

func (h *ListDialog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	listDialogReq := h.getListDialogRequest(request)

	err := h.validateListDialogRequest(listDialogReq)
	if err != nil {
		return err
	}

	dialogMessages, err := h.DialogRepository.GetDialogMessagesBySenderIDAndReceiverID(ctx, listDialogReq.From, listDialogReq.To)
	if err != nil {
		return internalapi.NewServerError(fmt.Errorf("list dialog handler, failed to fetch dialoag messages from repository: %w", err))
	}

	listDialogResponse := make([]dialogMessage, 0, len(dialogMessages))

	for _, dialogMsg := range dialogMessages {
		listDialogResponse = append(listDialogResponse, dialogMessage{
			ID:   dialogMsg.ID,
			From: dialogMsg.From,
			To:   dialogMsg.To,
			Text: dialogMsg.Text,
		})
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(&listDialogResponse)
	if err != nil {
		return internalapi.NewServerError(fmt.Errorf("list dialog handler, cannot encode response: %w", err))
	}

	return nil
}

func (h *ListDialog) getListDialogRequest(request *http.Request) listDialogRequest {
	return listDialogRequest{
		From: request.URL.Query().Get("from"),
		To:   request.URL.Query().Get("to"),
	}
}

func (h *ListDialog) validateListDialogRequest(listDialogReq listDialogRequest) error {
	if listDialogReq.From == "" {
		return internalapi.NewInvalidRequestErrorMissingRequiredParameter("from")
	}

	uuidv4Regexp := regexp.MustCompile(`(?i)^[a-f\d]{8}-[a-f\d]{4}-4[a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$`)
	if !uuidv4Regexp.MatchString(listDialogReq.From) {
		return internalapi.NewInvalidRequestErrorInvalidParameter("from", nil)
	}

	if listDialogReq.To == "" {
		return internalapi.NewInvalidRequestErrorMissingRequiredParameter("to")
	}

	if !uuidv4Regexp.MatchString(listDialogReq.To) {
		return internalapi.NewInvalidRequestErrorInvalidParameter("to", nil)
	}

	return nil
}
