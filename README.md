# Muumuu Domain API in Go

[![GoDoc](https://pkg.go.dev/badge/github.com/kenzo0107/muumuu-domain.svg)](https://pkg.go.dev/github.com/kenzo0107/muumuu-domain) [![test](https://github.com/kenzo0107/muumuu-domain/actions/workflows/test.yml/badge.svg)](https://github.com/kenzo0107/muumuu-domain/actions/workflows/test.yml) [![lint](https://github.com/kenzo0107/muumuu-domain/actions/workflows/lint.yml/badge.svg)](https://github.com/kenzo0107/muumuu-domain/actions/workflows/lint.yml)

Go client for the [Muumuu Domain API v2](https://muumuu-domain.com/developers/openapi-me.html).

## Installing

```sh
go get github.com/kenzo0107/muumuu-domain
```

## Example

```go
package main

import (
	"context"
	"log"
	"os"

	muumuu "github.com/kenzo0107/muumuu-domain"
)

func main() {
	c, err := muumuu.New(os.Getenv("MUUMUU_DOMAIN_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	domains, err := c.GetDomains(ctx, &muumuu.InputGetDomains{FQDN: "example.com"})
	if err != nil {
		log.Fatal(err)
	}

	record, err := c.CreateDNSRecord(ctx, domains.Domains[0].ID, &muumuu.InputCreateDNSRecord{
		FQDN:  "www.example.com",
		Type:  muumuu.DNSRecordTypeA,
		Value: "192.0.2.1",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created: %+v", record)
}
```

### Sandbox

```go
c, err := muumuu.New(token, muumuu.OptionBaseURL(muumuu.BaseURLSandbox))
```

### Error handling

API errors are returned as `*muumuu.Error`, which carries the HTTP status, the error `Code`, and field `Details`.

```go
d, err := c.GetDomain(ctx, "MU00000001")
if muumuu.IsNotFound(err) {
	// e.g. remove the resource from Terraform state
}
```

`IsRateLimited(err)` reports HTTP 429, and `Error.RetryAfter` holds the `Retry-After` header value.

## Supported APIs

| API | Method |
| --- | --- |
| `GET /me/domains` | `GetDomains` |
| `GET /me/domains/{domain-id}` | `GetDomain` |
| `PATCH /me/domains/{domain-id}` | `UpdateDomain` |
| `GET /me/domains/{domain-id}/dns-records` | `GetDNSRecords` |
| `POST /me/domains/{domain-id}/dns-records` | `CreateDNSRecord` |
| `PATCH /me/domains/{domain-id}/dns-records/{record-id}` | `UpdateDNSRecord` |
| `DELETE /me/domains/{domain-id}/dns-records/{record-id}` | `DeleteDNSRecord` |
| (paginates `GET /me/domains/{domain-id}/dns-records`) | `GetDNSRecord` |
| `GET /me/dns-records` | `GetDNSRecordsByDomainFQDN` |
| `GET /me/domains/{domain-id}/nameservers` | `GetNameservers` |
| `PATCH /me/domains/{domain-id}/nameservers` | `UpdateNameservers` |
| `GET /domain-search` | `SearchDomains` |
| `GET /maintenances` | `GetMaintenances` |
| `POST /me/domain-purchase/quote` | `QuoteDomainPurchase` |
| `POST /me/domain-purchase` | `PurchaseDomain` |
| `POST /me/domain-purchase/direct` | `PurchaseDomainDirect` |
| `GET /me/domain-purchase/{purchase_id}/status` | `GetDomainPurchaseStatus` |
| `GET /me/payment-history` | `GetPaymentHistory` |
| `GET /me/personal-access-tokens` | `GetPersonalAccessTokens` |
| `DELETE /me/personal-access-tokens/{id}` | `DeletePersonalAccessToken` |

## Development

```sh
mise run fmt
mise run lint
mise run test

# against the sandbox (set MUUMUU_DOMAIN_TEST_DOMAIN_ID to also run DNS record CRUD)
MUUMUU_DOMAIN_TOKEN=muu_pat_sandbox_xxx mise run test-live
```

## License

[MIT License](LICENSE)
