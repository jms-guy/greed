package plaidservice

import (
	"context"
	"github.com/plaid/plaid-go/v36/plaid"
)

//Retrieves account data listed in a Plaid item
func GetAccounts(client *plaid.APIClient, ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error){

	accountsGetRequest := plaid.NewAccountsGetRequest(accessToken)

	accountsGetResp, httpResp, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*accountsGetRequest,
	).Execute()
	if err != nil {
		return nil, "", err
	}

	accounts := accountsGetResp.GetAccounts()

	return accounts, httpResp.Header.Get("X-Request-Id"), nil
}

//Gets institution name for an item
func GetItemInstitution(client *plaid.APIClient, ctx context.Context, accessToken string) (string, error) {
	itemGetReq := plaid.NewItemGetRequest(accessToken)
	itemResp, _, err := client.PlaidApi.ItemGet(ctx).ItemGetRequest(*itemGetReq).Execute()
	if err != nil {
		return "", err 
	}

	item := itemResp.GetItem()
	institutionID := item.GetInstitutionId()

	instReq := plaid.NewInstitutionsGetByIdRequest(institutionID, []plaid.CountryCode{plaid.COUNTRYCODE_CA})
	instResp, _, err := client.PlaidApi.InstitutionsGetById(ctx).InstitutionsGetByIdRequest(*instReq).Execute()
	if err != nil {
		return "", err 
	}

	inst := instResp.GetInstitution()

	return inst.GetName(), nil
}

func GetBalances(client *plaid.APIClient, ctx context.Context, accessToken string) (plaid.AccountsGetResponse, string, error) {
	balancesGetReq := plaid.NewAccountsBalanceGetRequest(accessToken)
	balancesGetResp, httpResp, err := client.PlaidApi.AccountsBalanceGet(ctx).AccountsBalanceGetRequest(*balancesGetReq).Execute()

	if err != nil {
		return plaid.AccountsGetResponse{}, "", err
	}

	return balancesGetResp, httpResp.Header.Get("X-Request-Id"), nil
}