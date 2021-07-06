package ennismore

import (
	"context"
	"fmt"
	"github.com/ennismore/em-domain/v2/common"
	pb "github.com/ennismore/em-domain/v2/common/booking"
	"github.com/ennismore/em-domain/v2/common/payment"
	userPb "github.com/ennismore/em-domain/v2/common/user"
	countryPb "github.com/ennismore/em-domain/v2/country"
	"github.com/ennismore/em-domain/v2/service/api/booking"
	"github.com/ennismore/em-domain/v2/service/bookingrepository"
	emConfig "github.com/ennismore/em-domain/v2/service/config"
	"github.com/ennismore/em-domain/v2/service/id"
	"github.com/ennismore/em-domain/v2/service/userrepository"
	statePb "github.com/ennismore/em-domain/v2/state"
	"log"
	"strconv"
	"strings"
)

type PaymentMethod struct {
	Name, Phone, Email                                         string
	AddressLine1, AddressLine2, City, State, PostCode, Country string
	PaymentType                                                string
	CardLast4, CardBrand                                       string
	CardExpMonth, CardExpYear                                  int32
}

type ConfirmBookingRequest struct {
	BookingMetadata    *payment.PaymentMetaData
	PaymentTransaction *payment.PaymentTransaction
	PaymentMethod      PaymentMethod
}

type ConfirmBookingResponse struct {
	OperaId              int32
	HotelContext         *common.HotelContext
	User                 *userPb.User
	SkipConfirm          bool
	CurrentBookingStatus pb.BookingStatus
}

func ConfirmBooking(grpcClients *Clients, req *ConfirmBookingRequest, fromIdToContext FindHotelContext) (*ConfirmBookingResponse, error) {
	idResp, err := grpcClients.IdObfuscatorClient.Get(context.Background(), &id.GetRequest{ExternalId: req.BookingMetadata.BookingId})
	if err != nil {
		return nil, fmt.Errorf("cannot get HotelContext from ID %v", err)
	}

	hotelContext, err := fromIdToContext(idResp.HotelReferenceId)
	if err != nil {
		return nil, fmt.Errorf("cannot get HotelContext from ID %v", err)
	}

	//get booking - check status is XX
	summaryResponse, err := grpcClients.BookingSummary(context.Background(), &booking.BookingSummaryRequest{
		HotelContext: hotelContext,
		BookingId:    strconv.Itoa(int(idResp.OperaId)),
	})
	if err != nil {
		return &ConfirmBookingResponse{HotelContext: hotelContext, OperaId: idResp.OperaId}, fmt.Errorf("booking summary error %v", err)
	}

	summary, err := NewBookingFromSummary(summaryResponse)
	if err != nil {
		return &ConfirmBookingResponse{HotelContext: hotelContext, OperaId: idResp.OperaId}, fmt.Errorf("booking malformed %v", err)
	}

	repoResponse, err := grpcClients.BookingRepositoryClient.Get(context.Background(), &bookingrepository.GetRequest{
		OperaId:      summary.OperaId,
		HotelContext: hotelContext,
	})
	if err != nil {
		log.Printf("cannot get booking summary %v", err)
		return nil, fmt.Errorf("cannot get booking summary %v", err)
	}
	summary.Locale = repoResponse.Locale
	summary.CurrentStatus = repoResponse.Status

	if !summary.IsTemporaryStatus() {
		return &ConfirmBookingResponse{HotelContext: hotelContext, OperaId: idResp.OperaId, CurrentBookingStatus: summary.CurrentStatus}, fmt.Errorf("booking not temporary status %v", summary.CurrentStatus)
	}

	//confirm booking
	resp := &ConfirmBookingResponse{
		OperaId:      idResp.OperaId,
		HotelContext: hotelContext,
	}

	title := common.Profile_TITLE_NOT_SET
	val, ok := common.Profile_Title_value[req.BookingMetadata.Title]
	if ok {
		title = common.Profile_Title(val)
	}

	config, err := grpcClients.ConfigServiceClient.GetHotelConfig(context.Background(), &emConfig.HotelConfigRequest{
		Context: hotelContext,
	})
	if err != nil {
		return resp, fmt.Errorf("cannot load hotel config %v", err)
	}

	// if country USA, then add STATE
	var state statePb.UsaState //this will send as nil
	var country = countryPb.Country_UK
	countryId, ok := countryPb.Country_value[req.PaymentMethod.Country]
	if ok {
		country = countryPb.Country(countryId)
		if country == countryPb.Country_US {
			stateId, ok := statePb.UsaState_value[req.PaymentMethod.State]
			if ok {
				state = statePb.UsaState(stateId)
			} else {
				log.Printf("state id not valid %s", req.PaymentMethod.State)
			}
		}
	} else {
		log.Printf("country id not valid %s", req.PaymentMethod.Country)
	}

	//get user
	firstName, lastName := getName(req.PaymentMethod.Name)
	profile := &common.Profile{
		Title:     title,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     req.PaymentMethod.Phone,
		Email:     req.PaymentMethod.Email,
		Address: &common.Address{
			AddressLine: []string{
				req.PaymentMethod.AddressLine1,
				req.PaymentMethod.AddressLine2,
			},
			City:     req.PaymentMethod.City,
			UsaState: state,
			PostCode: req.PaymentMethod.PostCode,
			Country:  country,
		},
		MarketingOptIn: req.BookingMetadata.MarketingOptIn,
		Locale:         summary.Locale,
	}

	user, err := grpcClients.GetPmsProfileId(context.Background(), &userrepository.GetPmsProfileIdRequest{
		HotelContext: hotelContext,
		PmsId:        config.Hotels.Pms.PmsId,
		Profile:      profile,
	})
	if err != nil {
		return resp, fmt.Errorf("cannot get user account %v", err)
	}

	resp.User = user.User
	profile.Id = user.PmsProfileId.ProfileId

	_, err = grpcClients.ConfirmBooking(context.Background(), &booking.ConfirmBookingRequest{
		HotelContext:       config.Hotels.HotelContext,
		BookingId:          strconv.Itoa(int(idResp.OperaId)),
		Comment:            req.BookingMetadata.Comment,
		Charity:            req.BookingMetadata.Charity,
		FlexyTime:          req.BookingMetadata.FlexyTime,
		Dog:                req.BookingMetadata.Dog,
		Cot:                req.BookingMetadata.Cot,
		SpecialAssistance:  req.BookingMetadata.SpecialAssistance,
		Profile:            profile,
		PaymentTransaction: req.PaymentTransaction,
	})
	if err != nil {
		return resp, fmt.Errorf("cannot confirm booking %v", err)
	}

	return resp, nil
}

const DefaultLastName = "******"

func getName(name string) (string, string) {
	//Opera needs both - Apple Pay sometimes only provides one name
	nameSplitBySpace := strings.Split(strings.Trim(name, " "), " ")
	firstName := nameSplitBySpace[0]
	lastName := DefaultLastName
	if len(nameSplitBySpace) > 1 {
		lastName = strings.Join(nameSplitBySpace[1:], " ")
	}
	return firstName, lastName
}
