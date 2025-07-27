## Public Server Endpoints
---

API endpoints are grouped by operation types

### Auth Operations - /api/auth

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/register` | `POST` | [UserDetails](https://github.com/jms-guy/greed/blob/main/models/request.go#L22) | [User](https://github.com/jms-guy/greed/blob/main/models/response.go#L90) | Creates a new user record |
| `/login` | `POST` | [UserDetails](https://github.com/jms-guy/greed/blob/main/models/request.go#L22) | [Credentials](https://github.com/jms-guy/greed/blob/main/models/response.go#L81) | Creates a "session" for a user |
| `/logout` | `POST` | [RefreshRequest](https://github.com/jms-guy/greed/blob/main/models/request.go#L18) | | Revokes a user's session |
| `/refresh` |  `POST` | [RefreshRequest](https://github.com/jms-guy/greed/blob/main/models/request.go#L18) | [RefreshResponse](https://github.com/jms-guy/greed/blob/main/models/response.go#L74) | Generates a new JWT/refresh token for user |
| `/reset-password` | `POST` | [ResetPassword](https://github.com/jms-guy/greed/blob/main/models/request.go#L33) | | Resets a user's forgotten password |
| `/email/send` | `POST` | [EmailVerification](https://github.com/jms-guy/greed/blob/main/models/request.go#L44) | | Sends a verification code to user's submitted email |
| `/email/verify` | `POST` | [EmailVerificationWithCode](https://github.com/jms-guy/greed/blob/main/models/request.go#L28) | | Verifies a user's email with a code sent to them |

### User Operations - /api/users

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/me` | `GET` | | | Returns a user record |
| `/me` | `DELETE` | | | Deletes a user record |
| `/update-password` | `PUT` | [UpdatePassword](https://github.com/jms-guy/greed/blob/main/models/request.go#L39) | [UpdatedPassword](https://github.com/jms-guy/greed/blob/main/models/response.go#L69) | Updates a user's password - requires an email code |

### Plaid Operations - /plaid

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/get-link-token` | `POST` | | [LinkResponse](https://github.com/jms-guy/greed/blob/main/models/response.go#L64) | Gets a Link token from Plaid to return to client |
| `/get-link-token-update` | `POST` | | [LinkResponse](https://github.com/jms-guy/greed/blob/main/models/response.go#L64) | Gets a Link token from Plaid to return to client, containing user's Plaid Access token for update mode |
| `/get-access-token` | `POST` | [AccessTokenRequest](https://github.com/jms-guy/greed/blob/main/models/request.go#L13) | [AccessResponse](https://github.com/jms-guy/greed/blob/main/models/response.go#L56) | Exchanges a client's public token for an access token from Plaid |

### Item Operations - /api/items

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/` | `GET` | | [ItemName](https://github.com/jms-guy/greed/blob/main/models/response.go#L8) | Returns a list of Plaid items for user |
| `/webhook-records` | `GET` | | [WebhookRecord](https://github.com/jms-guy/greed/blob/main/models/response.go#L117) | Returns records of Plaid webhook alerts related to user's items |
| `/webhook-records` | `PUT` | [ProcessWebhook](https://github.com/jms-guy/greed/blob/main/models/request.go#L49) | | Processes a user's webhooks of a given type, after user has resolved them |
| `/{item-id}/name` | `PUT` | [UpdateItemName](https://github.com/jms-guy/greed/blob/main/models/request.go#L9) | | Updates an item's name in record |
| `/{item-id}/` | `DELETE` | | | Deletes an item |
| `/{item-id}/accounts` | `GET` | | [Accounts](https://github.com/jms-guy/greed/blob/main/models/response.go#L14) | Returns list of accounts for a user's specified item |
| `/{item-id}/access/accounts` | `POST` | | [Accounts](https://github.com/jms-guy/greed/blob/main/models/response.go#L14) | Creates/Updates account records for Plaid item. Restricted access for demo users |
| `/{item-id}/access/balances` | `PUT` | | [UpdatedBalance](https://github.com/jms-guy/greed/blob/main/models/response.go#L47) | Update accounts database records with real-time balances. Restricted access for demo users |
| `/{item-id}/access/transactions` | `POST` | | [Transaction](https://github.com/jms-guy/greed/blob/main/models/response.go#L35) | Sync database transaction records for item with Plaid. Restricted access for demo users |

### Account Operations - /api/accounts

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/` | `GET` | | [Account](https://github.com/jms-guy/greed/blob/main/models/response.go#L14) | Returns list of all accounts for user |
| `/{account-id}/data` | `GET` | | [Account](https://github.com/jms-guy/greed/blob/main/models/response.go#L14) | Returns a single account record for user |
| `/{account-id}` | `DELETE` | | | Delete's an account record |
| `/{account-id}/transactions` | `GET` | | [Transaction](https://github.com/jms-guy/greed/blob/main/models/response.go#L35)/[MerchantSummary](https://github.com/jms-guy/greed/blob/main/models/response.go#L109) | Get all transaction records for account |
| `/{account-id}/transactions` | `DELETE` | | | Delete all transaction records for account |
| `/{account-id}/transactions/monetary` | `GET` | | [MonetaryData](https://github.com/jms-guy/greed/blob/main/models/response.go#L101) | Get monetary data for history of account |
| `/{account-id}/transactions/monetary/{year}-{month}` | `GET` | | [MonetaryData](https://github.com/jms-guy/greed/blob/main/models/response.go#L101) | Get monetary data for given month |


### Plaid Link Redirects

| Endpoint | Http Method | Description |
| :----:  | :----:  | :----:  |
| `/link` | `GET` | Provides the redirect page for handling Plaid's Link flow |
| `/link-update-mode` | `GET` | Provides redirect page for handling Plaid's Link Update mode flow |