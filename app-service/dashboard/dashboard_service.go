package dashboard

import (
	"context"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/util"
)

type MonthyExpense_GroupBy string
type MonthyIncome_GroupBy string

var (
	MonthyExpense_GroupBy_Teacher    MonthyExpense_GroupBy = "TEACHER"
	MonthyExpense_GroupBy_Instrument MonthyExpense_GroupBy = "INSTRUMENT"
)

var (
	MonthyIncome_GroupBy_Student    MonthyIncome_GroupBy = "STUDENT"
	MonthyIncome_GroupBy_Instrument MonthyIncome_GroupBy = "INSTRUMENT"
)

type OverviewResult struct {
	Data []OverviewResultItem `json:"data"`
}
type OverviewResultItem struct {
	Label      string  `json:"label"`
	Value      int64   `json:"value"`
	Percentage float32 `json:"percentage"`
}

type MonthlySummaryResult struct {
	Data []MonthlySummaryResultItem `json:"data"`
}
type MonthlySummaryResultItem struct {
	Label      string  `json:"label"`
	Value      int64   `json:"value"`
	Percentage float32 `json:"percentage"`
}

type DashboardService interface {
	GetExpenseOverview(ctx context.Context, spec GetExpenseOverviewSpec) (OverviewResult, error)
	GetExpenseMonthlySummary(ctx context.Context, spec GetExpenseMontlySummarySpec) (MonthlySummaryResult, error)

	GetIncomeOverview(ctx context.Context, spec GetIncomeOverviewSpec) (OverviewResult, error)
	GetIncomeMonthlySummary(ctx context.Context, spec GetIncomeMontlySummarySpec) (MonthlySummaryResult, error)

	GetNetIncomeOverview(ctx context.Context, spec GetNetIncomeOverviewSpec) (OverviewResult, error)

	GetTeacherPaymentDetails(ctx context.Context)
}

type GetExpenseOverviewSpec struct {
	util.TimeSpec
	TeacherIDs    []entity.TeacherID
	InstrumentIDs []entity.InstrumentID
}
type GetExpenseMontlySummarySpec struct {
	util.TimeSpec
	GroupBy       MonthyExpense_GroupBy
	TeacherIDs    []entity.TeacherID
	InstrumentIDs []entity.InstrumentID
}

type GetIncomeOverviewSpec struct {
	util.TimeSpec
	StudentIDs    []entity.StudentID
	InstrumentIDs []entity.InstrumentID
}
type GetIncomeMontlySummarySpec struct {
	util.TimeSpec
	GroupBy       MonthyIncome_GroupBy
	StudentIDs    []entity.StudentID
	InstrumentIDs []entity.InstrumentID
}

type GetNetIncomeOverviewSpec struct {
	util.TimeSpec
}
