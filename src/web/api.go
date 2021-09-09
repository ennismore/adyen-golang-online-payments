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

func handleError(method string, c *gin.Context, err error, httpRes *http.Response) {
	log.Printf("Error in %s: %s\n", method, err.Error())
	if httpRes != nil && httpRes.StatusCode >= 300 {
		c.JSON(httpRes.StatusCode, err.Error())
		return
	}
	c.JSON(http.StatusBadRequest, err.Error())
}
