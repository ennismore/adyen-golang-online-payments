package web

import (
  "fmt"
  "github.com/adyen/adyen-go-api-library/v5/src/checkout"
  "github.com/adyen/adyen-go-api-library/v5/src/common"
  //commonPb "github.com/ennismore/em-domain/v2/common"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
  "log"
  "net/http"
  "net/url"
)

var (
	lastShopperReference = "" //this is VERY concurrent friendly
	lastOrderRef = ""
	lastTokenisedPaymentMethodId = ""
	lastPaymentData = ""
	last3DSAutenticationValue = ""
)

func ConfirmHandler(c *gin.Context) {

	var req checkout.PaymentMethodsRequest
	req.MerchantAccount = merchantAccount
	req.ShopperReference = lastShopperReference

  log.Printf("Request for %s API::\n%+v\n", "ConfirmHandler", req)
	res, httpRes, err := client.Checkout.PaymentMethods(&req)
	if err != nil {
		handleError("ConfirmHandler", c, err, httpRes)
		return
	}
	//&[{Brand:mc ExpiryMonth:03 ExpiryYear:2030 HolderName:M RYAN Iban: Id:8416248835802593 LastFour:5454 Name:MasterCard OwnerName: ShopperEmail: SupportedShopperInteractions:[Ecommerce ContAuth] Type:scheme}]
	log.Printf("StoredPaymentMethods ::\n%+v\n",
		res,
	)

  returnCode := common.Received
  details := checkout.PaymentDetailsResponse{
    ResultCode: &returnCode,
  }

	c.JSON(http.StatusOK, details)
	return

	//var req checkout.PaymentMethodsRequest
	//  req.MerchantAccount = merchantAccount
	//  //
	//  //req.ShopperReference = lastShopperReference
	//
	//  res, httpRes, err := client.Checkout.PaymentMethods(&req)
	//  if err != nil {
	//    handleError("ConfirmHandler", c, err, httpRes)
	//    return
	//  }
	//  //&[{Brand:mc ExpiryMonth:03 ExpiryYear:2030 HolderName:M RYAN Iban: Id:8416248835802593 LastFour:5454 Name:MasterCard OwnerName: ShopperEmail: SupportedShopperInteractions:[Ecommerce ContAuth] Type:scheme}]
	//  log.Printf("StoredPaymentMethods ::\n%+v\n",
	//   res,
	//  )
	//if res.StoredPaymentMethods != nil {
	//  storedPayment := (*res.StoredPaymentMethods)[0]
	//
	//  authCode := (*n.AdditionalData)["authCode"].(string)
	//  charge := &payment.PaymentTransaction_Charge{
	//    TransactionId: authCode,
	//    Amount: &commonPb.Amount{
	//      Value:        int32(n.Amount.Value),
	//      Decimal:      2,
	//      CurrencyCode: n.Amount.Currency,
	//    },
	//  }
	//
	//  expMth, _ := strconv.Atoi(storedPayment.ExpiryMonth) //why would you pass a string here!
	//  expYr, _ := strconv.Atoi(storedPayment.ExpiryYear)   //why would you pass a string here!
	//  paymentTx := &payment.PaymentTransaction{
	//    Id: storedPayment.Id, //todo not sure - PaymentIntentId or SetupIntentId for Stripe
	//    Card: &payment.PaymentTransaction_CreditCard{
	//      Source:          "credit_card",
	//      Last4:           storedPayment.LastFour,
	//      ExpirationMonth: int32(expMth),
	//      ExpirationYear:  int32(expYr),
	//      Brand:           storedPayment.Brand,
	//    },
	//    Charge: charge,
	//  }
	//
	//  paymentMethod := ennismore.PaymentMethod{
	//    Name:         "Mark Ryan",
	//    Phone:        "0123456789",
	//    Email:        "mryan321+adyen@gmail.com",
	//    AddressLine1: "123 The Street",
	//    City:         "The City",
	//    State:        "The State",
	//    PostCode:     "AB1 2CD",
	//    Country:      "GB",
	//  }
	//
	//  metaSpecialAssistance, _ := strconv.ParseBool((*n.AdditionalData)["metadata.specialAssistance"].(string))
	//  metaMarketingOptin, _ := strconv.ParseBool((*n.AdditionalData)["metadata.marketingOptIn"].(string))
	//  confirmBookingReq := &ennismore.ConfirmBookingRequest{
	//    BookingMetadata: &payment.PaymentMetaData{
	//      BookingId:         (*n.AdditionalData)["metadata.bookingId"].(string),
	//      SpecialAssistance: metaSpecialAssistance,
	//      MarketingOptIn:    metaMarketingOptin,
	//    },
	//    PaymentTransaction: paymentTx,
	//    PaymentMethod:      paymentMethod,
	//  }
	//
	//  resp, err := ennismore.ConfirmBooking(grpcClients, confirmBookingReq, referenceIdMapper)
	//  if err != nil {
	//    log.Printf("Error confirming ::\n%+v\n", err)
	//  } else {
	//    log.Printf("Successfully confirmed ::\n%+v\n", resp.OperaId)
	//  }
	//}
}

func WebhookHandler(c *gin.Context) {
	log.Println("Webhook received")
	body, err := c.GetRawData()
	if err != nil {
		handleError("WebhookHandler", c, err, nil)
		return
	}
	notification, err := client.Notification.HandleNotificationRequest(string(body))
	if err != nil {
		handleError("WebhookHandler", c, err, nil)
		return
	}
	log.Printf("Notification recieved::\n%+v\n%+v\n%+v\n%+v\n%+v\n%+v\n%+v\n",
		notification.GetNotificationItems()[0].EventCode,
		notification.GetNotificationItems()[0].Success,
		notification.GetNotificationItems()[0].AdditionalData, //contains metadata
		notification.GetNotificationItems()[0].PspReference,
		notification.GetNotificationItems()[0].MerchantReference,
		notification.GetNotificationItems()[0].OriginalReference,
		notification.GetNotificationItems()[0].Reason,
	)

	if notification.GetNotificationItems()[0].EventCode == "AUTHORISATION" {
		//confirm here??? not received on authentication_only
		log.Printf("AUTHORISATION received!!!!")
		//confirmBooking(c, notification.GetNotificationItems()[0])
	}
	c.JSON(http.StatusOK, "[accepted]")
	return
}

// PaymentMethodsHandler retrieves a list of available payment methods from Adyen API
func PaymentMethodsHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var req checkout.PaymentMethodsRequest

	if err := c.BindJSON(&req); err != nil {
		handleError("PaymentMethodsHandler", c, err, nil)
		return
	}
	req.MerchantAccount = merchantAccount
	req.Channel = "Web"
	log.Printf("Request for %s API::\n%+v\n", "PaymentMethods", req)
	res, httpRes, err := client.Checkout.PaymentMethods(&req)
	if err != nil {
		handleError("PaymentMethodsHandler", c, err, httpRes)
		return
	}
	c.JSON(http.StatusOK, res)
	return
}

// PaymentsHandler makes payment using Adyen API
func PaymentsHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var req checkout.PaymentRequest

	if err := c.BindJSON(&req); err != nil {
		handleError("PaymentsHandler", c, err, nil)
		return
	}

	//todo: metadata from UI
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	//req.Metadata["PaymentMethod"] = "PAY_NOW" //todo
	req.Metadata["bookingId"] = "e30f495844f2f9331201eb222cc8a8542f501830"
	req.Metadata["specialAssistance"] = "true"
	req.Metadata["marketingOptIn"] = "true"
	req.Metadata["email"] = req.ShopperEmail

	//todo add a new status - PAYMENT_PENDING - attach payment response to booking - any relevant ids
	//todo fireBookingUpdate event? Or sync update might be better...

	req.MerchantAccount = merchantAccount   // required
	orderRef := uuid.Must(uuid.NewRandom()) //todo our booking id?
	req.Reference = orderRef.String()       // required
  lastOrderRef = orderRef.String()
	req.Channel = "Web"                     // required
	req.Origin = "http://localhost:9000"    // required for 3ds2 native flow
	req.ShopperIP = c.ClientIP()            // required by some issuers for 3ds2
	// required for 3ds2 redirect flow
	req.ReturnUrl = fmt.Sprintf("http://localhost:9000/api/handleShopperRedirect?orderRef=%s", orderRef)

	//tokenise
	req.StorePaymentMethod = true
	req.RecurringProcessingModel = "CardOnFile"
	userUuid := uuid.Must(uuid.NewRandom()) //todo our user uuid?
	req.ShopperReference = userUuid.String()
	lastShopperReference = userUuid.String()
  req.AdditionalData = map[string]string{
    // required for 3ds2 native flow
    "allow3DS2": "true",
  }
  pmType := getPaymentType(req.PaymentMethod)

	//todo decide pay journey and handle accordingly... some UI content depends on this....
	//can pass back:
	//PAY_NOW,
  //req.ShopperInteraction = "Ecommerce"
	//req.Amount = checkout.Amount{
	//	Currency: findCurrency(pmType),
	//	Value:    1000, // value is 10€ in minor units
	//}

	//PAY_LATER_DELAYED_AUTH (90 days mmmmmm: https://docs.adyen.com/online-payments/3d-secure/other-3ds-flows/authentication-only#authentication-data-expiry)
  req.ShopperInteraction = "ContAuth"
  req.Amount = checkout.Amount{
    Currency: findCurrency(pmType),
    Value:    1000, // value is 10€ in minor units
  }
	req.AdditionalData["executeThreeD"] = "true"
	req.ThreeDSAuthenticationOnly = true

	//PAY_LATER
	//zeroauth
  //req.ShopperInteraction = "ContAuth"
  //req.Amount = checkout.Amount{
  // Currency: findCurrency(pmType),
  // Value:    0,
  //}

	log.Printf("Request for %s API::\n%+v\n", "Payments", req)
	res, httpRes, err := client.Checkout.Payments(&req)
	log.Printf("Response for %s API::\n%+v\n", "Payments", res)
	log.Printf("HTTP Response for %s API::\n%+v\n", "Payments", httpRes)
	if err != nil {
		handleError("PaymentsHandler", c, err, httpRes)
		return
	}

  lastPaymentData = res.Action.(*checkout.CheckoutThreeDS2Action).PaymentData

	c.JSON(http.StatusOK, res)
	return
}

func ChargeHandler(c *gin.Context) {
  req := checkout.PaymentRequest{}

  //todo: metadata from UI
  if req.Metadata == nil {
    req.Metadata = make(map[string]string)
  }
  //req.Metadata["PaymentMethod"] = "PAY_NOW" //todo
  req.Metadata["bookingId"] = "e30f495844f2f9331201eb222cc8a8542f501830"
  req.Metadata["specialAssistance"] = "true"
  req.Metadata["marketingOptIn"] = "true"
  req.Metadata["email"] = req.ShopperEmail

  req.MerchantAccount = merchantAccount
  req.ShopperReference = lastShopperReference
  req.Reference = lastOrderRef
  req.ReturnUrl = fmt.Sprintf("http://localhost:9000/api/handleShopperRedirect?orderRef=%s", lastOrderRef)
  req.ShopperInteraction = "ContAuth"
  req.RecurringProcessingModel = "CardOnFile"
  req.Amount = checkout.Amount{
  	Currency: "EUR",
  	Value:    1000,
  }
  req.PaymentMethod = map[string]string {
    "type": "scheme",
    "storedPaymentMethodId": lastTokenisedPaymentMethodId,
    //"encryptedSecurityCode": "adyenjs_0_1_18$MT6ppy0FAMVMLH..."
  }

  log.Printf("Request for %s API::\n%+v\n", "Payments", req)
  res, httpRes, err := client.Checkout.Payments(&req)
  log.Printf("Response for %s API::\n%+v\n", "Payments", res)
  log.Printf("HTTP Response for %s API::\n%+v\n", "Payments", httpRes)
  if err != nil {
    handleError("PaymentsHandler", c, err, httpRes)
    return
  }

  c.JSON(http.StatusOK, res)
  return
}

func DelayedAuthChargeHandler(c *gin.Context) {

  d := checkout.PaymentCompletionDetails{}
  d.Threeds2ChallengeResult = last3DSAutenticationValue

  req := checkout.DetailsRequest{}
  req.ThreeDSAuthenticationOnly = false
  req.PaymentData = lastPaymentData
  req.Details = d

  log.Printf("Request for %s API::\n%+v\n", "PaymentDetails", req)
  res, httpRes, err := client.Checkout.PaymentsDetails(&req)
  log.Printf("Response for %s API::\n%+v\n", "PaymentDetails", res)
  log.Printf("HTTP Response for %s API::\n%+v\n", "PaymentDetails", httpRes)
  if err != nil {
    handleError("PaymentDetailsHandler", c, err, httpRes)
    return
  }

  c.JSON(http.StatusOK, res)

  return
}

// PaymentDetailsHandler gets payment details using Adyen API
func PaymentDetailsHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var req checkout.DetailsRequest

	if err := c.BindJSON(&req); err != nil {
		handleError("PaymentDetailsHandler", c, err, nil)
		return
	}
	log.Printf("Request for %s API::\n%+v\n", "PaymentDetails", req)
	res, httpRes, err := client.Checkout.PaymentsDetails(&req)
	log.Printf("Response for %s API::\n%+v\n", "PaymentDetails", res)
	log.Printf("HTTP Response for %s API::\n%+v\n", "PaymentDetails", httpRes)
	if err != nil {
		handleError("PaymentDetailsHandler", c, err, httpRes)
		return
	}

	//confirm here???? not in webhook...

  lastTokenisedPaymentMethodId = res.AdditionalData["recurring.recurringDetailReference"]
  last3DSAutenticationValue = req.Details.ThreeDSResult

	c.JSON(http.StatusOK, res)

	return
}

// RedirectHandler handles POST and GET redirects from Adyen API
func RedirectHandler(c *gin.Context) {
	log.Println("Redirect received")
	var details checkout.PaymentCompletionDetails

	if err := c.ShouldBind(&details); err != nil {
		handleError("RedirectHandler", c, err, nil)
		return
	}

	details.RedirectResult = c.Query("redirectResult")
	details.Payload = c.Query("payload")

	req := checkout.DetailsRequest{Details: details}

	log.Printf("Request for %s API::\n%+v\n", "PaymentDetails", req)
	res, httpRes, err := client.Checkout.PaymentsDetails(&req)
	log.Printf("HTTP Response for %s API::\n%+v\n", "PaymentDetails", httpRes)
	if err != nil {
		handleError("RedirectHandler", c, err, httpRes)
		return
	}
	log.Printf("Response for %s API::\n%+v\n", "PaymentDetails", res)

	if res.PspReference != "" {
		var redirectURL string
		// Conditionally handle different result codes for the shopper
		switch *res.ResultCode {
		case common.Authorised:
			redirectURL = "/result/success"
			break
		case common.Pending:
		case common.Received:
			redirectURL = "/result/pending"
			break
		case common.Refused:
			redirectURL = "/result/failed"
			break
		default:
			{
				reason := res.RefusalReason
				if reason == "" {
					reason = res.ResultCode.String()
				}
				redirectURL = fmt.Sprintf("/result/error?reason=%s", url.QueryEscape(reason))
				break
			}
		}
		c.Redirect(
			http.StatusFound,
			redirectURL,
		)
		return
	}
	c.JSON(httpRes.StatusCode, httpRes.Status)
	return
}

/* Utils */

func findCurrency(typ string) string {
	switch typ {
	case "ach":
		return "USD"
	case "wechatpayqr":
	case "alipay":
		return "CNY"
	case "dotpay":
		return "PLN"
	case "boletobancario":
	case "boletobancario_santander":
		return "BRL"
	default:
		return "EUR"
	}
	return ""
}

func getPaymentType(pm interface{}) string {
	switch v := pm.(type) {
	case *checkout.CardDetails:
		return v.Type
	case *checkout.IdealDetails:
		return v.Type
	case *checkout.DotpayDetails:
		return v.Type
	case *checkout.GiropayDetails:
		return v.Type
	case *checkout.AchDetails:
		return v.Type
	case *checkout.KlarnaDetails:
		return v.Type
	case map[string]interface{}:
		return v["type"].(string)
	}
	return ""
}

func handleError(method string, c *gin.Context, err error, httpRes *http.Response) {
	log.Printf("Error in %s: %s\n", method, err.Error())
	if httpRes != nil && httpRes.StatusCode >= 300 {
		c.JSON(httpRes.StatusCode, err.Error())
		return
	}
	c.JSON(http.StatusBadRequest, err.Error())
}
