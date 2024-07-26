package output

import (
	"sonamusica-backend/app-service/dashboard"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/errs"
)

// ============================== EXPENSE ==============================

type GetDashboardExpenseOverviewRequest struct {
	DateRange     YearMonthRangeFilter  `json:"dateRange"`
	TeacherIDs    []entity.TeacherID    `json:"teacherIds"`
	InstrumentIDs []entity.InstrumentID `json:"instrumentIds"`
}
type GetDashboardExpenseOverviewResponse struct {
	Data    GetDashboardExpenseOverviewResult `json:"data"`
	Message string                            `json:"message,omitempty"`
}

type GetDashboardExpenseOverviewResult struct {
	Results []dashboard.OverviewResultItem `json:"results"`
}

func (r GetDashboardExpenseOverviewRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if validationErr := r.DateRange.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetDashboardExpenseMonthlySummaryRequest struct {
	SelectedDate  YearMonthFilter                 `json:"selectedDate"`
	GroupBy       dashboard.MonthyExpense_GroupBy `json:"groupBy"`
	TeacherIDs    []entity.TeacherID              `json:"teacherIds"`
	InstrumentIDs []entity.InstrumentID           `json:"instrumentIds"`
}
type GetDashboardExpenseMonthlySummaryResponse struct {
	Data    GetDashboardExpenseMonthlySummaryResult `json:"data"`
	Message string                                  `json:"message,omitempty"`
}

type GetDashboardExpenseMonthlySummaryResult struct {
	Results []dashboard.MonthlySummaryResultItem `json:"results"`
}

func (r GetDashboardExpenseMonthlySummaryRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	// for dashboard monthly summary, it doesn't make sense to have empty selectedDate.
	// user should pick a specific month of a year
	if r.SelectedDate.Year == 0 || r.SelectedDate.Month == 0 {
		errorDetail["selectedDate.year"] = "selectedDate.year must not be empty"
		errorDetail["selectedDate.month"] = "selectedDate.month must not be empty"
	} else {
		if validationErr := r.SelectedDate.Validate(); validationErr != nil {
			for key, value := range validationErr.GetErrorDetail() {
				errorDetail[key] = value
			}
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

// ============================== INCOME ==============================

type GetDashboardIncomeOverviewRequest struct {
	DateRange     YearMonthRangeFilter  `json:"dateRange"`
	StudentIDs    []entity.StudentID    `json:"studentIDs"`
	InstrumentIDs []entity.InstrumentID `json:"instrumentIds"`
}
type GetDashboardIncomeOverviewResponse struct {
	Data    GetDashboardIncomeOverviewResult `json:"data"`
	Message string                           `json:"message,omitempty"`
}

type GetDashboardIncomeOverviewResult struct {
	Results []dashboard.OverviewResultItem `json:"results"`
}

func (r GetDashboardIncomeOverviewRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if validationErr := r.DateRange.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetDashboardIncomeMonthlySummaryRequest struct {
	SelectedDate  YearMonthFilter                `json:"selectedDate"`
	GroupBy       dashboard.MonthyIncome_GroupBy `json:"groupBy"`
	StudentIDs    []entity.StudentID             `json:"studentIDs"`
	InstrumentIDs []entity.InstrumentID          `json:"instrumentIds"`
}
type GetDashboardIncomeMonthlySummaryResponse struct {
	Data    GetDashboardIncomeMonthlySummaryResult `json:"data"`
	Message string                                 `json:"message,omitempty"`
}

type GetDashboardIncomeMonthlySummaryResult struct {
	Results []dashboard.MonthlySummaryResultItem `json:"results"`
}

func (r GetDashboardIncomeMonthlySummaryRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	// for dashboard monthly summary, it doesn't make sense to have empty selectedDate.
	// user should pick a specific month of a year
	if r.SelectedDate.Year == 0 || r.SelectedDate.Month == 0 {
		errorDetail["selectedDate.year"] = "selectedDate.year must not be empty"
		errorDetail["selectedDate.month"] = "selectedDate.month must not be empty"
	} else {
		if validationErr := r.SelectedDate.Validate(); validationErr != nil {
			for key, value := range validationErr.GetErrorDetail() {
				errorDetail[key] = value
			}
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}
