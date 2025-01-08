package output

import (
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/user_action_log"
	"sonamusica-backend/errs"
)

const (
	MaxPage_FetchUserActionLogs           = -1
	MaxResultsPerPage_FetchUserActionLogs = 1000
)

type FetchUserActionLogsRequest struct {
	PaginationRequest
	TimeFilter
	UserID        identity.UserID            `json:"userID,omitempty"`
	PrivilegeType identity.UserPrivilegeType `json:"privilegeType,omitempty"`
	Endpoint      string                     `json:"endpoint,omitempty"`
	Method        string                     `json:"method,omitempty"`
	StatusCode    uint16                     `json:"statusCode,omitempty"`
}
type FetchUserActionLogsResponse struct {
	Data    FetchUserActionLogsResult `json:"data"`
	Message string                    `json:"message,omitempty"`
}
type FetchUserActionLogsResult struct {
	Results []user_action_log.UserActionLog `json:"results"`
	PaginationResponse
}

func (r FetchUserActionLogsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_FetchUserActionLogs, MaxResultsPerPage_FetchUserActionLogs); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if validationErr := r.TimeFilter.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}
