package teaching

// CalculateSLTFeeQuarterFromEP calculates the fee of a single SLT (StudentLearningToken) quota, which is a quarter of the course price.
//
// The submitted values from EP (EnrollmentPayment) (which are fees & balanceTopUp) are sometimes 1/4, 2/4, 3/4, 4/4, or even 5/4.
// So to store in quarter, we need to divide the fee with the balanceTopUp.
//
// On invalid balanceTopUp value (<= 0), it'll fallback to use the OneCourseCycle (normally it's 4).
//
// Examples: let's assume course fee is 400k/month, one month is one course cycle, each cycle is 4x. These EnrollmentPaymen payment options are very common to happen:
//  1. 1/4 is 100k with 1 balanceTopUp
//  2. 2/4 is 200k with 2 balanceTopUp
//  3. 3/4 is 300k with 3 balanceTopUp
//  4. 4/4 is 400k with 4 balanceTopUp (most common use case)
//  5. 5/4 is 500k with 5 balanceTopUp
// Meanwhile all these different ways of EnrollmentPayment should top up the same SLT, as they still adhere to 400k/month.
func CalculateSLTFeeQuarterFromEP(fee int32, balanceTopUp int32) int32 {
	if balanceTopUp <= 0 {
		return fee / Default_OneCourseCycle
	}
	return fee / balanceTopUp
}
