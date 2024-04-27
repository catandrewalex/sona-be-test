package teaching

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	latestPaymentDate, err := s.mySQLQueries.GetLatestEnrollmentPaymentDateByStudentId(ctx, int64(studentEnrollmentID))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("mySQLQueries.GetLatestEnrollmentPaymentDateByStudentId(): %w", err)
		}
	}
	lastDateBeforePenalty := time.Date(latestPaymentDate.(time.Time).Year(), latestPaymentDate.(time.Time).AddDate(0, 1, 0).Month(), 10, 0, 0, 0, 0, time.UTC)
	penaltyFeeValue := teaching.Default_PenaltyFeeValue * int32(time.Since(lastDateBeforePenalty).Hours()/24)
	fmt.Printf("latestPaymentDate: %v\n", latestPaymentDate)
	fmt.Printf("lastDateBeforePenalty: %v\n", lastDateBeforePenalty)

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
		for _, rowResult := range rowResults {
			if util.Int32ToBool(rowResult.IsPaid) == true {
				return errs.ErrModifyingPaidAttendance
			}
			attendanceIDs = append(attendanceIDs, entity.AttendanceID(rowResult.ID))
			attendanceIDsInt = append(attendanceIDsInt, rowResult.ID)
		}

		err = qtx.EditAttendances(newCtx, mysql.EditAttendancesParams{
			TeacherID:             sql.NullInt64{Int64: int64(spec.TeacherID), Valid: true},
			Date:                  spec.Date,
			UsedStudentTokenQuota: spec.UsedStudentTokenQuota,
			Duration:              spec.Duration,
			Note:                  spec.Note,
			Ids:                   attendanceIDsInt,
		})
		if err != nil {
			return fmt.Errorf("qtx.EditAttendances(): %w", err)
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
		for _, rowResult := range rowResults {
			if util.Int32ToBool(rowResult.IsPaid) == true {
				return errs.ErrModifyingPaidAttendance
			}
			deletedAttendanceIDs = append(deletedAttendanceIDs, entity.AttendanceID(rowResult.ID))
			attendanceIDsInt = append(attendanceIDsInt, rowResult.ID)
		}

		err = qtx.DeleteAttendancesByIds(newCtx, attendanceIDsInt)
		if err != nil {
			return fmt.Errorf("qtx.DeleteAttendancesByIds(): %w", err)
		}

		return nil
	})
	if err != nil {
		return []entity.AttendanceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return deletedAttendanceIDs, nil
}

func (s teachingServiceImpl) GetTeacherSalaryInvoices(ctx context.Context, spec teaching.GetTeacherSalaryInvoicesSpec) ([]teaching.TeacherSalaryInvoice, error) {
	getAttendancesForTeacherSalarySpec := entity.GetAttendancesForTeacherSalarySpec{
		TeacherID: spec.TeacherID,
		ClassID:   spec.ClassID,
		TimeSpec:  spec.TimeSpec,
	}

	attendances, err := s.entityService.GetAttendancesForTeacherSalary(ctx, getAttendancesForTeacherSalarySpec)
	if err != nil {
		return []teaching.TeacherSalaryInvoice{}, fmt.Errorf("entityService.GetAttendances(): %v", err)
	}

	teacherSalaryInvoices := make([]teaching.TeacherSalaryInvoice, 0, len(attendances))
	for _, attendance := range attendances {
		teacherSalaryInvoices = append(teacherSalaryInvoices, teaching.TeacherSalaryInvoice{
			Attendance:                    attendance,
			CourseFeeFullValue:            int32(float64(attendance.StudentLearningToken.CourseFeeValue) * attendance.UsedStudentTokenQuota / float64(teaching.Default_OneCourseCycle)),
			TransportFeeFullValue:         int32(float64(attendance.StudentLearningToken.TransportFeeValue) * attendance.UsedStudentTokenQuota / float64(teaching.Default_OneCourseCycle)),
			CourseFeeSharingPercentage:    teaching.Default_CourseFeeSharingPercentage,
			TransportFeeSharingPercentage: teaching.Default_TransportFeeSharingPercentage,
		})
	}

	return teacherSalaryInvoices, nil
}

func (s teachingServiceImpl) SubmitTeacherSalaries(ctx context.Context, specs []teaching.SubmitTeacherSalariesSpec) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		insertSpecs := make([]entity.InsertTeacherSalarySpec, 0, len(specs))
		affectedAttendancedIDsInt64 := make([]int64, 0, len(specs))
		for _, spec := range specs {
			insertSpecs = append(insertSpecs, entity.InsertTeacherSalarySpec{
				AttendanceID:          spec.AttendanceID,
				PaidCourseFeeValue:    spec.PaidCourseFeeValue,
				PaidTransportFeeValue: spec.PaidTransportFeeValue,
			})
			affectedAttendancedIDsInt64 = append(affectedAttendancedIDsInt64, int64(spec.AttendanceID))
		}

		_, err := s.entityService.InsertTeacherSalaries(newCtx, insertSpecs)
		if err != nil {
			return fmt.Errorf("entityService.InsertTeacherSalaries(): %w", err)
		}

		err = qtx.SetAttendancesIsPaidStatusByIds(newCtx, mysql.SetAttendancesIsPaidStatusByIdsParams{
			IsPaid: 1,
			Ids:    affectedAttendancedIDsInt64,
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) EditTeacherSalaries(ctx context.Context, specs []teaching.EditTeacherSalariesSpec) ([]entity.TeacherSalaryID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountTeacherSalariesByIds)
	if errV != nil {
		return []entity.TeacherSalaryID{}, errV
	}

	teacherSalaryIDs := make([]entity.TeacherSalaryID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			teacherSalaryIDs = append(teacherSalaryIDs, spec.TeacherSalaryID)
			err := qtx.EditTeacherSalary(newCtx, mysql.EditTeacherSalaryParams{
				PaidCourseFeeValue:    spec.PaidCourseFeeValue,
				PaidTransportFeeValue: spec.PaidTransportFeeValue,
				ID:                    int64(spec.TeacherSalaryID),
			})
			if err != nil {
				return fmt.Errorf("qtx.EditTeacherSalary(): %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return []entity.TeacherSalaryID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return teacherSalaryIDs, nil
}

func (s teachingServiceImpl) RemoveTeacherSalaries(ctx context.Context, teacherSalaryIDs []entity.TeacherSalaryID) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		teacherSalaryIDsint64 := make([]int64, 0, len(teacherSalaryIDs))
		for _, teacherSalaryID := range teacherSalaryIDs {
			teacherSalaryIDsint64 = append(teacherSalaryIDsint64, int64(teacherSalaryID))
		}
		affectedAttendancedIDsInt64, err := qtx.GetTeacherSalaryAttendanceIdsByIds(newCtx, teacherSalaryIDsint64)
		if err != nil {
			return fmt.Errorf("GetTeacherSalaryAttendanceIdsByIds(): %w", err)
		}

		err = s.entityService.DeleteTeacherSalaries(newCtx, teacherSalaryIDs)
		if err != nil {
			return fmt.Errorf("entityService.DeleteTeacherSalaries(): %w", err)
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
