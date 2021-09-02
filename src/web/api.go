package web

import (
  "fmt"
  "github.com/adyen/adyen-go-api-library/v5/src/checkout"
  "github.com/adyen/adyen-go-api-library/v5/src/common"
  "github.com/gin-gonic/gin"
  "log"
  "net/http"
  "net/url"
)

//func ChargeHandler(c *gin.Context) {
//  req := checkout.PaymentRequest{}
//
//  //todo: metadata from UI
//  if req.Metadata == nil {
//    req.Metadata = make(map[string]string)
//  }
//  //req.Metadata["PaymentMethod"] = "PAY_NOW" //todo
//  req.Metadata["bookingId"] = "e30f495844f2f9331201eb222cc8a8542f501830"
//  req.Metadata["specialAssistance"] = "true"
//  req.Metadata["marketingOptIn"] = "true"
//  req.Metadata["email"] = req.ShopperEmail
//
//  req.MerchantAccount = merchantAccount
//  req.ShopperReference = lastShopperReference
//  req.Reference = lastOrderRef
//  req.ReturnUrl = fmt.Sprintf("http://localhost:9000/api/handleShopperRedirect?orderRef=%s", lastOrderRef)
//  req.ShopperInteraction = "ContAuth"
//  req.RecurringProcessingModel = "CardOnFile"
//  req.Amount = checkout.Amount{
//  	Currency: "EUR",
//  	Value:    1000,
//  }
//  req.PaymentMethod = map[string]string {
//    "type": "scheme",
//    "storedPaymentMethodId": lastTokenisedPaymentMethodId,
//  }
//
//  log.Printf("Request for %s API::\n%+v\n", "Payments", req)
//  res, httpRes, err := client.Checkout.Payments(&req)
//  log.Printf("Response for %s API::\n%+v\n", "Payments", res)
//  log.Printf("HTTP Response for %s API::\n%+v\n", "Payments", httpRes)
//  if err != nil {
//    handleError("PaymentsHandler", c, err, httpRes)
//    return
//  }
//
//  c.JSON(http.StatusOK, res)
//  return
//}
//
//func DelayedAuthChargeHandler(c *gin.Context) {
//
//  log.Printf("Threeds2ChallengeResult ::: %s\n", last3DSAutenticationValue)
//  log.Printf("PaymentData ::: %s\n", lastPaymentData)
//
//  d := checkout.PaymentCompletionDetails{}
//  d.Threeds2ChallengeResult = last3DSAutenticationValue
//
//  req := checkout.DetailsRequest{}
//  req.ThreeDSAuthenticationOnly = false //todo this didnt work - fork lib?
//  req.PaymentData = lastPaymentData
//  req.Details = d
//
//  log.Printf("Request for %s API::\n%+v\n", "PaymentDetails", req)
//  res, httpRes, err := client.Checkout.PaymentsDetails(&req)
//  log.Printf("Response for %s API::\n%+v\n", "PaymentDetails", res)
//  log.Printf("HTTP Response for %s API::\n%+v\n", "PaymentDetails", httpRes)
//  if err != nil {
//    handleError("PaymentDetailsHandler", c, err, httpRes)
//    return
//  }
//
//  c.JSON(http.StatusOK, res)
//
//  return
//}

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
