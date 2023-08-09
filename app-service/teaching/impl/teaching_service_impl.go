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

func (s teachingServiceImpl) SearchEnrollmentPayments(ctx context.Context, timeFilter util.TimeSpec) ([]entity.EnrollmentPayment, error) {
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
		if err != nil && !errors.Is(err, sql.ErrNoRows) { // ignore 404 NotFound error
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
		unpaidQuota := slt.Quota * -1
		penaltyFeeValue = ((unpaidQuota-1)/teaching.Default_OneCourseCycle + 1) * teaching.Default_PenaltyFeeValue // penaly = penaltyValue (in Rupiah) per 1 course cycle
	}

	return teaching.StudentEnrollmentInvoice{
		BalanceTopUp:      teaching.Default_BalanceTopUp,
		PenaltyFeeValue:   penaltyFeeValue,
		CourseFeeValue:    courseFeeValue,
		TransportFeeValue: studentEnrollment.ClassInfo.TransportFee,
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
				ValuePenalty:        spec.PenaltyFeeValue,
			},
		})
		if err != nil {
			return fmt.Errorf("entityService.InsertEnrollmentPayments(): %w", err)
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
					Quota:               spec.BalanceTopUp,
					CourseFeeValue:      spec.CourseFeeValue,
					TransportFeeValue:   spec.TransportFeeValue,
				},
			})
			if err != nil {
				return fmt.Errorf("entityService.InsertStudentLearningTokens(): %w", err)
			}
		} else {
			err := qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
				Quota: spec.BalanceTopUp,
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

func (s teachingServiceImpl) EditEnrollmentPaymentBalance(ctx context.Context, spec teaching.EditStudentEnrollmentPaymentBalanceSpec) error {
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
		if err != nil {
			return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(): %w", err)
		}

		quotaChange := spec.BalanceTopUp - prevEP.BalanceTopUp
		err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
			Quota: quotaChange,
			ID:    updatedSLT.ID,
		})
		if err != nil {
			return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
		}

		err = qtx.UpdateEnrollmentPaymentBalance(newCtx, mysql.UpdateEnrollmentPaymentBalanceParams{
			BalanceTopUp: spec.BalanceTopUp,
			ID:           int64(spec.EnrollmentPaymentID),
		})
		if err != nil {
			return fmt.Errorf("entityService.UpdateEnrollmentPayments(): %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) RemoveEnrollmentPayment(ctx context.Context, enrollmentPaymentID entity.EnrollmentPaymentID) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		prevEP, err := qtx.GetEnrollmentPaymentById(newCtx, int64(enrollmentPaymentID))
		if err != nil {
			return fmt.Errorf("qtx.GetEnrollmentPaymentById(): %w", err)
		}

		updatedSLT, err := qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(newCtx, mysql.GetSLTByEnrollmentIdAndCourseFeeAndTransportFeeParams{
			EnrollmentID:      prevEP.StudentEnrollmentID,
			CourseFeeValue:    prevEP.CourseFeeValue,
			TransportFeeValue: prevEP.TransportFeeValue,
		})
		if err != nil {
			return fmt.Errorf("qtx.GetSLTByEnrollmentIdAndCourseFeeAndTransportFee(): %w", err)
		}

		quotaChange := -1 * prevEP.BalanceTopUp
		err = qtx.IncrementSLTQuotaById(newCtx, mysql.IncrementSLTQuotaByIdParams{
			Quota: quotaChange,
			ID:    updatedSLT.ID,
		})
		if err != nil {
			return fmt.Errorf("qtx.IncrementSLTQuotaById(): %w", err)
		}

		err = s.entityService.DeleteEnrollmentPayments(newCtx, []entity.EnrollmentPaymentID{
			entity.EnrollmentPaymentID(prevEP.StudentEnrollmentID),
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

func (s teachingServiceImpl) AddPresence(ctx context.Context, spec teaching.AddPresenceSpec) error {
	return nil
}
