## Commands for CLI Use

- `ping`
    - Pings the server, checking connection health

- `register <name>`
     - Register a new user account with the server

- `login <name>`
    - Create a new user session with the server

- `logout`
    - Exits current user session

- `delete <user | item> <username_or_item_name>`
    - Deletes a specified user/item record from database. Must be logged in to use
    - Delete user is currently unavailable, only allows for deleting of items

- `verify`
    - Verifies a user's submitted email address, if verification was skipped during user registration 

- `add-item`
    - Connect a financial institution to your account

- `items` 
    - Lists a user's item records. An item is a link to a financial institution, containing all account records for that institution

- `changepw`
    - Updates a user's password. Must have a verified email address

- `resetpw <email>`
    - Resets a user's forgotten password. Must have a verified email address

- `fetch <item-name>`
    - Retrieves all account and transaction data for item from third party, populating database with records. Should only be used on a new item, afterwards use sync command

- `sync <item-name>`
    - Updates account and transaction data for an item, providing the latest data from the financial institution

- `update <item-name>`
    - Re-authenticates user's financial institute through Plaid Link Update mode

- `rename <current-item-name> <new-item-name>`
    - Rename an item 

- `info <account-name>`
    - View extended information for a given account

- `export <account-name>`
    - Export an account's transaction history into a CSV file
    - Export directory is based on operating system
        - Windows: C:\\Users\\user\\Documents\\greed_exports
        - Linux: /home/user/greed_exports

- `logs` 
    - View in-depth error logs stored in local database

- `default <account | item | clear> <account_name | item_name>`
    - Set a default account or item to be used in place of certain command arguments, allowing better user experience
    - Typing an account or item name in these affected commands will override the default set for a single use
    - Clear command resets defaults
        - Ex. `default account "Example Checking Account"`, `default item "Example Item Name"`, `default clear`
        - Allows for `get transactions --summary` instead of `get transactions "Example Checking Account" --summary`

### Get

The most useful command, it has several subcommands, and many flags.
- `get accounts [item-name]`
    - Returns a list of accounts. If an item name is specified, it will return accounts only for that item. Otherwise it will return all accounts for user
- `get transactions <account-name> [flags]`
    - Returns transactions for an account, takes many optional flags that can be used to sort and display transaction data on a paginated table
    - Flags
        - Merchant: Filter transactions by merchant name (`--merchant <merchant-name>`)
        - Category: Filter transactions by category (`--category <category-type>`)
        - Channel: Filter transactions by payment channel (`--channel <channel-type>`)
        - Date: Filter transactions for a specific date (`--date <date>`)(date format 'year-month-day')
        - Start/End: Filter transactions based on a given start and/or end date (`--start <date>`, `--end <date>`)
        - Min/Max: Filter transactions with a given minimum/maximum dollar amount (`--min <amount>`, `--max <amount>`)
        - Limit: Filters transactions by limiting the number shown (`--limit <number>`)
        - Pgsize: Specify the number of records to show on the table at any one time (`--pgsize <number>`) 
        - Order: Reorder the transactions shown by date (`--order <ASC>`)
        - Summary: Provides a summary of transactions. Overrides most other flags. Useful with the [date] & [merchant] flags (`--summary`)
- `get income <account-name> [flag]`
    - Returns aggregate income/expenses data for account history
    - Flags
        -Mode: Include visual output of data (`--mode <graph>`)