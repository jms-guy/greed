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
		return nil, httpResp.Header.Get("X-Request-Id"), err
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
		return plaid.AccountsGetResponse{}, httpResp.Header.Get("X-Request-Id"), err
	}

	return balancesGetResp, httpResp.Header.Get("X-Request-Id"), nil
}

//Calls transactions/sync Plaid endpoint, getting all transaction data for a specific account.
//Returns last Plaid request ID in loop
func GetTransactions(client *plaid.APIClient, ctx context.Context, accessToken, cursor string) (
	added, modified []plaid.Transaction, removed []plaid.RemovedTransaction, nextCursor, reqID string, err error) {

	hasMore := true
	yes := true
	options := plaid.TransactionsSyncRequestOptions{
		IncludePersonalFinanceCategory: &yes,
	}

	reqID = ""

	for hasMore {
		request := plaid.NewTransactionsSyncRequest(accessToken)
		request.SetOptions(options)
		if cursor != "" {
			request.SetCursor(cursor)
		}
		resp, httpResp, err := client.PlaidApi.TransactionsSync(ctx).TransactionsSyncRequest(*request).Execute()
		if err != nil {
			if httpResp != nil {
                reqID = httpResp.Header.Get("X-Request-Id")
            }
			return added, modified, removed, "", reqID, err 
		}

		added = append(added, resp.GetAdded()...)
		modified = append(modified, resp.GetModified()...)
		removed = append(removed, resp.GetRemoved()...)

		hasMore = resp.GetHasMore()
		nextCursor = resp.GetNextCursor()

		if httpResp != nil {
            reqID = httpResp.Header.Get("X-Request-Id")
        }

		cursor = nextCursor
	}
	return added, modified, removed, nextCursor, reqID, nil
}