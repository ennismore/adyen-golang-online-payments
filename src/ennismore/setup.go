package ennismore

import (
	"context"
	"errors"
	"github.com/ennismore/em-domain/v2/common"
	"github.com/ennismore/em-domain/v2/service/config"
	"log"
)

type FindHotelContext func(hotelReferenceId string) (*common.HotelContext, error)

func GetHotelConfigMap(emConfigClient config.ConfigServiceClient) (map[string]*config.Hotel, error) {
	log.Printf("Get All Hotel Contexts")

	request := &config.AllHotelConfigRequest{Locale: &common.Locale{Lang: common.Locale_EN}}

	singleHotelConfig, err := emConfigClient.GetAllHotelConfig(context.Background(), request)
	if err != nil {
		log.Printf("failed to load all hotel config: %s", err.Error())
		return nil, err
	}

	hotels := make(map[string]*config.Hotel)
	for _, hotel := range singleHotelConfig.Hotels {
		hotels[hotel.ReferenceId] = hotel
	}

	return hotels, nil
}

func FindHotelContextInit(hotels map[string]*config.Hotel) func(hotelCode string) (*common.HotelContext, error) {
	return func(hotelCode string) (*common.HotelContext, error) {
		for k, v := range hotels {
			if hotelCode == k {
				return v.HotelContext, nil
			}
		}
		log.Printf("not found valid hotel context [%s]", hotelCode)
		return nil, errors.New("not found valid hotel context")
	}
}
