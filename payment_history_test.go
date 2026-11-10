package muumuu

import (
	"context"
	"net/http"
	"testing"
)

func TestGetPaymentHistory(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/payment-history", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testQuery(t, r, "amount-max=5000&amount-min=0&date-from=2024-01-01&date-to=2024-12-31&domain-name=example&service-type=domain")
		respond(t, w, http.StatusOK, `{"data":[{
			"id":"MB00000001","paid-at":"2024-01-01","service-type":"domain","target-name":"example.com",
			"description":"取得（2024-01-01～2025-01-01）","amount":1210,
			"amount-breakdown":{"base-price":1000,"tax":110,"surcharge":100,"discount":0},
			"payment-method":"creditcard","status":2,"refunded":false,"refunded-at":null,
			"contract-start-date":"2024-01-01","contract-end-date":"2025-01-01"
		}],"meta":{"total":1,"page":1,"page-size":20}}`)
	})

	got, err := c.GetPaymentHistory(context.Background(), &InputGetPaymentHistory{
		DateFrom:    "2024-01-01",
		DateTo:      "2024-12-31",
		AmountMin:   Int(0),
		AmountMax:   Int(5000),
		ServiceType: "domain",
		DomainName:  "example",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &OutputGetPaymentHistory{
		PaymentHistory: []PaymentHistory{{
			ID:                "MB00000001",
			PaidAt:            String("2024-01-01"),
			ServiceType:       "domain",
			TargetName:        String("example.com"),
			Description:       String("取得（2024-01-01～2025-01-01）"),
			Amount:            1210,
			AmountBreakdown:   PaymentHistoryAmountBreakdown{BasePrice: 1000, Tax: 110, Surcharge: 100},
			PaymentMethod:     String("creditcard"),
			Status:            2,
			ContractStartDate: String("2024-01-01"),
			ContractEndDate:   String("2025-01-01"),
		}},
		Meta: PaginationMeta{Total: 1, Page: 1, PageSize: 20},
	})
}
