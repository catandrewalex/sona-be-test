package impl

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/dashboard"
	"sonamusica-backend/app-service/identity"
)

// ============================== EXPENSE ==============================

func NewOverviewResultItems_FromMySQLExpenseOverview(rows []mysql.GetExpenseOverviewRow) []dashboard.OverviewResultItem {
	resultItems := make([]dashboard.OverviewResultItem, 0, len(rows))
	for _, row := range rows {
		resultItems = append(resultItems, dashboard.OverviewResultItem{
			Label: row.YearWithMonth,
			Value: row.TotalPaidCourseFee + row.TotalPaidTransportFee,
		})
	}

	return resultItems
}

func NewMSResultItems_FromMySQLExpenseMSByTeacher(rows []mysql.GetExpenseMonthlySummaryGroupedByTeacherRow) []dashboard.MonthlySummaryResultItem {
	resultItems := make([]dashboard.MonthlySummaryResultItem, 0, len(rows))
	for _, row := range rows {
		userDetail := identity.UnmarshalUserDetail(row.UserDetail, mainLog)
		resultItems = append(resultItems, dashboard.MonthlySummaryResultItem{
			Label: userDetail.String(),
			Value: row.TotalPaidCourseFee + row.TotalPaidTransportFee,
		})
	}

	return resultItems
}

func NewMSResultItems_FromMySQLExpenseMSByInstrument(rows []mysql.GetExpenseMonthlySummaryGroupedByInstrumentRow) []dashboard.MonthlySummaryResultItem {
	resultItems := make([]dashboard.MonthlySummaryResultItem, 0, len(rows))
	for _, row := range rows {
		resultItems = append(resultItems, dashboard.MonthlySummaryResultItem{
			Label: row.Instrument.Name,
			Value: row.TotalPaidCourseFee + row.TotalPaidTransportFee,
		})
	}

	return resultItems
}

// ============================== INCOME ==============================

func NewOverviewResultItems_FromMySQLIncomeOverview(rows []mysql.GetIncomeOverviewRow) []dashboard.OverviewResultItem {
	resultItems := make([]dashboard.OverviewResultItem, 0, len(rows))
	for _, row := range rows {
		resultItems = append(resultItems, dashboard.OverviewResultItem{
			Label: row.YearWithMonth,
			Value: row.TotalCourseFee + row.TotalTransportFee + row.TotalPenaltyFeeValue,
		})
	}

	return resultItems
}

func NewMSResultItems_FromMySQLIncomeMSByTeacher(rows []mysql.GetIncomeMonthlySummaryGroupedByStudentRow) []dashboard.MonthlySummaryResultItem {
	resultItems := make([]dashboard.MonthlySummaryResultItem, 0, len(rows))
	for _, row := range rows {
		userDetail := identity.UnmarshalUserDetail(row.UserDetail, mainLog)
		resultItems = append(resultItems, dashboard.MonthlySummaryResultItem{
			Label: userDetail.String(),
			Value: row.TotalCourseFee + row.TotalTransportFee + row.TotalPenaltyFeeValue,
		})
	}

	return resultItems
}

func NewMSResultItems_FromMySQLIncomeMSByInstrument(rows []mysql.GetIncomeMonthlySummaryGroupedByInstrumentRow) []dashboard.MonthlySummaryResultItem {
	resultItems := make([]dashboard.MonthlySummaryResultItem, 0, len(rows))
	for _, row := range rows {
		resultItems = append(resultItems, dashboard.MonthlySummaryResultItem{
			Label: row.Instrument.Name,
			Value: row.TotalCourseFee + row.TotalTransportFee + row.TotalPenaltyFeeValue,
		})
	}

	return resultItems
}
