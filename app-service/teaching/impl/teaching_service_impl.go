package teaching

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/app-service/util"
	"sonamusica-backend/config"
	"sonamusica-backend/errs"
	"sonamusica-backend/logging"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("TeachingService", logging.GetLevel(configObject.LogLevel))
)

const (
	pagination_FirstPage = 1
	pagination_FetchAll  = 10000
)

type teachingServiceImpl struct {
	mySQLQueries *relational_db.MySQLQueries

	entityService entity.EntityService
}

var _ teaching.TeachingService = (*teachingServiceImpl)(nil)

func NewTeachingServiceImpl(mySQLQueries *relational_db.MySQLQueries, entityService entity.EntityService) *teachingServiceImpl {
	return &teachingServiceImpl{
		mySQLQueries:  mySQLQueries,
		entityService: entityService,
	}
}

func (s teachingServiceImpl) SearchEnrollmentPayment(ctx context.Context, timeFilter util.TimeSpec) ([]entity.EnrollmentPayment, error) {
	paginationSpec := util.PaginationSpec{
		Page:           pagination_FirstPage,
		ResultsPerPage: pagination_FetchAll,
	}
	getEnrollmentPaymentsResult, err := s.entityService.GetEnrollmentPayments(ctx, paginationSpec, timeFilter, true)
	if err != nil {
		return []entity.EnrollmentPayment{}, fmt.Errorf("entityService.GetEnrollmentPayments(): %v", err)
	}

	return getEnrollmentPaymentsResult.EnrollmentPayments, nil
}

func (s teachingServiceImpl) GetEnrollmentPaymentInvoice(ctx context.Context, studentEnrollmentID entity.StudentEnrollmentID) (teaching.StudentEnrollmentInvoice, error) {
	studentEnrollment, err := s.entityService.GetStudentEnrollmentById(ctx, studentEnrollmentID)
	if err != nil {
		return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("entityService.GetStudentEnrollmentById(): %w", err)
	}

	// calculate Course Fee
	courseFeeValue := studentEnrollment.ClassInfo.Course.DefaultFee
	teacherID, err := s.mySQLQueries.GetClassTeacherId(ctx, int64(studentEnrollment.ClassInfo.ClassID))
	if err != nil {
		return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("mySQLQueries.GetClassTeacherId(): %w", err)
	}
	if teacherID.Valid {
		teacherSpecialFee, err := s.mySQLQueries.GetTeacherSpecialFeesByTeacherIdAndCourseId(ctx, mysql.GetTeacherSpecialFeesByTeacherIdAndCourseIdParams{
			TeacherID: teacherID.Int64,
			CourseID:  int64(studentEnrollment.ClassInfo.Course.CourseID),
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) { // ignore not found error, and use the default course value
			return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("mySQLQueries.GetTeacherSpecialFeesByTeacherIdAndCourseId(): %w", err)
		}
		if teacherSpecialFee.ID > 0 {
			courseFeeValue = teacherSpecialFee.Fee
		}
	}

	// calculate Course Fee Penalty (e.g. due to late payment)
	latestPaymentDate, err := s.mySQLQueries.GetLatestEnrollmentPaymentDateByStudentEnrollmentId(ctx, sql.NullInt64{Int64: int64(studentEnrollmentID), Valid: true})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("mySQLQueries.GetLatestEnrollmentPaymentDateByStudentEnrollmentId(): %w", err)
		}
	}

	var lastPaymentDate *time.Time = nil
	lastDateBeforePenalty := util.MaxDateTime
	if latestPaymentDate != nil {
		temp := latestPaymentDate.(time.Time)
		lastPaymentDate = &temp
		lastDateBeforePenalty = time.Date(lastPaymentDate.Year(), lastPaymentDate.AddDate(0, 1, 0).Month(), 10, 0, 0, 0, 0, util.DefaultTimezone)
	}
	daysLate := int32(time.Since(lastDateBeforePenalty).Hours() / 24)
	var penaltyFeeValue int32 = 0
	if daysLate > 0 {
		penaltyFeeValue = teaching.Default_PenaltyFeeValue * daysLate
	}

	// calculate transport fee (splitted unionly across all class students)
	splittedTransportFee := studentEnrollment.ClassInfo.TransportFee
	classIdToTotalStudents, err := s.mySQLQueries.GetClassesTotalStudentsByClassIds(ctx, []int64{int64(studentEnrollment.ClassInfo.ClassID)})
	if err != nil {
		return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("mySQLQueries.GetClassesTotalStudentsByClassIds(): %w", err)
	}
	if len(classIdToTotalStudents) > 0 && classIdToTotalStudents[0].TotalStudents > 1 {
		splittedTransportFee /= int32(classIdToTotalStudents[0].TotalStudents)
	}

	return teaching.StudentEnrollmentInvoice{
		BalanceTopUp:      teaching.Default_BalanceTopUp,
		PenaltyFeeValue:   penaltyFeeValue,
		CourseFeeValue:    courseFeeValue,
		TransportFeeValue: splittedTransportFee,
		LastPaymentDate:   lastPaymentDate,
		DaysLate:          daysLate,
	}, nil
}

func (s teachingServiceImpl) SubmitEnrollmentPayment(ctx context.Context, spec teaching.SubmitStudentEnrollmentPaymentSpec) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		_, err := s.entityService.InsertEnrollmentPayments(newCtx, []entity.InsertEnrollmentPaymentSpec{
			{
				StudentEnrollmentID: spec.StudentEnrollmentID,
				PaymentDate:         spec.PaymentDate,
				BalanceTopUp:        spec.BalanceTopUp,
				CourseFeeValue:      spec.CourseFeeValue,
				TransportFeeValue:   spec.TransportFeeValue,
				PenaltyFeeValue:     spec.PenaltyFeeValue,
			},
		})
		if err != nil {
			return fmt.Errorf("entityService.InsertEnrollmentPayments(): %w", err)
		}

		// sum all negative quotas to reduce balanceTUpValue, and reset those SLTs with negative quota to 0
		var balanceTopUpMinusPenalty = float64(spec.BalanceTopUp)
		slts, err := qtx.GetSLTWithNegativeQuotaByEnrollmentId(newCtx, int64(spec.StudentEnrollmentID))
		if err != nil {
			return fmt.Errorf("qtx.GetSLTWithNegativeQuotaByEnrollmentId(): %w", err)
		}
		negativeQuotaSLTIDs := make([]int64, 0, len(slts))
		for _, slt := range slts {
			if slt.Quota >= 0 {
				continue
			}
			balanceTopUpMinusPenalty += slt.Quota
			negativeQuotaSLTIDs = append(negativeQuotaSLTIDs, slt.ID)
		}
		// NOTE: we actually can combine these: (1) sum all negative quotas, and (2) reset the quota to 0 into a single SQL method.
		//  But, for the sake of better control, I decided to do this separately, with the cost of more DB I/O.
		err = qtx.ResetStudentLearningTokenQuotaByIds(newCtx, negativeQuotaSLTIDs)
		if err != nil {
			return fmt.Errorf("qtx.ResetStudentLearningTokenQuotaByIds(): %w", err)
		}

		// Upsert StudentLearningTokens
		existingSLT, err := qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(newCtx, mysql.GetSLTByEnrollmentIdAndCourseFeeAndTransportFeeParams{
			EnrollmentID:      int64(spec.StudentEnrollmentID),
			CourseFeeValue:    spec.CourseFeeValue,
			TransportFeeValue: spec.TransportFeeValue,
		})
		isNeedInsert := errors.Is(err, sql.ErrNoRows)
		if isNeedInsert {
			err = nil
		}
		if err != nil {
			return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(): %w", err)
		}

		if isNeedInsert {
			_, err = s.entityService.InsertStudentLearningTokens(newCtx, []entity.InsertStudentLearningTokenSpec{
				{
					StudentEnrollmentID: spec.StudentEnrollmentID,
					Quota:               balanceTopUpMinusPenalty,
					CourseFeeValue:      spec.CourseFeeValue,
					TransportFeeValue:   spec.TransportFeeValue,
				},
			})
			if err != nil {
				return fmt.Errorf("entityService.InsertStudentLearningTokens(): %w", err)
			}
		} else {
			err := qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota: balanceTopUpMinusPenalty,
				ID:    existingSLT.ID,
			})
			if err != nil {
				return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) EditEnrollmentPayment(ctx context.Context, spec teaching.EditStudentEnrollmentPaymentSpec) (entity.EnrollmentPaymentID, error) {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		prevEP, err := qtx.GetEnrollmentPaymentById(newCtx, int64(spec.EnrollmentPaymentID))
		if err != nil {
			return fmt.Errorf("qtx.GetEnrollmentPaymentById(): %w", err)
		}

		updatedSLT, err := qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(newCtx, mysql.GetSLTByEnrollmentIdAndCourseFeeAndTransportFeeParams{
			EnrollmentID:      prevEP.StudentEnrollmentID,
			CourseFeeValue:    prevEP.CourseFeeValue,
			TransportFeeValue: prevEP.TransportFeeValue,
		})
		skipSLTUpdate := errors.Is(err, sql.ErrNoRows)
		if skipSLTUpdate {
			mainLog.Warn("EnrollmentPayment with ID='%d' doesn't have studentLearningToken, check for bad data possibility. Skipping to update studentLearningToken.", prevEP.EnrollmentPaymentID)
			err = nil
		}
		if err != nil {
			return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(): %w", err)
		}

		if !skipSLTUpdate {
			quotaChange := float64(spec.BalanceTopUp - prevEP.BalanceTopUp)
			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota: quotaChange,
				ID:    updatedSLT.ID,
			})
			if err != nil {
				return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
			}
		}

		err = qtx.UpdateEnrollmentPaymentDateAndBalance(newCtx, mysql.UpdateEnrollmentPaymentDateAndBalanceParams{
			PaymentDate:  spec.PaymentDate,
			BalanceTopUp: spec.BalanceTopUp,
			ID:           int64(spec.EnrollmentPaymentID),
		})
		if err != nil {
			return fmt.Errorf("entityService.UpdateEnrollmentPayment(): %w", err)
		}

		return nil
	})
	if err != nil {
		return spec.EnrollmentPaymentID, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return spec.EnrollmentPaymentID, nil
}

func (s teachingServiceImpl) RemoveEnrollmentPayment(ctx context.Context, enrollmentPaymentID entity.EnrollmentPaymentID) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		prevEP, err := qtx.GetEnrollmentPaymentById(newCtx, int64(enrollmentPaymentID))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				mainLog.Warn("EnrollmentPayment with ID='%d' doesn't exist. Skipping to delete the entity.", prevEP.EnrollmentPaymentID)
				// we don't return not found error as this is a deletion method --> missing entity which has the requested ID is ok.
				return nil
			}
			return fmt.Errorf("qtx.GetEnrollmentPaymentById(): %w", err)
		}

		updatedSLT, err := qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(newCtx, mysql.GetSLTByEnrollmentIdAndCourseFeeAndTransportFeeParams{
			EnrollmentID:      prevEP.StudentEnrollmentID,
			CourseFeeValue:    prevEP.CourseFeeValue,
			TransportFeeValue: prevEP.TransportFeeValue,
		})
		skipSLTUpdate := errors.Is(err, sql.ErrNoRows)
		if skipSLTUpdate {
			mainLog.Warn("EnrollmentPayment with ID='%d' doesn't have studentLearningToken, check for bad data possibility. Skipping to update studentLearningToken.", prevEP.EnrollmentPaymentID)
			err = nil
		}
		if err != nil {
			return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(): %w", err)
		}

		if !skipSLTUpdate {
			quotaChange := -1 * prevEP.BalanceTopUp
			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota: float64(quotaChange),
				ID:    updatedSLT.ID,
			})
			if err != nil {
				return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
			}
		}

		err = s.entityService.DeleteEnrollmentPayments(newCtx, []entity.EnrollmentPaymentID{
			entity.EnrollmentPaymentID(prevEP.EnrollmentPaymentID),
		})
		if err != nil {
			return fmt.Errorf("entityService.DeleteEnrollmentPayments(): %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) SearchClass(ctx context.Context, spec teaching.SearchClassSpec) ([]entity.Class, error) {
	paginationSpec := util.PaginationSpec{
		Page:           pagination_FirstPage,
		ResultsPerPage: pagination_FetchAll,
	}
	getClassesSpec := entity.GetClassesSpec{
		IncludeDeactivated: true,
		TeacherID:          spec.TeacherID,
		StudentID:          spec.StudentID,
		CourseID:           spec.CourseID,
	}
	getClassResult, err := s.entityService.GetClasses(ctx, paginationSpec, getClassesSpec)
	if err != nil {
		return []entity.Class{}, fmt.Errorf("entityService.GetClasses(): %v", err)
	}

	return getClassResult.Classes, nil
}

func (s teachingServiceImpl) GetAttendancesByClassID(ctx context.Context, spec teaching.GetAttendancesByClassIDSpec) (teaching.GetAttendancesByClassIDResult, error) {
	getAttendancesSpec := entity.GetAttendancesSpec{
		ClassID:   spec.ClassID,
		StudentID: spec.StudentID,
		TimeSpec:  spec.TimeSpec,
	}
	getAttendancesResult, err := s.entityService.GetAttendances(ctx, spec.PaginationSpec, getAttendancesSpec)
	if err != nil {
		return teaching.GetAttendancesByClassIDResult{}, fmt.Errorf("entityService.GetAttendances(): %v", err)
	}

	return teaching.GetAttendancesByClassIDResult{
		Attendances:      getAttendancesResult.Attendances,
		PaginationResult: getAttendancesResult.PaginationResult,
	}, nil
}

func (s teachingServiceImpl) AddAttendance(ctx context.Context, spec teaching.AddAttendanceSpec, autoCreateSLT bool) ([]entity.AttendanceID, error) {
	attendanceIDs := make([]entity.AttendanceID, 0)

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		studentEnrollments, err := qtx.GetStudentEnrollmentsByClassId(newCtx, int64(spec.ClassID))
		if err != nil {
			return fmt.Errorf("qtx.GetStudentEnrollmentsByClassId(): %w", err)
		}
		if len(studentEnrollments) == 0 {
			return fmt.Errorf("classID='%d': %w", spec.ClassID, errs.ErrClassHaveNoStudent)
		}

		studentEnrollmentIDsInt64 := make([]int64, 0, len(studentEnrollments))
		for _, studentEnrollment := range studentEnrollments {
			studentEnrollmentIDsInt64 = append(studentEnrollmentIDsInt64, studentEnrollment.StudentEnrollmentID)
		}

		enrollmentIDToEarliestSLTID := make(map[int64]entity.StudentLearningTokenID, 0)
		// students may have > 1 SLT, we'll pick the one with earliest non-zero quota.
		//   if all <= 0, we decrement the last SLT (thus becoming negative).
		earliestAvailableSLTs, err := qtx.GetEarliestAvailableSLTsByStudentEnrollmentIds(newCtx, studentEnrollmentIDsInt64)
		if err != nil {
			return fmt.Errorf("qtx.GetEarliestAvailableSLTsByStudentEnrollmentIds(): %w", err)
		}

		for _, earliestAvailableSLT := range earliestAvailableSLTs {
			enrollmentIDToEarliestSLTID[earliestAvailableSLT.EnrollmentID] = entity.StudentLearningTokenID(earliestAvailableSLT.StudentLearningTokenID)

			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota: spec.UsedStudentTokenQuota * -1,
				ID:    earliestAvailableSLT.StudentLearningTokenID,
			})
			if err != nil {
				return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
			}
		}

		// Insert attendances
		specs := make([]entity.InsertAttendanceSpec, 0, len(studentEnrollments))
		for _, studentEnrollment := range studentEnrollments {
			sltID, ok := enrollmentIDToEarliestSLTID[studentEnrollment.StudentEnrollmentID]
			if !ok {
				if !autoCreateSLT {
					return fmt.Errorf("studentEnrollment='%d': %w", studentEnrollment.StudentEnrollmentID, errs.ErrStudentEnrollmentHaveNoLearningToken)
				}

				mainLog.Warn("studentEnrollment='%d' doesn't have any studentLearningToken (SLT). Creating a new negative quota SLT as 'autoCreateSLT' is true.", studentEnrollment.StudentEnrollmentID)
				newSLTID, err := s.autoRegisterSLT(newCtx, entity.StudentEnrollmentID(studentEnrollment.StudentEnrollmentID), spec.UsedStudentTokenQuota*-1)
				if err != nil {
					return fmt.Errorf("autoRegisterSLT(): %w", err)
				}
				sltID = newSLTID
			}

			specs = append(specs, entity.InsertAttendanceSpec{
				ClassID:                spec.ClassID,
				TeacherID:              spec.TeacherID,
				StudentID:              entity.StudentID(studentEnrollment.StudentID),
				StudentLearningTokenID: sltID,
				Date:                   spec.Date,
				UsedStudentTokenQuota:  spec.UsedStudentTokenQuota,
				Duration:               spec.Duration,
				Note:                   spec.Note,
			})
		}

		attendanceIDs, err = s.entityService.InsertAttendances(newCtx, specs)
		if err != nil {
			return fmt.Errorf("entityService.InsertAttendances(): %w", err)
		}

		return nil
	})
	if err != nil {
		return []entity.AttendanceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return attendanceIDs, nil
}

func (s teachingServiceImpl) autoRegisterSLT(ctx context.Context, studentEnrollmentID entity.StudentEnrollmentID, quota float64) (entity.StudentLearningTokenID, error) {
	enrollmentID := entity.StudentEnrollmentID(studentEnrollmentID)
	invoice, err := s.GetEnrollmentPaymentInvoice(ctx, enrollmentID)
	if err != nil {
		return entity.StudentLearningTokenID_None, fmt.Errorf("GetEnrollmentPaymentInvoice(): %w", err)
	}

	newSLTIDs, err := s.entityService.InsertStudentLearningTokens(ctx, []entity.InsertStudentLearningTokenSpec{
		{
			StudentEnrollmentID: studentEnrollmentID,
			Quota:               quota,
			CourseFeeValue:      invoice.CourseFeeValue,
			TransportFeeValue:   invoice.TransportFeeValue,
		},
	})
	if err != nil {
		return entity.StudentLearningTokenID_None, fmt.Errorf("entityService.InsertStudentLearningTokens(): %w", err)
	}

	return newSLTIDs[0], nil
}

func (s teachingServiceImpl) EditAttendance(ctx context.Context, spec teaching.EditAttendanceSpec) ([]entity.AttendanceID, error) {
	errV := util.ValidateUpdateSpecs(ctx, []teaching.EditAttendanceSpec{spec}, s.mySQLQueries.CountAttendancesByIds)
	if errV != nil {
		return []entity.AttendanceID{}, errV
	}

	attendanceIDs := make([]entity.AttendanceID, 0)

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		rowResults, err := qtx.GetAttendanceIdsOfSameClassAndDate(newCtx, int64(spec.AttendanceID))
		if err != nil {
			return fmt.Errorf("qtx.GetAttendanceIdsOfSameClassAndDate(): %w", err)
		}

		attendanceIDsInt := make([]int64, 0, len(rowResults))
		sltIDIntToUsedQuota := make(map[int64]float64, len(rowResults))
		for _, rowResult := range rowResults {
			if util.Int32ToBool(rowResult.IsPaid) {
				return errs.ErrModifyingPaidAttendance
			}
			attendanceIDs = append(attendanceIDs, entity.AttendanceID(rowResult.ID))
			attendanceIDsInt = append(attendanceIDsInt, rowResult.ID)

			sltIDIntToUsedQuota[rowResult.TokenID] = rowResult.UsedStudentTokenQuota
		}

		err = qtx.EditAttendances(newCtx, mysql.EditAttendancesParams{
			TeacherID:             int64(spec.TeacherID),
			Date:                  spec.Date,
			UsedStudentTokenQuota: spec.UsedStudentTokenQuota,
			Duration:              spec.Duration,
			Note:                  spec.Note,
			Ids:                   attendanceIDsInt,
		})
		if err != nil {
			return fmt.Errorf("qtx.EditAttendances(): %w", err)
		}

		for sltIDInt, usedQuota := range sltIDIntToUsedQuota {
			quotaChange := float64(usedQuota - spec.UsedStudentTokenQuota)
			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota: quotaChange,
				ID:    sltIDInt,
			})
			if err != nil {
				return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return []entity.AttendanceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return attendanceIDs, nil
}

func (s teachingServiceImpl) RemoveAttendance(ctx context.Context, attendanceID entity.AttendanceID) ([]entity.AttendanceID, error) {
	deletedAttendanceIDs := make([]entity.AttendanceID, 0)

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		rowResults, err := qtx.GetAttendanceIdsOfSameClassAndDate(newCtx, int64(attendanceID))
		if err != nil {
			return fmt.Errorf("qtx.GetAttendanceIdsOfSameClassAndDate(): %w", err)
		}

		attendanceIDsInt := make([]int64, 0, len(rowResults))
		sltIDIntToUsedQuota := make(map[int64]float64, len(rowResults))
		for _, rowResult := range rowResults {
			if util.Int32ToBool(rowResult.IsPaid) {
				return errs.ErrModifyingPaidAttendance
			}
			deletedAttendanceIDs = append(deletedAttendanceIDs, entity.AttendanceID(rowResult.ID))
			attendanceIDsInt = append(attendanceIDsInt, rowResult.ID)

			sltIDIntToUsedQuota[rowResult.TokenID] = rowResult.UsedStudentTokenQuota
		}

		err = qtx.DeleteAttendancesByIds(newCtx, attendanceIDsInt)
		if err != nil {
			return fmt.Errorf("qtx.DeleteAttendancesByIds(): %w", err)
		}

		for sltIDInt, usedQuota := range sltIDIntToUsedQuota {
			quotaChange := usedQuota
			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota: quotaChange,
				ID:    sltIDInt,
			})
			if err != nil {
				return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return []entity.AttendanceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return deletedAttendanceIDs, nil
}

func (s teachingServiceImpl) GetUnpaidTeachers(ctx context.Context, spec teaching.GetUnpaidTeachersSpec) (teaching.GetUnpaidTeachersResult, error) {
	spec.Pagination.SetDefaultOnInvalidValues()
	limit, offset := spec.Pagination.GetLimitAndOffset()

	spec.TimeSpec.SetDefaultForZeroValues()

	teacherRows, err := s.mySQLQueries.GetUnpaidTeachers(ctx, mysql.GetUnpaidTeachersParams{
		StartDate: spec.StartDatetime,
		EndDate:   spec.EndDatetime,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return teaching.GetUnpaidTeachersResult{}, fmt.Errorf("mySQLQueries.GetUnpaidTeachers(): %w", err)
	}

	unpaidTeachers := NewUnpaidTeachersFromGetUnpaidTeachersRow(teacherRows)

	totalResults, err := s.mySQLQueries.CountUnpaidTeachers(ctx, mysql.CountUnpaidTeachersParams{
		StartDate: spec.StartDatetime,
		EndDate:   spec.EndDatetime,
	})
	if err != nil {
		return teaching.GetUnpaidTeachersResult{}, fmt.Errorf("mySQLQueries.CountUnpaidTeachers(): %w", err)
	}

	return teaching.GetUnpaidTeachersResult{
		UnpaidTeachers:   unpaidTeachers,
		PaginationResult: *util.NewPaginationResult(int(totalResults), spec.Pagination.ResultsPerPage, spec.Pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetTeacherPaymentInvoiceItems(ctx context.Context, spec teaching.GetTeacherPaymentInvoiceItemsSpec) ([]teaching.TeacherPaymentInvoiceItem, error) {
	getAttendancesByTeacherIdSpec := entity.GetAttendancesByTeacherIdSpec{
		TeacherID: spec.TeacherID,
		TimeSpec:  spec.TimeSpec,
	}

	attendances, err := s.entityService.GetAttendancesByTeacherId(ctx, getAttendancesByTeacherIdSpec)
	if err != nil {
		return []teaching.TeacherPaymentInvoiceItem{}, fmt.Errorf("entityService.GetAttendancesByTeacherId(): %v", err)
	}
	// sort the attendances ascendingly by date
	sort.SliceStable(attendances, func(i, j int) bool {
		return attendances[i].Date.Before(attendances[j].Date)
	})

	// group the Attendances by StudentLearningToken
	sltIdToAttendances := make(map[entity.StudentLearningTokenID][]teaching.AttendanceForInvoiceItem)
	sltIdToSLT := make(map[entity.StudentLearningTokenID]entity.StudentLearningToken_Minimal)
	sltIdToStudent := make(map[entity.StudentLearningTokenID]entity.StudentInfo_Minimal)
	studentIdToStudent := make(map[entity.StudentID]entity.StudentInfo_Minimal)
	studentIdToClass := make(map[entity.StudentID]entity.ClassInfo_Minimal)
	for _, attendance := range attendances {
		attendancesForInvoiceItem := teaching.AttendanceForInvoiceItem{
			AttendanceInfo_Minimal: entity.AttendanceInfo_Minimal{
				AttendanceID:          attendance.AttendanceID,
				TeacherInfo:           attendance.TeacherInfo,
				Date:                  attendance.Date,
				UsedStudentTokenQuota: attendance.UsedStudentTokenQuota,
				Duration:              attendance.Duration,
				Note:                  attendance.Note,
				IsPaid:                attendance.IsPaid,
			},
			GrossCourseFeeValue:           int32(float64(attendance.StudentLearningToken.CourseFeeValue) * attendance.UsedStudentTokenQuota / float64(teaching.Default_OneCourseCycle)),
			GrossTransportFeeValue:        int32(float64(attendance.StudentLearningToken.TransportFeeValue) * attendance.UsedStudentTokenQuota / float64(teaching.Default_OneCourseCycle)),
			CourseFeeSharingPercentage:    teaching.Default_CourseFeeSharingPercentage,
			TransportFeeSharingPercentage: teaching.Default_TransportFeeSharingPercentage,
		}

		sltId := attendance.StudentLearningToken.StudentLearningTokenID
		prevValues, ok := sltIdToAttendances[sltId]
		if ok {
			sltIdToAttendances[sltId] = append(prevValues, attendancesForInvoiceItem)
		} else {
			sltIdToAttendances[sltId] = []teaching.AttendanceForInvoiceItem{attendancesForInvoiceItem}
		}
		sltIdToSLT[sltId] = attendance.StudentLearningToken
		sltIdToStudent[sltId] = attendance.StudentInfo
		studentIdToStudent[attendance.StudentInfo.StudentID] = attendance.StudentInfo
		studentIdToClass[attendance.StudentInfo.StudentID] = attendance.ClassInfo
	}

	// then group the StudentLearningTokens by Student
	studentIdToSLTs := make(map[entity.StudentID][]teaching.Attendance_SLT)
	for sltId, attendancesForInvoiceItem := range sltIdToAttendances {
		slt := sltIdToSLT[sltId]
		attendance_SLT := teaching.Attendance_SLT{
			StudentLearningToken_Minimal: slt,
			Attendances:                  attendancesForInvoiceItem,
		}

		student := sltIdToStudent[sltId]
		prevValues, ok := studentIdToSLTs[student.StudentID]
		if ok {
			studentIdToSLTs[student.StudentID] = append(prevValues, attendance_SLT)
		} else {
			studentIdToSLTs[student.StudentID] = []teaching.Attendance_SLT{attendance_SLT}
		}
	}

	// then group the Students by Class, with SLTs per Student are sorted by ID (the SLT ID is automatically generated, so sort by ID == sort by date)
	classIdToStudents := make(map[entity.ClassID][]teaching.Attendance_Student)
	classIdToClass := make(map[entity.ClassID]entity.ClassInfo_Minimal)
	for studentId, attendance_SLTs := range studentIdToSLTs {
		sort.SliceStable(attendance_SLTs, func(i, j int) bool {
			return attendance_SLTs[i].StudentLearningTokenID < attendance_SLTs[j].StudentLearningTokenID
		})

		student := studentIdToStudent[studentId]
		class := studentIdToClass[studentId]
		attendence_Student := teaching.Attendance_Student{
			StudentInfo_Minimal:   student,
			StudentLearningTokens: attendance_SLTs,
		}

		prevValues, ok := classIdToStudents[class.ClassID]
		if ok {
			classIdToStudents[class.ClassID] = append(prevValues, attendence_Student)
		} else {
			classIdToStudents[class.ClassID] = []teaching.Attendance_Student{attendence_Student}
		}
		classIdToClass[class.ClassID] = class
	}

	// construct the TeacherPaymentInvoiceItem, with Students per Class are sorted by Student name
	teacherPaymentInvoiceItems := make([]teaching.TeacherPaymentInvoiceItem, 0, 1)
	for classId, attendance_Students := range classIdToStudents {
		sort.SliceStable(attendance_Students, func(i, j int) bool {
			return attendance_Students[i].StudentInfo_Minimal.String() < attendance_Students[j].StudentInfo_Minimal.String()
		})

		class := classIdToClass[classId]
		teacherPaymentInvoiceItems = append(teacherPaymentInvoiceItems, teaching.TeacherPaymentInvoiceItem{
			ClassInfo_Minimal: class,
			Students:          attendance_Students,
		})
	}
	// finally, sort the TeacherPaymentInvoiceItem by Class name
	sort.SliceStable(teacherPaymentInvoiceItems, func(i, j int) bool {
		return teacherPaymentInvoiceItems[i].ClassInfo_Minimal.String() < teacherPaymentInvoiceItems[j].ClassInfo_Minimal.String()
	})

	return teacherPaymentInvoiceItems, nil
}

func (s teachingServiceImpl) SubmitTeacherPayments(ctx context.Context, specs []teaching.SubmitTeacherPaymentsSpec) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		insertSpecs := make([]entity.InsertTeacherPaymentSpec, 0, len(specs))
		affectedAttendancedIDsInt64 := make([]int64, 0, len(specs))
		for _, spec := range specs {
			insertSpecs = append(insertSpecs, entity.InsertTeacherPaymentSpec{
				AttendanceID:          spec.AttendanceID,
				PaidCourseFeeValue:    spec.PaidCourseFeeValue,
				PaidTransportFeeValue: spec.PaidTransportFeeValue,
			})
			affectedAttendancedIDsInt64 = append(affectedAttendancedIDsInt64, int64(spec.AttendanceID))
		}

		_, err := s.entityService.InsertTeacherPayments(newCtx, insertSpecs)
		if err != nil {
			return fmt.Errorf("entityService.InsertTeacherPayments(): %w", err)
		}

		err = qtx.SetAttendancesIsPaidStatusByIds(newCtx, mysql.SetAttendancesIsPaidStatusByIdsParams{
			IsPaid: 1,
			Ids:    affectedAttendancedIDsInt64,
		})
		if err != nil {
			return fmt.Errorf("qtx.SetAttendancesIsPaidStatusByIds(): %v", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) EditTeacherPayments(ctx context.Context, specs []teaching.EditTeacherPaymentsSpec) ([]entity.TeacherPaymentID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountTeacherPaymentsByIds)
	if errV != nil {
		return []entity.TeacherPaymentID{}, errV
	}

	teacherPaymentIDs := make([]entity.TeacherPaymentID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			teacherPaymentIDs = append(teacherPaymentIDs, spec.TeacherPaymentID)
			err := qtx.EditTeacherPayment(newCtx, mysql.EditTeacherPaymentParams{
				PaidCourseFeeValue:    spec.PaidCourseFeeValue,
				PaidTransportFeeValue: spec.PaidTransportFeeValue,
				ID:                    int64(spec.TeacherPaymentID),
			})
			if err != nil {
				return fmt.Errorf("qtx.EditTeacherPayment(): %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return []entity.TeacherPaymentID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return teacherPaymentIDs, nil
}

func (s teachingServiceImpl) RemoveTeacherPayments(ctx context.Context, teacherPaymentIDs []entity.TeacherPaymentID) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		teacherPaymentIDsint64 := make([]int64, 0, len(teacherPaymentIDs))
		for _, teacherPaymentID := range teacherPaymentIDs {
			teacherPaymentIDsint64 = append(teacherPaymentIDsint64, int64(teacherPaymentID))
		}
		affectedAttendancedIDsInt64, err := qtx.GetTeacherPaymentAttendanceIdsByIds(newCtx, teacherPaymentIDsint64)
		if err != nil {
			return fmt.Errorf("GetTeacherPaymentAttendanceIdsByIds(): %w", err)
		}

		err = s.entityService.DeleteTeacherPayments(newCtx, teacherPaymentIDs)
		if err != nil {
			return fmt.Errorf("entityService.DeleteTeacherPayments(): %w", err)
		}

		err = qtx.SetAttendancesIsPaidStatusByIds(newCtx, mysql.SetAttendancesIsPaidStatusByIdsParams{
			IsPaid: 0,
			Ids:    affectedAttendancedIDsInt64,
		})
		if err != nil {
			return fmt.Errorf("qtx.SetAttendancesIsPaidStatusByIds(): %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}
