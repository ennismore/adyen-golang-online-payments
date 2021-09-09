const clientKey = document.getElementById("clientKey").innerHTML;
const type = document.getElementById("type").innerHTML;

const booking_id = "823ba3da78e5324e77f7a96259218bf788df3930";

async function initCheckout() {
  try {
    //this can be cached....according to docs
    const paymentMethodsResponse = await callBookingEngineApi("/payments/adyen", {});
    var adyenResponse = JSON.parse(atob(paymentMethodsResponse.raw))
    const configuration = {
      paymentMethodsResponse: filterUnimplemented(adyenResponse),
      clientKey,
      locale: "en_US",
      environment: "test",
      showPayButton: true,
      paymentMethodsConfiguration: {
        ideal: {
          showImage: true,
        },
        card: {
          hasHolderName: true,
          holderNameRequired: true,
          name: "Credit or debit card",
          amount: {
            value: 1000,
            currency: "EUR",
          },
        },
      },
      onSubmit: (state, component) => {
        if (state.isValid) {
          console.log("onSubmit");
          payload = {
            metadata: {
              //https://github.com/ennismore/em-domain/blob/65c4ceaecca0e894106534d6cb1c62c35a387e65/clients/typescript-fetch/api.ts#L2070
              booking_id: booking_id,
              comment: "special requests...",
              charity: 3,
              flexy_time: { check_in: "HOUR_0_TO_1", check_out: "HOUR_17_TO_18" },
              dog: false,
              cot: "COT",
              special_assistance: true,
              title: "MR",
              email: "mryan321+adyen@gmail.com",
              marketing_opt_in: true,
              first_name: "Mark",
              last_name: "Testing",
              address_line_1: "123 The Street",
              city: "The City",
              state: "The State",
              country: "GB",
              post_code: "AB1 2CD",
              phone: "0123456789"
            },
            raw: state.data
          }
          handleBookingEngineSubmission(payload, component, "/payments/adyen/new");
        }
      },
      onAdditionalDetails: (state, component) => {
        //3ds
        console.log("onAdditionalDetails");
        payload = {
          booking_id: booking_id,
          raw: state.data
        }
        handleBookingEngineSubmission(payload, component, "/payments/adyen/additional-details");
      },
    };

    const checkout = new AdyenCheckout(configuration);
    checkout.create(type).mount(document.getElementById(type));
  } catch (error) {
    console.error(error);
    alert("Error occurred. Look at console for details");
  }
}

function filterUnimplemented(pm) {
  pm.paymentMethods = pm.paymentMethods.filter((it) =>
    ["ach", "scheme", "dotpay", "giropay", "ideal", "directEbanking", "klarna_paynow", "klarna", "klarna_account"].includes(it.type)
  );
  return pm;
}

async function handleBookingEngineSubmission(payload, component, url, confirm) {
  try {
    const res = await callBookingEngineApi(url, payload);
    var adyenResponse = JSON.parse(atob(res.raw))
    handleServerResponse(adyenResponse, component);
  } catch (error) {
    console.error(error);
    alert("Error occurred. Look at console for details");
  }
}

async function callBookingEngineApi(url, data) {
  const res = await fetch(`http://localhost:8000/v2${url}`, {
    method: "POST",
    body: data ? JSON.stringify(data) : "",
    headers: {
      "X-Api-Key": "op6RBIpH81ER1kpb28tF",
      "Content-Type": "application/json",
    },
  });

  console.log(`callServer ::: ${res}`);
  return await res.json();
}

// Handles responses sent from your server to the client
async function handleServerResponse(res, component) {
  if (res.action) {
    component.handleAction(res.action);
  } else {
    let confirmResp;
    switch (res.resultCode) {
      case "Confirmed":
      case "Authorised":
      case "AuthenticationNotRequired":
        confirmResp = await callBookingEngineApi("/payments/adyen/confirm", {booking_id: booking_id});
        console.debug(confirmResp)
        window.location.href = "/result/success";
        break;
      case "AuthenticationFinished":
      case "Pending":
      case "Received":
        confirmResp = await callBookingEngineApi("/payments/adyen/confirm", {booking_id: booking_id});
        console.debug(confirmResp)
        window.location.href = "/result/pending";
        break;
      case "Refused":
        window.location.href = "/result/failed";
        break;
      default:
        window.location.href = `/result/error?reason=${res.resultCode}`;
        break;
    }
  }
}

initCheckout();
