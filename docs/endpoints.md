## Public Server Endpoints
---

API endpoints are grouped by operation types

### Auth Operations - /api/auth

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/register` | `POST` | [UserDetails](https://github.com/jms-guy/greed/blob/main/models/request.go#L36) | [User](https://github.com/jms-guy/greed/blob/main/models/response.go#L120) | Creates a new user record |
| `/login` | `POST` | [UserDetails](https://github.com/jms-guy/greed/blob/main/models/request.go#L36) | [Credentials](https://github.com/jms-guy/greed/blob/main/models/response.go#L109) | Creates a "session" for a user |
| `/logout` | `POST` | [RefreshRequest](https://github.com/jms-guy/greed/blob/main/models/request.go#L28) | | Revokes a user's session |
| `/refresh` |  `POST` | [RefreshRequest](https://github.com/jms-guy/greed/blob/main/models/request.go#L28) | [RefreshResponse](https://github.com/jms-guy/greed/blob/main/models/response.go#L100) | Generates a new JWT/refresh token for user |
| `/reset-password` | `POST` | [ResetPassword](https://github.com/jms-guy/greed/blob/main/models/request.go#L53) | | Resets a user's forgotten password |
| `/email/send` | `POST` | [EmailVerification](https://github.com/jms-guy/greed/blob/main/models/request.go#L70) | | Sends a verification code to user's submitted email |
| `/email/verify` | `POST` | [EmailVerificationWithCode](https://github.com/jms-guy/greed/blob/main/models/request.go#L45) | | Verifies a user's email with a code sent to them |

### User Operations - /api/users

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/me` | `GET` | | | Returns a user record |
| `/me` | `DELETE` | | | Deletes a user record |
| `/update-password` | `PUT` | | [UpdatePassword](https://github.com/jms-guy/greed/blob/main/models/response.go#L93) | Updates a user's password - requires an email code |

### Plaid Operations - /plaid

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/get-link-token` | `POST` | | [LinkResponse](https://github.com/jms-guy/greed/blob/main/models/response.go#L86) | Gets a Link token from Plaid to return to client |
| `/get-access-token` | `POST` | [AccessTokenRequest](https://github.com/jms-guy/greed/blob/main/models/request.go#L19) | [AccessResponse](https://github.com/jms-guy/greed/blob/main/models/response.go#L76) | Exchanges a client's public token for an access token from Plaid |

### Item Operations - /api/items

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/` | `GET` | | | Returns a list of Plaid items for user |
| `/{item-id}/name` | `PUT` | [UpdateItemName](https://github.com/jms-guy/greed/blob/main/models/request.go#L12) | Updates an item's name in record |
| `/{item-id}/` | `DELETE` | | | Deletes an item's records from database |
| `/{item-id}/accounts` | `GET` | | [Accounts](https://github.com/jms-guy/greed/blob/main/models/response.go#L22) | Returns list of accounts for a user's specified item |
| `/{item-id}/access/accounts` | `POST` | | | Creates/Updates account records for Plaid item |
| `/{item-id}/access/balances` | `PUT` | | | Update accounts database records with real-time balances |
| `/{item-id}/access/transactions` | `POST` | | | Sync database transaction records for item with Plaid |

### Account Operations - /api/accounts

| Endpoint | Http Method | Request JSON Struct | Response JSON Struct | Description |
| :----:  | :----:  | :----:  | :----:  | :----:  |
| `/` | `GET` | | [Account](https://github.com/jms-guy/greed/blob/main/models/response.go#L33) | Returns list of all accounts for user |
| `/{account-id}/data` | `GET` | | [Account](https://github.com/jms-guy/greed/blob/main/models/response.go#L33) | Returns a single account record for user |
| `/{account-id}` | `DELETE` | | | Delete's an account record |
| `/{account-id}/transactions` | `GET` | | [Transaction](https://github.com/jms-guy/greed/blob/main/models/response.go#L51)/[MerchantSummary](https://github.com/jms-guy/greed/blob/main/models/response.go#L146) | Get all transaction records for account |
| `/{account-id}/transactions` | `DELETE` | | | Delete all transaction records for account |
| `/{account-id}/transactions/monetary` | `GET` | | [MonetaryData](https://github.com/jms-guy/greed/blob/main/models/response.go#L135) | Get monetary data for history of account |
| `/{account-id}/transactions/monetary/{year}-{month}` | `GET` | | [MonetaryData](https://github.com/jms-guy/greed/blob/main/models/response.go#L135) | Get monetary data for given month |
