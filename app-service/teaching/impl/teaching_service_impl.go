package teaching

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	enrollmentPayments, err := s.entityService.GetEnrollmentPayments(ctx, paginationSpec, timeFilter, true)
	if err != nil {
		return []entity.EnrollmentPayment{}, fmt.Errorf("entityService.GetEnrollmentPayments(): %v", err)
	}

	return enrollmentPayments.EnrollmentPayments, nil
}

func (s teachingServiceImpl) CalculateStudentEnrollmentInvoice(ctx context.Context, studentEnrollmentID entity.StudentEnrollmentID) (teaching.StudentEnrollmentInvoice, error) {
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
	var penaltyFeeValue int32 = 0
	slts, err := s.mySQLQueries.GetSLTWithNegativeQuotaByEnrollmentId(ctx, int64(studentEnrollmentID))
	if err != nil {
		return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("mySQLQueries.GetSLTWithNegativeQuotaByEnrollmentId(): %w", err)
	}
	for _, slt := range slts {
		if slt.Quota >= 0 {
			continue
		}
		// make the value positive
		unpaidQuota := slt.Quota * -1
		// penaly = penaltyValue (in Rupiah) per 1 course cycle
		penaltyFeeValue += int32((unpaidQuota-1)/teaching.Default_OneCourseCycle+1) * teaching.Default_PenaltyFeeValue
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

func (s teachingServiceImpl) GetPresencesByClassID(ctx context.Context, spec teaching.GetPresencesByClassIDSpec) (teaching.GetPresencesByClassIDResult, error) {
	getPresencesSpec := entity.GetPresencesSpec{
		ClassID:   spec.ClassID,
		StudentID: spec.StudentID,
		TimeSpec:  spec.TimeSpec,
	}
	getPresencesResult, err := s.entityService.GetPresences(ctx, spec.PaginationSpec, getPresencesSpec)
	if err != nil {
		return teaching.GetPresencesByClassIDResult{}, fmt.Errorf("entityService.GetPresences(): %v", err)
	}

	return teaching.GetPresencesByClassIDResult{
		Presences:        getPresencesResult.Presences,
		PaginationResult: getPresencesResult.PaginationResult,
	}, nil
}

func (s teachingServiceImpl) AddPresence(ctx context.Context, spec teaching.AddPresenceSpec, autoCreateSLT bool) ([]entity.PresenceID, error) {
	presenceIDs := make([]entity.PresenceID, 0)

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
		//   if all <= 0, we decrement the last SLT (thus becoming negative for paymentPenalty later).
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

		// Insert presences
		specs := make([]entity.InsertPresenceSpec, 0, len(studentEnrollments))
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

			specs = append(specs, entity.InsertPresenceSpec{
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

		presenceIDs, err = s.entityService.InsertPresences(newCtx, specs)
		if err != nil {
			return fmt.Errorf("entityService.InsertPresences(): %w", err)
		}

		return nil
	})
	if err != nil {
		return []entity.PresenceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return presenceIDs, nil
}

func (s teachingServiceImpl) autoRegisterSLT(ctx context.Context, studentEnrollmentID entity.StudentEnrollmentID, quota float64) (entity.StudentLearningTokenID, error) {
	enrollmentID := entity.StudentEnrollmentID(studentEnrollmentID)
	invoice, err := s.CalculateStudentEnrollmentInvoice(ctx, enrollmentID)
	if err != nil {
		return entity.StudentLearningTokenID_None, fmt.Errorf("CalculateStudentEnrollmentInvoice(): %w", err)
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

func (s teachingServiceImpl) EditPresence(ctx context.Context, spec teaching.EditPresenceSpec) ([]entity.PresenceID, error) {
	presenceIDs := make([]entity.PresenceID, 0)

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		rowResults, err := qtx.GetPresenceIdsOfSameClassAndDate(newCtx, int64(spec.PresenceID))
		if err != nil {
			return fmt.Errorf("qtx.GetPresenceIdsOfSameClassAndDate(): %w", err)
		}

		presenceIDsInt := make([]int64, 0, len(rowResults))
		for _, rowResult := range rowResults {
			if util.Int32ToBool(rowResult.IsPaid) == true {
				return errs.ErrModifyingPaidPresence
			}
			presenceIDs = append(presenceIDs, entity.PresenceID(rowResult.ID))
			presenceIDsInt = append(presenceIDsInt, rowResult.ID)
		}

		err = qtx.EditPresences(newCtx, mysql.EditPresencesParams{
			TeacherID:             sql.NullInt64{Int64: int64(spec.TeacherID), Valid: true},
			Date:                  spec.Date,
			UsedStudentTokenQuota: spec.UsedStudentTokenQuota,
			Duration:              spec.Duration,
			Note:                  spec.Note,
			Ids:                   presenceIDsInt,
		})
		if err != nil {
			return fmt.Errorf("qtx.EditPresences(): %w", err)
		}

		return nil
	})
	if err != nil {
		return []entity.PresenceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return presenceIDs, nil
}

func (s teachingServiceImpl) RemovePresence(ctx context.Context, presenceID entity.PresenceID) ([]entity.PresenceID, error) {
	deletedPresenceIDs := make([]entity.PresenceID, 0)

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		rowResults, err := qtx.GetPresenceIdsOfSameClassAndDate(newCtx, int64(presenceID))
		if err != nil {
			return fmt.Errorf("qtx.GetPresenceIdsOfSameClassAndDate(): %w", err)
		}

		presenceIDsInt := make([]int64, 0, len(rowResults))
		for _, rowResult := range rowResults {
			if util.Int32ToBool(rowResult.IsPaid) == true {
				return errs.ErrModifyingPaidPresence
			}
			deletedPresenceIDs = append(deletedPresenceIDs, entity.PresenceID(rowResult.ID))
			presenceIDsInt = append(presenceIDsInt, rowResult.ID)
		}

		err = qtx.DeletePresencesByIds(newCtx, presenceIDsInt)
		if err != nil {
			return fmt.Errorf("qtx.DeletePresencesByIds(): %w", err)
		}

		return nil
	})
	if err != nil {
		return []entity.PresenceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return deletedPresenceIDs, nil
}
