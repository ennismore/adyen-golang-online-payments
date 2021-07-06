const clientKey = document.getElementById("clientKey").innerHTML;
const type = document.getElementById("type").innerHTML;

async function initCheckout() {
  try {
    const paymentMethodsResponse = await callServer("/api/getPaymentMethods", {});
    const configuration = {
      paymentMethodsResponse: filterUnimplemented(paymentMethodsResponse),
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
          handleSubmission(state, component, "/api/initiatePayment");
        }
      },
      onAdditionalDetails: (state, component) => {
        console.log("onAdditionalDetails");
        handleApiCall(state, component, "/api/submitAdditionalDetails");
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

async function handleApiCall(state, component, url) {
  try {
    const res = await callServer(url, state.data);
    console.log(res);
    await handleSubmission(state, component, "/api/confirm");
  } catch (error) {
    console.error(error);
    alert("Error occurred. Look at console for details");
  }
}

// Event handlers called when the shopper selects the pay button,
// or when additional information is required to complete the payment
async function handleSubmission(state, component, url) {
  try {
    const res = await callServer(url, state.data);
    handleServerResponse(res, component);
  } catch (error) {
    console.error(error);
    alert("Error occurred. Look at console for details");
  }
}

// Calls your server endpoints
async function callServer(url, data) {
  const res = await fetch(url, {
    method: "POST",
    body: data ? JSON.stringify(data) : "",
    headers: {
      "Content-Type": "application/json",
    },
  });

  console.log(`callServer ::: ${res}`);
  return await res.json();
}

// Handles responses sent from your server to the client
function handleServerResponse(res, component) {
  if (res.action) {
    component.handleAction(res.action);
  } else {
    switch (res.resultCode) {
      case "Confirmed":
      case "Authorised":
      case "AuthenticationNotRequired":
        window.location.href = "/result/success";
        break;
      case "AuthenticationFinished":
      case "Pending":
      case "Received":
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
