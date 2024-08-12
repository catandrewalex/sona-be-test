package impl

import (
	"context"
	"database/sql"
	"fmt"

	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/dashboard"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/config"
	"sonamusica-backend/logging"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("DashboardService", logging.GetLevel(configObject.LogLevel))
)

type dashboardServiceImpl struct {
	mySQLQueries *relational_db.MySQLQueries

	entityService entity.EntityService
}

var _ dashboard.DashboardService = (*dashboardServiceImpl)(nil)

func NewDashboardServiceImpl(mySQLQueries *relational_db.MySQLQueries, entityService entity.EntityService) *dashboardServiceImpl {
	return &dashboardServiceImpl{
		mySQLQueries:  mySQLQueries,
		entityService: entityService,
	}
}

func (s dashboardServiceImpl) GetExpenseOverview(ctx context.Context, spec dashboard.GetExpenseOverviewSpec) (dashboard.OverviewResult, error) {
	timeFilter := spec.TimeSpec
	err := timeFilter.ValidateZeroValues()
	if err != nil {
		return dashboard.OverviewResult{}, fmt.Errorf("ValidateZeroValues(): %v", err)
	}

	teacherIDs := make([]sql.NullInt64, 0)
	instrumentIDs := make([]int64, 0)
	for _, teacherID := range spec.TeacherIDs {
		teacherIDs = append(teacherIDs, sql.NullInt64{Int64: int64(teacherID), Valid: true})
	}
	for _, instrumentID := range spec.InstrumentIDs {
		instrumentIDs = append(instrumentIDs, int64(instrumentID))
	}

	useTeacherFilter := len(teacherIDs) > 0
	useInstrumentFilter := len(instrumentIDs) > 0

	expenseOverviewRows, err := s.mySQLQueries.GetExpenseOverview(ctx, mysql.GetExpenseOverviewParams{
		StartDate:           timeFilter.StartDatetime,
		EndDate:             timeFilter.EndDatetime,
		TeacherIds:          teacherIDs,
		UseTeacherFilter:    useTeacherFilter,
		InstrumentIds:       instrumentIDs,
		UseInstrumentFilter: useInstrumentFilter,
	})
	if err != nil {
		return dashboard.OverviewResult{}, fmt.Errorf("mySQLQueries.GetExpenseOverview(): %w", err)
	}

	overviewResultItems := NewOverviewResultItems_FromMySQLExpenseOverview(expenseOverviewRows)
	CalculateOverviewResultItemsPercentage(&overviewResultItems)

	return dashboard.OverviewResult{
		Data: overviewResultItems,
	}, nil
}

func (s dashboardServiceImpl) GetExpenseMonthlySummary(ctx context.Context, spec dashboard.GetExpenseMontlySummarySpec) (dashboard.MonthlySummaryResult, error) {
	timeFilter := spec.TimeSpec
	err := timeFilter.ValidateZeroValues()
	if err != nil {
		return dashboard.MonthlySummaryResult{}, fmt.Errorf("ValidateZeroValues(): %v", err)
	}

	teacherIDs := make([]sql.NullInt64, 0)
	instrumentIDs := make([]int64, 0)
	for _, teacherID := range spec.TeacherIDs {
		teacherIDs = append(teacherIDs, sql.NullInt64{Int64: int64(teacherID), Valid: true})
	}
	for _, instrumentID := range spec.InstrumentIDs {
		instrumentIDs = append(instrumentIDs, int64(instrumentID))
	}

	useTeacherFilter := len(teacherIDs) > 0
	useInstrumentFilter := len(teacherIDs) > 0

	var monthlySummaryResultItems []dashboard.MonthlySummaryResultItem
	switch spec.GroupBy {
	case dashboard.MonthyExpense_GroupBy_Teacher:
		monthlySummaryRows, err := s.mySQLQueries.GetExpenseMonthlySummaryGroupedByTeacher(ctx, mysql.GetExpenseMonthlySummaryGroupedByTeacherParams{
			StartDate:           timeFilter.StartDatetime,
			EndDate:             timeFilter.EndDatetime,
			TeacherIds:          teacherIDs,
			UseTeacherFilter:    useTeacherFilter,
			InstrumentIds:       instrumentIDs,
			UseInstrumentFilter: useInstrumentFilter,
		})
		if err != nil {
			return dashboard.MonthlySummaryResult{}, fmt.Errorf("mySQLQueries.GetExpenseMonthlySummaryGroupedByTeacher(): %w", err)
		}
		monthlySummaryResultItems = NewMSResultItems_FromMySQLExpenseMSByTeacher(monthlySummaryRows)

	case dashboard.MonthyExpense_GroupBy_Instrument:
		monthlySummaryRows, err := s.mySQLQueries.GetExpenseMonthlySummaryGroupedByInstrument(ctx, mysql.GetExpenseMonthlySummaryGroupedByInstrumentParams{
			StartDate:           timeFilter.StartDatetime,
			EndDate:             timeFilter.EndDatetime,
			TeacherIds:          teacherIDs,
			UseTeacherFilter:    useTeacherFilter,
			InstrumentIds:       instrumentIDs,
			UseInstrumentFilter: useInstrumentFilter,
		})
		if err != nil {
			return dashboard.MonthlySummaryResult{}, fmt.Errorf("mySQLQueries.GetExpenseMonthlySummaryGroupedByInstrument(): %w", err)
		}
		monthlySummaryResultItems = NewMSResultItems_FromMySQLExpenseMSByInstrument(monthlySummaryRows)

	default:
		return dashboard.MonthlySummaryResult{}, fmt.Errorf("invalid 'GroupBy' option: %s", spec.GroupBy)
	}

	CalculateMonthlySummaryResultItemsPercentage(&monthlySummaryResultItems)

	return dashboard.MonthlySummaryResult{
		Data: monthlySummaryResultItems,
	}, nil
}

func (s dashboardServiceImpl) GetIncomeOverview(ctx context.Context, spec dashboard.GetIncomeOverviewSpec) (dashboard.OverviewResult, error) {
	timeFilter := spec.TimeSpec
	err := timeFilter.ValidateZeroValues()
	if err != nil {
		return dashboard.OverviewResult{}, fmt.Errorf("ValidateZeroValues(): %v", err)
	}

	studentIDs := make([]int64, 0)
	instrumentIDs := make([]int64, 0)
	for _, studentID := range spec.StudentIDs {
		studentIDs = append(studentIDs, int64(studentID))
	}
	for _, instrumentID := range spec.InstrumentIDs {
		instrumentIDs = append(instrumentIDs, int64(instrumentID))
	}

	useStudentFilter := len(studentIDs) > 0
	useInstrumentFilter := len(studentIDs) > 0

	incomeOverviewRows, err := s.mySQLQueries.GetIncomeOverview(ctx, mysql.GetIncomeOverviewParams{
		StartDate:           timeFilter.StartDatetime,
		EndDate:             timeFilter.EndDatetime,
		StudentIds:          studentIDs,
		UseStudentFilter:    useStudentFilter,
		InstrumentIds:       instrumentIDs,
		UseInstrumentFilter: useInstrumentFilter,
	})
	if err != nil {
		return dashboard.OverviewResult{}, fmt.Errorf("mySQLQueries.GetIncomeOverview(): %w", err)
	}

	overviewResultItems := NewOverviewResultItems_FromMySQLIncomeOverview(incomeOverviewRows)
	CalculateOverviewResultItemsPercentage(&overviewResultItems)

	return dashboard.OverviewResult{
		Data: overviewResultItems,
	}, nil
}

func (s dashboardServiceImpl) GetIncomeMonthlySummary(ctx context.Context, spec dashboard.GetIncomeMontlySummarySpec) (dashboard.MonthlySummaryResult, error) {
	timeFilter := spec.TimeSpec
	err := timeFilter.ValidateZeroValues()
	if err != nil {
		return dashboard.MonthlySummaryResult{}, fmt.Errorf("ValidateZeroValues(): %v", err)
	}

	studentIDs := make([]int64, 0)
	instrumentIDs := make([]int64, 0)
	for _, studentID := range spec.StudentIDs {
		studentIDs = append(studentIDs, int64(studentID))
	}
	for _, instrumentID := range spec.InstrumentIDs {
		instrumentIDs = append(instrumentIDs, int64(instrumentID))
	}

	useTeacherFilter := len(studentIDs) > 0
	useInstrumentFilter := len(studentIDs) > 0

	var monthlySummaryResultItems []dashboard.MonthlySummaryResultItem
	switch spec.GroupBy {
	case dashboard.MonthyIncome_GroupBy_Student:
		monthlySummaryRows, err := s.mySQLQueries.GetIncomeMonthlySummaryGroupedByStudent(ctx, mysql.GetIncomeMonthlySummaryGroupedByStudentParams{
			StartDate:           timeFilter.StartDatetime,
			EndDate:             timeFilter.EndDatetime,
			StudentIds:          studentIDs,
			UseStudentFilter:    useTeacherFilter,
			InstrumentIds:       instrumentIDs,
			UseInstrumentFilter: useInstrumentFilter,
		})
		if err != nil {
			return dashboard.MonthlySummaryResult{}, fmt.Errorf("mySQLQueries.GetIncomeMonthlySummaryGroupedByTeacher(): %w", err)
		}
		monthlySummaryResultItems = NewMSResultItems_FromMySQLIncomeMSByTeacher(monthlySummaryRows)

	case dashboard.MonthyIncome_GroupBy_Instrument:
		monthlySummaryRows, err := s.mySQLQueries.GetIncomeMonthlySummaryGroupedByInstrument(ctx, mysql.GetIncomeMonthlySummaryGroupedByInstrumentParams{
			StartDate:           timeFilter.StartDatetime,
			EndDate:             timeFilter.EndDatetime,
			StudentIds:          studentIDs,
			UseStudentFilter:    useTeacherFilter,
			InstrumentIds:       instrumentIDs,
			UseInstrumentFilter: useInstrumentFilter,
		})
		if err != nil {
			return dashboard.MonthlySummaryResult{}, fmt.Errorf("mySQLQueries.GetIncomeMonthlySummaryGroupedByInstrument(): %w", err)
		}
		monthlySummaryResultItems = NewMSResultItems_FromMySQLIncomeMSByInstrument(monthlySummaryRows)

	default:
		return dashboard.MonthlySummaryResult{}, fmt.Errorf("invalid 'GroupBy' option: %s", spec.GroupBy)
	}

	CalculateMonthlySummaryResultItemsPercentage(&monthlySummaryResultItems)

	return dashboard.MonthlySummaryResult{
		Data: monthlySummaryResultItems,
	}, nil
}

func (s dashboardServiceImpl) GetNetIncomeOverview(ctx context.Context, spec dashboard.GetNetIncomeOverviewSpec) (dashboard.OverviewResult, error) {
	return dashboard.OverviewResult{}, nil
}

func (s dashboardServiceImpl) GetTeacherPaymentDetails(ctx context.Context) {

}
