package output

import (
	"fmt"
	"time"

	"sonamusica-backend/app-service/util"
	"sonamusica-backend/errs"
)

const (
	Default_MaxPage           = 50
	Default_MaxResultsPerPage = 10000
)

type YearMonthFilterType string

const (
	YearMonthFilterType_Standard          YearMonthFilterType = "STANDARD" // First to last day of current month
	YearMonthFilterType_CalculatingSalary YearMonthFilterType = "SALARY"   // 26th previous month to 25th current month
)

type ErrorResponse struct {
	Errors  map[string]string `json:"errors"`
	Message string            `json:"message,omitempty"`
}

type PaginationRequest struct {
	Page           int `json:"page"`
	ResultsPerPage int `json:"resultsPerPage"`
}

func (r PaginationRequest) Validate(maxPage int, maxResultsPerPage int) errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.Page <= 0 {
		errorDetail["page"] = "page must be greater than 0"
	} else if maxPage > 0 && r.Page > maxPage {
		errorDetail["page"] = fmt.Sprintf("page must be less than %d", maxPage)
	}
	if maxResultsPerPage > 0 && r.Page > maxResultsPerPage {
		errorDetail["resultsPerPage"] = "resultsPerPage must be greater than 0"
	} else if maxResultsPerPage > 0 && r.ResultsPerPage > maxResultsPerPage {
		errorDetail["resultsPerPage"] = fmt.Sprintf("resultsPerPage must be less than %d", maxPage)
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type PaginationResponse struct {
	TotalPages   int `json:"totalPages"`
	TotalResults int `json:"totalResults"`
	CurrentPage  int `json:"currentPage"`
}

func NewPaginationResponse(paginationResult util.PaginationResult) PaginationResponse {
	return PaginationResponse{
		TotalPages:   paginationResult.TotalPages,
		TotalResults: paginationResult.TotalResults,
		CurrentPage:  paginationResult.CurrentPage,
	}
}

type TimeFilter struct {
	StartDatetime time.Time `json:"startDatetime,omitempty"`
	EndDatetime   time.Time `json:"endDatetime,omitempty"`
}

func (r TimeFilter) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if !r.EndDatetime.IsZero() && r.StartDatetime.After(r.EndDatetime) {
		errorDetail["startDatetime"] = "startDatetime must be less than endDatetime"
		errorDetail["endDatetime"] = "endDatetime must be greater than startDatetime"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type YearMonthFilter struct {
	Year  int `json:"year,omitempty"`
	Month int `json:"month,omitempty"`
}

func (d YearMonthFilter) Validate() errs.ValidationError {
	// we allow empty YearMonthFilter
	if d.Year == 0 && d.Month == 0 {
		return nil
	}

	errorDetail := make(errs.ValidationErrorDetail, 0)
	if d.Year < 1960 || d.Year > 2100 {
		errorDetail["year"] = "year must be in between 1960-2100"
	}
	if d.Month != 0 { // we allow empty month, for a year-wide filter (e.g. year=2023, month=0, which means a whole year filter across 2023)
		if d.Month < 1 || d.Month > 12 {
			errorDetail["month"] = "month must be in between 1-12"
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

func (d YearMonthFilter) ToTimeFilter(filterType YearMonthFilterType) TimeFilter {
	if d.Year == 0 && d.Month == 0 { // an all-time filter
		return TimeFilter{}
	}

	timeFilter := TimeFilter{}
	if d.Year != 0 && d.Month == 0 { // a whole year filter
		timeFilter = TimeFilter{
			StartDatetime: time.Date(d.Year, 1, 1, 0, 0, 0, 0, util.DefaultTimezone),
			EndDatetime:   time.Date(d.Year, 1, 1, 23, 59, 59, 0, util.DefaultTimezone).AddDate(1, 0, -1),
		}
	} else { // a monthly filter
		switch filterType {
		case YearMonthFilterType_Standard:
			timeFilter = TimeFilter{
				StartDatetime: time.Date(d.Year, time.Month(d.Month), 1, 0, 0, 0, 0, util.DefaultTimezone),
				EndDatetime:   time.Date(d.Year, time.Month(d.Month), 1, 23, 59, 59, 0, util.DefaultTimezone).AddDate(0, 1, -1),
			}
		case YearMonthFilterType_CalculatingSalary:
			timeFilter = TimeFilter{
				StartDatetime: time.Date(d.Year, time.Month(d.Month), 26, 0, 0, 0, 0, util.DefaultTimezone).AddDate(0, -1, 0),
				EndDatetime:   time.Date(d.Year, time.Month(d.Month), 25, 23, 59, 59, 0, util.DefaultTimezone),
			}
		}
	}
	return timeFilter
}

type YearMonthRangeFilter struct {
	StartDate YearMonthFilter `json:"startDate"`
	EndDate   YearMonthFilter `json:"endDate"`
}

func (r YearMonthRangeFilter) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	// all years & months must be filled
	if r.StartDate.Year == 0 || r.StartDate.Month == 0 {
		errorDetail["startDate.year"] = "year must not be empty"
		errorDetail["startDate.month"] = "month must not be empty"
	}
	if r.EndDate.Year == 0 || r.EndDate.Month == 0 {
		errorDetail["endDate.year"] = "year must not be empty"
		errorDetail["endDate.month"] = "month must not be empty"
	}

	// check for YearMonthFilter validity
	if validationErr := r.StartDate.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}
	if validationErr := r.EndDate.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	// check for range validity
	startTime := time.Date(r.StartDate.Year, time.Month(r.StartDate.Month), 0, 0, 0, 0, 0, util.DefaultTimezone)
	endTime := time.Date(r.EndDate.Year, time.Month(r.EndDate.Month), 0, 0, 0, 0, 0, util.DefaultTimezone)
	if startTime.After(endTime) {
		errorDetail["startDate"] = "startDate must be less than endDate"
		errorDetail["endDate"] = "endDate must be greater than startDate"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

func (r YearMonthRangeFilter) ToTimeFilter(filterType YearMonthFilterType) TimeFilter {
	return TimeFilter{
		StartDatetime: time.Date(r.StartDate.Year, time.Month(r.StartDate.Month), 0, 0, 0, 0, 0, util.DefaultTimezone),
		EndDatetime:   time.Date(r.EndDate.Year, time.Month(r.EndDate.Month), 0, 0, 0, 0, 0, util.DefaultTimezone),
	}
}
