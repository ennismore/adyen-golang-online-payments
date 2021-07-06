package ennismore

import (
	"github.com/ennismore/awssdk-go/awssdk"
	"github.com/ennismore/em-domain/v2/common"
	"github.com/ennismore/em-domain/v2/common/booking"
	"github.com/ennismore/em-domain/v2/common/event"
	"github.com/ennismore/em-domain/v2/common/user"
	"log"
	"strconv"
	"strings"
)

func FireAndForget(q awssdk.SimpleQueue, hotelContext *common.HotelContext, user *user.User, operaId int32, paymentMetadata map[string]string, rawPaymentMethodValue, paymentBrand string, action event.BookingEvent_Action) {
	if paymentMetadata != nil {
		//will save raw value in booking repo payment_metadata
		paymentMetadata["payment_method"] = rawPaymentMethodValue
	}

	paymentMethod := booking.BookingPayment_NONE
	if rawPaymentMethodValue != "" {
		paymentMethod = booking.BookingPayment_OTHER
		v, ok := booking.BookingPayment_Method_value[strings.ToUpper(rawPaymentMethodValue)]
		if ok {
			paymentMethod = booking.BookingPayment_Method(v)
		} else {
			log.Printf("cannot map payment method [%s] to a BookingPayment\n", rawPaymentMethodValue)
		}
	}

	e := &event.BookingEvent{
		Booking: &event.BookingEvent_Booking{
			BookingAction: action,
			BookingId:     strconv.Itoa(int(operaId)),
			HotelContext:  hotelContext,
			User:          user,
			Payment: &booking.BookingPayment{
				Processor: common.PaymentProcessor_ADYEN,
				Metadata:  paymentMetadata,
				Method:    paymentMethod,
				Brand:     strings.ToUpper(paymentBrand),
			},
		},
	}

	_, err := q.Write(awssdk.ProtoBufEventProducerHandler(e))
	if err != nil {
		log.Printf("cannot generate booking event [%s]\n", err.Error())
	}
}
