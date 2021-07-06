package ennismore

import (
	"fmt"
	"github.com/ennismore/em-domain/v2/common"
	pb "github.com/ennismore/em-domain/v2/common/booking"
	"github.com/ennismore/em-domain/v2/common/payment"
	"github.com/ennismore/em-domain/v2/service/api/booking"
	"log"
	"time"
)

const dateFormat = "2006-01-02"

type Booking struct {
	OperaId                  string
	CurrentStatus            pb.BookingStatus
	Arrival, Departure       time.Time
	GrandTotal, DepositTotal *common.Amount
	DepositRequired          bool
	Locale                   *common.Locale
}

func NewBookingFromSummary(s *booking.BookingSummaryResponse) (*Booking, error) {
	//TODO issues with hardcoding array indexes?

	if len(s.RoomStay) == 0 {
		return nil, fmt.Errorf("booking has no rooms")
	}
	firstRoomStay := s.RoomStay[0] //this assumes multi room bookings always share same core booking details i.e. dates

	if len(s.TotalChargeBreakdown) == 0 {
		return nil, fmt.Errorf("booking has no total charge breakdown")
	}

	var depositRequired bool
	deposit := &common.Amount{
		Value:        0,
		Decimal:      s.TotalChargeBreakdown[0].GrandTotal.Decimal,
		CurrencyCode: s.TotalChargeBreakdown[0].GrandTotal.CurrencyCode,
	}
	if s.TotalChargeBreakdown[0].Deposit != nil {
		deposit = s.TotalChargeBreakdown[0].Deposit
		depositRequired = true
	}

	arrivalTime, err := time.Parse(dateFormat, firstRoomStay.From)
	if err != nil {
		return nil, err
	}

	departureTime, err := time.Parse(dateFormat, firstRoomStay.To)
	if err != nil {
		return nil, err
	}

	return &Booking{
		OperaId:         firstRoomStay.OperaId,
		DepositTotal:    deposit,
		Arrival:         arrivalTime,
		GrandTotal:      s.TotalChargeBreakdown[0].GrandTotal,
		Departure:       departureTime,
		DepositRequired: depositRequired,
	}, nil
}

func (b *Booking) IsTemporaryStatus() bool {
	return b.CurrentStatus == pb.BookingStatus_CREATED
}

func (b *Booking) GetRequiredPaymentMethod(liabilityExpiry time.Duration) payment.PaymentMethod {
	if b.DepositRequired {
		return payment.PaymentMethod_PAY_NOW
	} else {
		expiry := time.Now().Add(liabilityExpiry) //todo timezone dependent on Hotel location!
		isCheckoutAfterLiabilityExpiry := b.Departure.After(expiry)
		if isCheckoutAfterLiabilityExpiry {
			//When the customer check-out date is more than 90 days later, use the set up future payments flow to create a mandate to charge the customer off-session later. This comes without the guarantee of funds, or a liability shift.
			return payment.PaymentMethod_PAY_LATER
		} else {
			//When the customer check-out date is within 90 days, use the delayed card authorisation flow. This does not hold any amount on the customer's payment instrument, so comes without the guarantee of the funds when you do make the authorisation, but does come with a liability shift assuming the issuer accepts the 3D Secure cryptogram.
			return payment.PaymentMethod_PAY_LATER_DELAYED_AUTH
		}
	}
}

func (b *Booking) SanitiseDepositAmount() {
	//single use codes: https://github.com/ennismore/em-rate-service/blob/develop/migrations/3_single_use_code.up.sql
	//Opera will have zero total with no deposit - Stripe needs a postive value
	if b.DepositRequired == false && b.DepositTotal != nil && b.DepositTotal.Value == 0 {
		nonZeroDeposit := int32(100)
		if b.GrandTotal != nil && b.GrandTotal.Value > 0 {
			//use grand total if > 0
			nonZeroDeposit = b.GrandTotal.Value
		}
		log.Printf("booking [%s] deposit has ZERO amount, set it to %d\n", b.OperaId, nonZeroDeposit)
		b.DepositTotal.Value = nonZeroDeposit
	}
}
