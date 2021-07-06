# Adyen POC 

Forked from https://github.com/adyen-examples/adyen-golang-online-payments

POC to demonstrate feasibility of swapping our Stripe integration for Adyen - feature parity: 

- Pay Now
- Pay Later Delayed Auth
- Pay Later
- Applying a charge to each Pay Later mode

## Usage

Currently just hacking the code directly - edit for desired flow.

https://docs.adyen.com/development-resources/test-cards/test-card-numbers#test-3d-secure-2-authentication

### Adyen account setup

 - Disable Risk - in Risk settings (make sure a merchant is selected)
 - Account - API URLS - Additional Data Settings - check following:
   - Cardholder name
   - Shopper country
   - Card summary
   - Expiry date
   - Variant
   - Token information for digital wallets (e.g ApplePay)
   - Recurring details
  
Recurring details is only mandatory setting - the others would need checking.

## Food for thought

- implementation would be a case of exposing a few endpoints - its not too complex imo.
- its a different flow - UI driven and not async like Stripe
- it feels like there is more room for error - Stripe abstracts a lot of this away. 3DS for example - v1 and v2, journeys that dont require 3DS need special handling. This will need thorough testing.
- any Stripe derived data must be persisted in API - profile, credit card? we now have a user-repository so maybe a sensible step anyway - even for Stripe
- we should rename our current em-payment service to em-payment-stripe and move to an event driven flow

## Stripe cryptogram expiry

Not sure we got correct info from Stripe on this: https://ennismore.atlassian.net/browse/DEV-1895

## Moving Stripe to a sync flow 

I am definitely edging to moving away from our current webhook flow.....

If we implemented idempotency so for a single booking id you always get the same stripe intent id - then the UI attempted a second payment (after already completing and then failing API call to confirm) Stripe will error:

```
{
  "error": {
    "code": "payment_intent_unexpected_state",
    "doc_url": "https://stripe.com/docs/error-codes/payment-intent-unexpected-state",
    "message": "You cannot confirm this PaymentIntent because it has already succeeded after being previously confirmed.",
    "payment_intent": {
      "id": "pi_1J7yjkIB01GFU0AGEDL0ZiMs",
      "object": "payment_intent",
      "amount": 65000,
      "canceled_at": null,
      "cancellation_reason": null,
      "capture_method": "automatic",
      "client_secret": "pi_1J7yjkIB01GFU0AGEDL0ZiMs_secret_RKaOaVlxVtLS83zBx4UxiYlij",
      "confirmation_method": "automatic",
      "created": 1625041536,
      "currency": "gbp",
      "description": null,
      "last_payment_error": null,
      "livemode": false,
      "next_action": null,
      "payment_method": "pm_1J7yzDIB01GFU0AGVYA2qN5c",
      "payment_method_types": [
        "card"
      ],
      "receipt_email": null,
      "setup_future_usage": null,
      "shipping": null,
      "source": null,
      "status": "succeeded"
    },
    "type": "invalid_request_error"
  }
}
```

UI could handle this appropriately - in this case just skipping and attempt the confirm again.

The benefit is the UI should get detailed error from confirm - so he could then handle this specifically too. Instead of polling for a generic PAYMENT_ERROR state. Tbh - this state is even wrong - it should be CONFIRM_ERROR or similar.
