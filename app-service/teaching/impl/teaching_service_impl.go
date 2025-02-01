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
	"sonamusica-backend/app-service/identity"
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

func (s teachingServiceImpl) GetUserTeachingInfo(ctx context.Context, id identity.UserID) (teaching.UserTeachingInfo, error) {
	userTeachingInfoRow, err := s.mySQLQueries.GetUserTeacherIdAndStudentId(ctx, int64(id))
	if err != nil {
		return teaching.UserTeachingInfo{}, fmt.Errorf("mySQLQueries.GetUserTeacherIdAndStudentId(): %w", err)
	}

	// TODO: this struct is not created via result_construtor.go. Probably refactor?
	userTeachingInfo := teaching.UserTeachingInfo{
		TeacherID: entity.TeacherID(userTeachingInfoRow.TeacherID.Int64),
		StudentID: entity.StudentID(userTeachingInfoRow.StudentID.Int64),
		IsTeacher: userTeachingInfoRow.TeacherID.Valid,
		IsStudent: userTeachingInfoRow.StudentID.Valid,
	}

	return userTeachingInfo, nil
}

func (s teachingServiceImpl) IsUserInvolvedInClass(ctx context.Context, userId identity.UserID, classId entity.ClassID) (bool, error) {
	isInvolved, err := s.mySQLQueries.IsUserIdInvolvedInClassId(ctx, mysql.IsUserIdInvolvedInClassIdParams{
		UserID:  int64(userId),
		ClassID: int64(classId),
	})
	if err != nil {
		return false, fmt.Errorf("mySQLQueries.IsUserIdInvolvedInClassId(): %v", err)
	}

	return isInvolved, nil
}

func (s teachingServiceImpl) IsUserInvolvedInAttendance(ctx context.Context, userId identity.UserID, attendanceId entity.AttendanceID) (bool, error) {
	isInvolved, err := s.mySQLQueries.IsUserIdInvolvedInAttendanceId(ctx, mysql.IsUserIdInvolvedInAttendanceIdParams{
		UserID:       int64(userId),
		AttendanceID: int64(attendanceId),
	})
	if err != nil {
		return false, fmt.Errorf("mySQLQueries.IsUserIdInvolvedInAttendanceId(): %v", err)
	}

	return isInvolved, nil
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
	var courseFeeValueFinal int32
	var splittedTransportFeeFinal int32
	var penaltyFeeValueFinal int32
	var lastPaymentDateFinal *time.Time
	var daysLateFinal int32

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		studentEnrollment, err := s.entityService.GetStudentEnrollmentById(ctx, studentEnrollmentID)
		if err != nil {
			return fmt.Errorf("entityService.GetStudentEnrollmentById(): %w", err)
		}

		courseFeeValue := studentEnrollment.ClassInfo.Course.DefaultFee
		if studentEnrollment.ClassInfo.TeacherSpecialFee != 0 {
			courseFeeValue = studentEnrollment.ClassInfo.TeacherSpecialFee
		}

		// calculate Course Fee Penalty (e.g. due to late payment)
		latestPaymentDate, err := qtx.GetLatestEnrollmentPaymentDateByStudentEnrollmentId(newCtx, sql.NullInt64{Int64: int64(studentEnrollmentID), Valid: true})
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("qtx.GetLatestEnrollmentPaymentDateByStudentEnrollmentId(): %w", err)
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
		classIdToTotalStudents, err := qtx.GetClassesTotalStudentsByClassIds(newCtx, []int64{int64(studentEnrollment.ClassInfo.ClassID)})
		if err != nil {
			return fmt.Errorf("qtx.GetClassesTotalStudentsByClassIds(): %w", err)
		}
		if len(classIdToTotalStudents) > 0 && classIdToTotalStudents[0].TotalStudents > 1 {
			splittedTransportFee /= int32(classIdToTotalStudents[0].TotalStudents)
		}

		// assign to the top level variable, to be used outside logic block of the ExecuteInTransaction()
		courseFeeValueFinal = courseFeeValue
		splittedTransportFeeFinal = splittedTransportFee
		penaltyFeeValueFinal = penaltyFeeValue
		lastPaymentDateFinal = lastPaymentDate
		daysLateFinal = daysLate

		return nil
	})
	if err != nil {
		return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return teaching.StudentEnrollmentInvoice{
		BalanceTopUp:      teaching.Default_BalanceTopUp,
		BalanceBonus:      0,
		CourseFeeValue:    courseFeeValueFinal,
		TransportFeeValue: splittedTransportFeeFinal,
		PenaltyFeeValue:   penaltyFeeValueFinal,
		DiscountFeeValue:  0,
		LastPaymentDate:   lastPaymentDateFinal,
		DaysLate:          daysLateFinal,
	}, nil
}

func (s teachingServiceImpl) SubmitEnrollmentPayment(ctx context.Context, spec teaching.SubmitStudentEnrollmentPaymentSpec) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		_, err := s.entityService.InsertEnrollmentPayments(newCtx, []entity.InsertEnrollmentPaymentSpec{
			{
				StudentEnrollmentID: spec.StudentEnrollmentID,
				PaymentDate:         spec.PaymentDate,
				BalanceTopUp:        spec.BalanceTopUp,
				BalanceBonus:        spec.BalanceBonus,
				CourseFeeValue:      spec.CourseFeeValue,
				TransportFeeValue:   spec.TransportFeeValue,
				PenaltyFeeValue:     spec.PenaltyFeeValue,
				DiscountFeeValue:    spec.DiscountFeeValue,
			},
		})
		if err != nil {
			return fmt.Errorf("entityService.InsertEnrollmentPayments(): %w", err)
		}

		// Upsert StudentLearningTokens
		var totalBalanceTopUp = float64(spec.BalanceTopUp + spec.BalanceBonus)
		courseFeeQuarterValue := teaching.CalculateSLTFeeQuarterFromEP(spec.CourseFeeValue, spec.BalanceTopUp)
		transportFeeQuarterValue := teaching.CalculateSLTFeeQuarterFromEP(spec.TransportFeeValue, spec.BalanceTopUp)
		existingSLT, err := qtx.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter(newCtx, mysql.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarterParams{
			EnrollmentID:             int64(spec.StudentEnrollmentID),
			CourseFeeQuarterValue:    courseFeeQuarterValue,
			TransportFeeQuarterValue: transportFeeQuarterValue,
		})
		isNeedInsert := errors.Is(err, sql.ErrNoRows)
		if isNeedInsert {
			err = nil
		}
		if err != nil {
			return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter(): %w", err)
		}

		if isNeedInsert {
			_, err = s.entityService.InsertStudentLearningTokens(newCtx, []entity.InsertStudentLearningTokenSpec{
				{
					StudentEnrollmentID:      spec.StudentEnrollmentID,
					Quota:                    totalBalanceTopUp,
					CourseFeeQuarterValue:    courseFeeQuarterValue,
					TransportFeeQuarterValue: transportFeeQuarterValue,
				},
			})
			if err != nil {
				return fmt.Errorf("entityService.InsertStudentLearningTokens(): %w", err)
			}
		} else {
			err := qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota:         totalBalanceTopUp,
				LastUpdatedAt: time.Now().UTC(),
				ID:            existingSLT.ID,
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

		updatedSLT, err := qtx.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter(newCtx, mysql.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarterParams{
			EnrollmentID:             prevEP.StudentEnrollmentID,
			CourseFeeQuarterValue:    teaching.CalculateSLTFeeQuarterFromEP(prevEP.CourseFeeValue, prevEP.BalanceTopUp),
			TransportFeeQuarterValue: teaching.CalculateSLTFeeQuarterFromEP(prevEP.TransportFeeValue, prevEP.BalanceTopUp),
		})
		skipSLTUpdate := errors.Is(err, sql.ErrNoRows)
		if skipSLTUpdate {
			mainLog.Warn("EnrollmentPayment with ID='%d' doesn't have studentLearningToken, check for bad data possibility. Skipping to update studentLearningToken.", prevEP.EnrollmentPaymentID)
			err = nil
		}
		if err != nil {
			return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter(): %w", err)
		}

		if !skipSLTUpdate {
			quotaChange := float64(spec.BalanceBonus - prevEP.BalanceBonus)
			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota:         quotaChange,
				LastUpdatedAt: time.Now().UTC(),
				ID:            updatedSLT.ID,
			})
			if err != nil {
				return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
			}
		}

		err = qtx.UpdateEnrollmentPaymentOnSafeAttributes(newCtx, mysql.UpdateEnrollmentPaymentOnSafeAttributesParams{
			PaymentDate:      spec.PaymentDate,
			BalanceBonus:     spec.BalanceBonus,
			DiscountFeeValue: spec.DiscountFeeValue,
			ID:               int64(spec.EnrollmentPaymentID),
		})
		if err != nil {
			return fmt.Errorf("entityService.UpdateEnrollmentPaymentOnSafeAttributes(): %w", err)
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

		updatedSLT, err := qtx.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter(newCtx, mysql.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarterParams{
			EnrollmentID:             prevEP.StudentEnrollmentID,
			CourseFeeQuarterValue:    teaching.CalculateSLTFeeQuarterFromEP(prevEP.CourseFeeValue, prevEP.BalanceTopUp),
			TransportFeeQuarterValue: teaching.CalculateSLTFeeQuarterFromEP(prevEP.TransportFeeValue, prevEP.BalanceTopUp),
		})
		skipSLTUpdate := errors.Is(err, sql.ErrNoRows)
		if skipSLTUpdate {
			mainLog.Warn("EnrollmentPayment with ID='%d' doesn't have studentLearningToken, check for bad data possibility. Skipping to update studentLearningToken.", prevEP.EnrollmentPaymentID)
			err = nil
		}
		if err != nil {
			return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter(): %w", err)
		}

		if !skipSLTUpdate {
			quotaChange := -1 * (prevEP.BalanceTopUp + prevEP.BalanceBonus)
			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota:         float64(quotaChange),
				LastUpdatedAt: time.Now().UTC(),
				ID:            updatedSLT.ID,
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

func (s teachingServiceImpl) EditClassesConfigs(ctx context.Context, specs []teaching.EditClassConfigSpec) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			// no changes, skip the spec
			if spec.IsDeactivated == nil && spec.AutoOweAttendanceToken == nil {
				continue
			}
			var err error
			classId := int64(spec.ClassID)

			// route to corresponding SQL update query, based on available fields in the spec
			if spec.IsDeactivated != nil && spec.AutoOweAttendanceToken != nil {
				err = qtx.UpdateClassConfig(newCtx, mysql.UpdateClassConfigParams{
					AutoOweAttendanceToken: util.BoolToInt32(*spec.AutoOweAttendanceToken),
					IsDeactivated:          util.BoolToInt32(*spec.IsDeactivated),
					ID:                     classId,
				})
				if err != nil {
					return fmt.Errorf("qtx.UpdateClassConfig(): %w", err)
				}
			} else if spec.IsDeactivated != nil {
				err = qtx.UpdateClassActivation(newCtx, mysql.UpdateClassActivationParams{
					IsDeactivated: util.BoolToInt32(*spec.IsDeactivated),
					ID:            classId,
				})
				if err != nil {
					return fmt.Errorf("qtx.UpdateClassActivation(): %w", err)
				}
			} else if spec.AutoOweAttendanceToken != nil {
				err = qtx.UpdateClassAutoOweToken(newCtx, mysql.UpdateClassAutoOweTokenParams{
					AutoOweAttendanceToken: util.BoolToInt32(*spec.AutoOweAttendanceToken),
					ID:                     classId,
				})
				if err != nil {
					return fmt.Errorf("qtx.UpdateClassAutoOweToken(): %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}
func (s teachingServiceImpl) EditClassesCourses(ctx context.Context, specs []teaching.EditClassCourseSpec) error {
	// TODO(FerdiantJoshua): also automatically add new SLT, which correspond to the new course price.
	// if it already exists, use IncrementSLTQuotaById(), with incrementedQuota==0, so that it becomes the latest SLT.
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateClassCourse(newCtx, mysql.UpdateClassCourseParams{
				CourseID: int64(spec.CourseID),
				ID:       int64(spec.ClassID),
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateClassCourse(): %w", err)
			}

			enrollments, err := s.entityService.GetStudentEnrollmentsByClassId(newCtx, spec.ClassID)
			if err != nil {
				return fmt.Errorf("entityService.GetStudentEnrollmentsByClassId(): %w", err)
			}

			for _, enrollment := range enrollments {
				fmt.Printf("enrollmentClassInfo: %#v\n", enrollment.ClassInfo)
				courseFee := enrollment.ClassInfo.Course.DefaultFee
				if enrollment.ClassInfo.TeacherSpecialFee != 0 {
					courseFee = enrollment.ClassInfo.TeacherSpecialFee
				}
				existingSLT, err := qtx.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter(newCtx, mysql.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarterParams{
					EnrollmentID:             int64(enrollment.StudentEnrollmentID),
					CourseFeeQuarterValue:    courseFee / teaching.Default_OneCourseCycle,
					TransportFeeQuarterValue: enrollment.ClassInfo.TransportFee / teaching.Default_OneCourseCycle,
				})
				isNeedInsert := false
				if errors.Is(err, sql.ErrNoRows) {
					isNeedInsert = true
					err = nil
				} else if err != nil {
					return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter(): %w", err)
				}

				if isNeedInsert {
					// we don't need the newly added SLT ID, so we can safely ignore it
					_, err := s.autoRegisterSLT(newCtx, entity.StudentEnrollmentID(enrollment.StudentEnrollmentID), 0)
					if err != nil {
						return fmt.Errorf("autoRegisterSLT(): %w", err)
					}
				} else {
					// the goal is to set token's LastUpdatedAt to current date
					err := qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
						Quota:         0,
						LastUpdatedAt: time.Now().UTC(),
						ID:            existingSLT.ID,
					})
					if err != nil {
						return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) GetSLTsByClassID(ctx context.Context, classID entity.ClassID) ([]teaching.StudentIDToSLTs, error) {
	studentLearningTokenRows, err := s.mySQLQueries.GetSLTByClassIdForAttendanceInfo(ctx, int64(classID))
	if err != nil {
		return []teaching.StudentIDToSLTs{}, fmt.Errorf("mySQLQueries.GetSLTByClassIdForAttendanceInfo(): %w", err)
	}

	studentIdToSlts := make(map[entity.StudentID][]entity.StudentLearningToken_Minimal, 1)
	for _, sltRow := range studentLearningTokenRows {
		studentId := entity.StudentID(sltRow.StudentID)
		// TODO: refactor this, as this in a non-standard way of converting mysql's rows into Entity's struct. Check existing result_constructor.go(s) for reference.
		slt := entity.StudentLearningToken_Minimal{
			StudentLearningTokenID: entity.StudentLearningTokenID(sltRow.StudentLearningTokenID),
			Quota:                  sltRow.Quota,
			CourseFeeValue:         sltRow.CourseFeeQuarterValue * 4,
			TransportFeeValue:      sltRow.TransportFeeQuarterValue * 4,
			CreatedAt:              sltRow.CreatedAt,
			LastUpdatedAt:          sltRow.LastUpdatedAt,
		}

		currSLTs, ok := studentIdToSlts[studentId]
		if !ok {
			studentIdToSlts[studentId] = []entity.StudentLearningToken_Minimal{slt}
		} else {
			studentIdToSlts[studentId] = append(currSLTs, slt)
		}
	}

	getSLTsByClassIDResults := make([]teaching.StudentIDToSLTs, 0, len(studentIdToSlts))
	for studentId, slts := range studentIdToSlts {
		getSLTsByClassIDResults = append(getSLTsByClassIDResults, teaching.StudentIDToSLTs{
			StudentID:             studentId,
			StudentLearningTokens: slts,
		})
	}

	return getSLTsByClassIDResults, nil
}

func (s teachingServiceImpl) GetAttendancesByClassID(ctx context.Context, spec teaching.GetAttendancesByClassIDSpec) (teaching.GetAttendancesByClassIDResult, error) {
	getAttendancesSpec := entity.GetAttendancesSpec{
		ClassID:   spec.ClassID,
		StudentID: spec.StudentID,
		TimeSpec:  spec.TimeSpec,
	}
	getAttendancesResult, err := s.entityService.GetAttendances(ctx, spec.PaginationSpec, getAttendancesSpec, true)
	if err != nil {
		return teaching.GetAttendancesByClassIDResult{}, fmt.Errorf("entityService.GetAttendances(): %v", err)
	}

	return teaching.GetAttendancesByClassIDResult{
		Attendances:      getAttendancesResult.Attendances,
		PaginationResult: getAttendancesResult.PaginationResult,
	}, nil
}

func (s teachingServiceImpl) AddAttendancesBatch(ctx context.Context, specs []teaching.AddAttendanceSpec) ([]entity.AttendanceID, error) {
	attendanceIDs := make([]entity.AttendanceID, 0)

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			ids, err := s.AddAttendance(newCtx, spec)
			if err != nil {
				return fmt.Errorf("AddAttendance(): %w", err)
			}
			attendanceIDs = append(attendanceIDs, ids...)
		}

		return nil
	})
	if err != nil {
		return []entity.AttendanceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return attendanceIDs, nil
}

func (s teachingServiceImpl) AddAttendance(ctx context.Context, spec teaching.AddAttendanceSpec) ([]entity.AttendanceID, error) {
	attendanceIDs := make([]entity.AttendanceID, 0)

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		autoOweTokenMode, err := qtx.GetClassAutoOweTokenMode(newCtx, int64(spec.ClassID))
		if err != nil {
			return fmt.Errorf("qtx.GetClassAutoOweTokenMode(): %w", err)
		}
		autoOweSLT := util.Int32ToBool(autoOweTokenMode)

		studentEnrollments, err := s.entityService.GetStudentEnrollmentsByClassId(newCtx, spec.ClassID)
		if err != nil {
			return fmt.Errorf("entityService.GetStudentEnrollmentsByClassId(): %w", err)
		}
		if len(studentEnrollments) == 0 {
			return fmt.Errorf("classID='%d': %w", spec.ClassID, errs.ErrClassHaveNoStudent)
		}

		studentEnrollmentIDsInt64 := make([]int64, 0, len(studentEnrollments))
		for _, studentEnrollment := range studentEnrollments {
			studentEnrollmentIDsInt64 = append(studentEnrollmentIDsInt64, int64(studentEnrollment.StudentEnrollmentID))
		}

		enrollmentIDToEarliestSLTID := make(map[entity.StudentEnrollmentID]entity.StudentLearningTokenID, 0)
		// students may have > 1 SLT, we'll pick the one with earliest non-zero quota.
		//   if all <= 0, we decrement the last SLT (thus becoming negative).
		earliestAvailableSLTs, err := qtx.GetEarliestAvailableSLTsByStudentEnrollmentIds(newCtx, studentEnrollmentIDsInt64)
		if err != nil {
			return fmt.Errorf("qtx.GetEarliestAvailableSLTsByStudentEnrollmentIds(): %w", err)
		}

		for _, earliestAvailableSLT := range earliestAvailableSLTs {
			// if `Class` "autoOweAttendanceToken" is false, we will not let the resulting SLT quota (currentQuota - usedQuota) to be < 0.
			// instead, we will let the `Attendance` to have no SLT. We expect admin to assign the SLT manually.
			if !autoOweSLT && earliestAvailableSLT.Quota-spec.UsedStudentTokenQuota < 0 {
				continue
			}
			enrollmentID := entity.StudentEnrollmentID(earliestAvailableSLT.EnrollmentID)
			enrollmentIDToEarliestSLTID[enrollmentID] = entity.StudentLearningTokenID(earliestAvailableSLT.StudentLearningTokenID)

			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota:         spec.UsedStudentTokenQuota * -1,
				LastUpdatedAt: time.Now().UTC(),
				ID:            earliestAvailableSLT.StudentLearningTokenID,
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
				// only autoRegisterSLT with negative quota, if the `Class` "autoOweAttendanceToken" mode is true.
				// else, we let the `Attendance` to have no SLT. We expect admin to assign the SLT manually.
				if autoOweSLT {
					mainLog.Warn("studentEnrollment='%d' doesn't have any studentLearningToken (SLT). Creating a new negative quota SLT as 'autoCreateSLT' is true.", studentEnrollment.StudentEnrollmentID)
					newSLTID, err := s.autoRegisterSLT(newCtx, entity.StudentEnrollmentID(studentEnrollment.StudentEnrollmentID), spec.UsedStudentTokenQuota*-1)
					if err != nil {
						return fmt.Errorf("autoRegisterSLT(): %w", err)
					}
					sltID = newSLTID
				}
			}

			specs = append(specs, entity.InsertAttendanceSpec{
				ClassID:                spec.ClassID,
				TeacherID:              spec.TeacherID,
				StudentID:              entity.StudentID(studentEnrollment.StudentInfo.StudentID),
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
	var newSLTID entity.StudentLearningTokenID
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		invoice, err := s.GetEnrollmentPaymentInvoice(ctx, studentEnrollmentID)
		if err != nil {
			return fmt.Errorf("GetEnrollmentPaymentInvoice(): %w", err)
		}

		newSLTIDs, err := s.entityService.InsertStudentLearningTokens(ctx, []entity.InsertStudentLearningTokenSpec{
			{
				StudentEnrollmentID:      studentEnrollmentID,
				Quota:                    quota,
				CourseFeeQuarterValue:    teaching.CalculateSLTFeeQuarterFromEP(invoice.CourseFeeValue, invoice.BalanceTopUp),
				TransportFeeQuarterValue: teaching.CalculateSLTFeeQuarterFromEP(invoice.TransportFeeValue, invoice.BalanceTopUp),
			},
		})
		if err != nil {
			return fmt.Errorf("entityService.InsertStudentLearningTokens(): %w", err)
		}

		newSLTID = newSLTIDs[0]
		return nil
	})
	if err != nil {
		return entity.StudentLearningTokenID_None, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return newSLTID, nil
}

func (s teachingServiceImpl) AssignAttendanceToken(ctx context.Context, spec teaching.AssignAttendanceTokenSpec) error {
	errV := util.ValidateUpdateSpecs(ctx, []teaching.AssignAttendanceTokenSpec{spec}, s.mySQLQueries.CountAttendancesByIds)
	if errV != nil {
		return errV
	}

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		rowResult, err := qtx.GetAttendanceForTokenAssignment(newCtx, spec.GetInt64ID())
		if err != nil {
			return fmt.Errorf("GetAttendanceForTokenAssignment(): %w", err)
		}
		if util.Int32ToBool(rowResult.IsPaid) {
			return errs.ErrModifyingPaidAttendance
		}

		// update (1) previous token's quota, and (2) new token's quota
		if rowResult.TokenID.Valid {
			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota:         rowResult.UsedStudentTokenQuota,
				LastUpdatedAt: time.Now().UTC(),
				ID:            rowResult.TokenID.Int64,
			})
			if err != nil {
				return fmt.Errorf("IncrementSLTQuotaById(): %w", err)
			}
		}
		err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
			Quota:         -1 * rowResult.UsedStudentTokenQuota,
			LastUpdatedAt: time.Now().UTC(),
			ID:            int64(spec.StudentLearningTokenID),
		})
		if err != nil {
			return fmt.Errorf("IncrementSLTQuotaById(): %w", err)
		}

		// assign the new token to the attendance
		err = qtx.AssignAttendanceToken(newCtx, mysql.AssignAttendanceTokenParams{
			ID:      spec.GetInt64ID(),
			TokenID: sql.NullInt64{Int64: int64(spec.StudentLearningTokenID), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("AssignAttendanceToken(): %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
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

			if rowResult.TokenID.Valid {
				sltIDIntToUsedQuota[rowResult.TokenID.Int64] = rowResult.UsedStudentTokenQuota
			}
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
				Quota:         quotaChange,
				LastUpdatedAt: time.Now().UTC(),
				ID:            sltIDInt,
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

			if rowResult.TokenID.Valid {
				sltIDIntToUsedQuota[rowResult.TokenID.Int64] = rowResult.UsedStudentTokenQuota
			}
		}

		err = qtx.DeleteAttendancesByIds(newCtx, attendanceIDsInt)
		if err != nil {
			return fmt.Errorf("qtx.DeleteAttendancesByIds(): %w", err)
		}

		for sltIDInt, usedQuota := range sltIDIntToUsedQuota {
			quotaChange := usedQuota
			err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota:         quotaChange,
				LastUpdatedAt: time.Now().UTC(),
				ID:            sltIDInt,
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

func (s teachingServiceImpl) GetTeachersForPayment(ctx context.Context, spec teaching.GetTeachersForPaymentSpec) (teaching.GetTeachersForPaymentResult, error) {
	spec.Pagination.SetDefaultOnInvalidValues()
	limit, offset := spec.Pagination.GetLimitAndOffset()

	spec.TimeSpec.SetDefaultForZeroValues()

	var teachersForPayments []teaching.TeacherForPayment
	var totalResults int64
	if spec.IsPaid { // for paid teachers, we fetch data from `TeacherPayment`
		var teacherRows = make([]mysql.GetPaidTeachersRow, 0)
		err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
			var err error
			teacherRows, err = qtx.GetPaidTeachers(newCtx, mysql.GetPaidTeachersParams{
				StartDate: spec.StartDatetime,
				EndDate:   spec.EndDatetime,
				Limit:     int32(limit),
				Offset:    int32(offset),
			})
			if err != nil {
				return fmt.Errorf("qtx.GetPaidTeachers(): %w", err)
			}

			totalResults, err = qtx.CountUnpaidTeachers(newCtx, mysql.CountUnpaidTeachersParams{
				StartDate: spec.StartDatetime,
				EndDate:   spec.EndDatetime,
			})
			if err != nil {
				return fmt.Errorf("qtx.CountUnpaidTeachers(): %w", err)
			}
			return nil
		})
		if err != nil {
			return teaching.GetTeachersForPaymentResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
		}
		teachersForPayments = NewTeacherForPaymentsFromGetPaidTeachersRow(teacherRows)

	} else { // for unpaid teachers, we fetch data from `Attendance`
		var teacherRows = make([]mysql.GetUnpaidTeachersRow, 0)
		err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
			var err error
			teacherRows, err = qtx.GetUnpaidTeachers(newCtx, mysql.GetUnpaidTeachersParams{
				StartDate: spec.StartDatetime,
				EndDate:   spec.EndDatetime,
				Limit:     int32(limit),
				Offset:    int32(offset),
			})
			if err != nil {
				return fmt.Errorf("qtx.GetUnpaidTeachers(): %w", err)
			}

			totalResults, err = qtx.CountUnpaidTeachers(newCtx, mysql.CountUnpaidTeachersParams{
				StartDate: spec.StartDatetime,
				EndDate:   spec.EndDatetime,
			})
			if err != nil {
				return fmt.Errorf("qtx.CountUnpaidTeachers(): %w", err)
			}
			return nil
		})
		if err != nil {
			return teaching.GetTeachersForPaymentResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
		}
		teachersForPayments = NewTeacherForPaymentsFromGetUnpaidTeachersRow(teacherRows)
	}

	return teaching.GetTeachersForPaymentResult{
		TeachersForPayment: teachersForPayments,
		PaginationResult:   *util.NewPaginationResult(int(totalResults), spec.Pagination.ResultsPerPage, spec.Pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetTeacherPaymentInvoiceItems(ctx context.Context, spec teaching.GetTeacherPaymentInvoiceItemsSpec) ([]teaching.TeacherPaymentInvoiceItem, error) {
	getUnpaidAttendancesByTeacherIdSpec := entity.GetUnpaidAttendancesByTeacherIdSpec{
		TeacherID: spec.TeacherID,
		TimeSpec:  spec.TimeSpec,
	}

	attendances, err := s.entityService.GetUnpaidAttendancesByTeacherId(ctx, getUnpaidAttendancesByTeacherIdSpec)
	if err != nil {
		return []teaching.TeacherPaymentInvoiceItem{}, fmt.Errorf("entityService.GetUnpaidAttendancesByTeacherId(): %v", err)
	}

	tpiiBuilder := teaching.NewTeacherPaymentInvoiceItemBuilder()
	tpiiBuilder.AddAttendances(attendances)
	teacherPaymentInvoiceItems := tpiiBuilder.Build()

	return teacherPaymentInvoiceItems, nil
}

func (s teachingServiceImpl) GetExistingTeacherPaymentInvoiceItems(ctx context.Context, spec teaching.GetExistingTeacherPaymentInvoiceItemsSpec) ([]teaching.TeacherPaymentInvoiceItem, error) {
	getTeacherPaymentsByTeacherIdSpec := entity.GetTeacherPaymentsByTeacherIdSpec{
		TeacherID: spec.TeacherID,
		TimeSpec:  spec.TimeSpec,
	}

	teacherPayments, err := s.entityService.GetTeacherPaymentsByTeacherId(ctx, getTeacherPaymentsByTeacherIdSpec)
	if err != nil {
		return []teaching.TeacherPaymentInvoiceItem{}, fmt.Errorf("entityService.GetTeacherPaymentsByTeacherId(): %v", err)
	}

	tpiiBuilder := teaching.NewTeacherPaymentInvoiceItemBuilder()
	tpiiBuilder.AddTeacherPayments(teacherPayments)
	teacherPaymentInvoiceItems := tpiiBuilder.Build()

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

func (s teachingServiceImpl) ModifyTeacherPayments(ctx context.Context, specs []teaching.ModifyTeacherPaymentsSpec) (teaching.ModifyTeacherPaymentsResult, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountTeacherPaymentsByIds)
	if errV != nil {
		return teaching.ModifyTeacherPaymentsResult{}, errV
	}

	editedTeacherPaymentIDs := make([]entity.TeacherPaymentID, 0)

	editTeacherPaymentsSpecs := make([]teaching.EditTeacherPaymentsSpec, 0, len(specs)/2)
	removedTeacherPaymentIDs := make([]entity.TeacherPaymentID, 0, len(specs)/2)
	for _, spec := range specs {
		if !spec.IsDeleted {
			editTeacherPaymentsSpecs = append(editTeacherPaymentsSpecs, teaching.EditTeacherPaymentsSpec{
				TeacherPaymentID:      spec.TeacherPaymentID,
				PaidCourseFeeValue:    spec.PaidCourseFeeValue,
				PaidTransportFeeValue: spec.PaidTransportFeeValue,
			})
		} else {
			removedTeacherPaymentIDs = append(removedTeacherPaymentIDs, spec.TeacherPaymentID)
		}
	}

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		editedTeacherPaymentIDs, err = s.EditTeacherPayments(newCtx, editTeacherPaymentsSpecs)
		if err != nil {
			return fmt.Errorf("EditTeacherPayments(): %w", err)
		}

		err = s.RemoveTeacherPayments(newCtx, removedTeacherPaymentIDs)
		if err != nil {
			return fmt.Errorf("RemoveTeacherPayments(): %w", err)
		}

		return nil
	})
	if err != nil {
		return teaching.ModifyTeacherPaymentsResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return teaching.ModifyTeacherPaymentsResult{
		EditedTeacherPaymentIDs:  editedTeacherPaymentIDs,
		DeletedTeacherPaymentIDs: removedTeacherPaymentIDs,
	}, nil
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
