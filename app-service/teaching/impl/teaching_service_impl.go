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
	"sonamusica-backend/config"
	"sonamusica-backend/logging"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("TeachingService", logging.GetLevel(configObject.LogLevel))
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

func (s teachingServiceImpl) CalculateStudentEnrollmentInvoice(ctx context.Context, studentEnrollmentID entity.StudentEnrollmentID) (teaching.StudentEnrollmentInvoice, error) {
	studentEnrollment, err := s.entityService.GetStudentEnrollmentById(ctx, studentEnrollmentID)
	if err != nil {
		return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("entityService.GetStudentEnrollmentById(): %w", err)
	}

	fmt.Printf("%+v\n", studentEnrollment)

	// calculate Course Fee
	courseFeeValue := studentEnrollment.ClassInfo.Course.DefaultFee
	teacherID, err := s.mySQLQueries.GetClassTeacherId(ctx, int64(studentEnrollment.ClassInfo.ClassID))
	if err != nil {
		return teaching.StudentEnrollmentInvoice{}, fmt.Errorf("mySQLQueries.GetClassTeacherId(): %w", err)
	}
	if teacherID.Valid {
		teacherSpecialFee, err := s.mySQLQueries.GetTeacherSpecialFeesByTeacherIdAndCourseId(ctx, mysql.GetTeacherSpecialFeesByTeacherIdAndCourseIdParams{
			TeacherID: 0,
			CourseID:  int64(studentEnrollmentID),
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

func (s teachingServiceImpl) SubmitStudentEnrollmentPayment(ctx context.Context, spec teaching.SubmitStudentEnrollmentPaymentSpec) error {
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		paidValue := spec.CourseFeeValue + spec.TransportFeeValue
		_, err := s.entityService.InsertEnrollmentPayments(newCtx, []entity.InsertEnrollmentPaymentSpec{
			{
				StudentEnrollmentID: spec.StudentEnrollmentID,
				PaymentDate:         spec.PaymentDate,
				BalanceTopUp:        spec.BalanceTopUp,
				Value:               paidValue,
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

func (s teachingServiceImpl) AddPresence(ctx context.Context, spec teaching.AddPresenceSpec) error {
	return nil
}
