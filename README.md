
# GREED

## Overview

Greed is a financial application written almost entirely in Golang, used to view/track data easily across financial institutions/accounts. It utilizes the third-party API [Plaid](https://plaid.com) to connect your account with your financial institutions, and obtain your account and transaction history. This data can then be viewed in a user friendly format, which can be used to track past expenditures and plan future ones.

Unless you are running your own server for this application, CLI users are restricted in a 'demo' mode. Since this app utilizes paid Plaid functions, users are restricted in the number of calls they can make to certain server endpoints, specifically those that talk to Plaid. 

### Server Features

- RESTful API
- [Endpoints](https://github.com/jms-guy/greed/blob/main/docs/endpoints.md)
- 80% handler testing coverage
- Postgres database
- No storing of sensitive personal or financial information, with the exception of Plaid Access Tokens, which are encrypted at rest
- Simple IP-based rate limiter
- JWT authentication
- Integration with financial data aggregator [Plaid](https://plaid.com/)
- Plaid webhooks, allowing notification of users of updates avilable for their items
- Account-email verification utilizing [SendGrid](https://sendgrid.com/en-us)

### CLI Features

- [Cobra](https://github.com/spf13/cobra)-based CLI tool
- [CLI Commands](https://github.com/jms-guy/greed/blob/main/docs/CLI_commands.md)
- Client SQLite database
- Allows for registering, logging in/out, and deleting users
- Basic reporting of account information for financial institutions
- 24 months of account financial history
- In-depth transaction history reporting 
    - Utilizing paginated tables in terminal
    - Extensive sorting through amount, date, merchant, etc.
    - Allows summary reporting as well
        - ex. All transactions summed, showing count and total amount for each merchant, for each month
        - ex. All transactions for merchant 'A' for month 'X' summed, showing count, total amount, dates
        - ex. All transactions for merchant 'A' summed over full 24-month history, showing count, total amount
- Income/Expense viewing
    - View Income vs. Expenses per account
    - Viewable in tables or graphs
- Export data into a CSV file

### Future Features

- Recurring transaction detection
- Custom transaction tags and filtering
    - Tag certain merchants/transactions with custom labels (fixed expense, variable expense, vacation fund, etc.)