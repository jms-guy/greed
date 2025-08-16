![Test Status](https://github.com/jms-guy/greed/actions/workflows/CI.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jms-guy/greed)](https://goreportcard.com/report/github.com/jms-guy/greed)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


# GREED

## Table of Contents
- [Overview](#overview)
    - [Server Features](#server-features)
    - [CLI Features](#cli-features)
- [System Requirements](#system-requirements)
- [Installation](#installation-options)
- [Usage](#usage)
- [To-Do List](#to-do-list)
- [Contributing & Issues](#contributing--issues)

## Overview

Greed is a financial application written almost entirely in Golang, used to view/track data easily across financial institutions/accounts. It utilizes the third-party API [Plaid](https://plaid.com) to connect your account with your financial institutions, and obtain your account and transaction history. This data can then be viewed in a user friendly format, which can be used to track past expenditures and plan future ones.

Since this app utilizes paid Plaid functions, users are restricted in a 'demo' mode, in the number of calls they can make to certain server endpoints, specifically those that talk to Plaid. 

### Server Features

- RESTful API
- [Endpoints](https://github.com/jms-guy/greed/blob/main/docs/endpoints.md)
- Postgres database
- No storing of sensitive personal or financial information, with the exception of Plaid Access Tokens, which are encrypted at rest
- JWT authentication
- Integration with financial data aggregator [Plaid](https://plaid.com/)
- Plaid webhooks, allowing notification of users of updates available for their items
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
    - Viewable in tables and graphs
- Export data into a CSV file

## System Requirements
- No dependencies for binary installation
- [Docker](https://www.docker.com/) (for Docker installation)
- [Go](https://go.dev/doc/install) 1.24+ (for source installation)

## Installation Options

1. Download Binary (Recommended)
    - Download from [Releases](https://github.com/jms-guy/greed/releases)
    - Make it executable and run:

        - **Linux:**
        ```bash
        # Download the appropriate binary for your machine
        chmod +x greed-cli-linux-amd64
        ./greed-cli-linux-amd64 --help
        
        # Optional: Rename for easier use
        mv greed-cli-linux-amd64 greed
        ./greed register ExampleUser
        ```

        - **Windows:**
        ```powershell
        # Download greed-cli-windows-amd64.exe
        # No chmod needed, just run directly
        greed-cli-windows-amd64.exe --help
        greed-cli-windows-amd64.exe register ExampleUser
        ```

2. Docker
    - Have Docker installed
    - Pull the image:
    ```bash
    docker pull jmsguy/greed-cli
    ```
    - Create directories:
        - **Linux:**
        ```bash
        mkdir -p ~/.config/greed
        ```
        - **Windows:**
        ```powershell
        mkdir "$env:USERPROFILE\.config\greed"
        ```
    
    - Add an alias for easy input:
        
        - **Linux/macOS:**
        ```bash
        alias greed='docker run -it -p 8080:8080 --user $(id -u):$(id -g) -e HOME=/home/user -v ~/.config/greed:/home/user/.config/greed jmsguy/greed-cli'
        ```
        
        - **Windows (PowerShell):**
        ```powershell
        function greed { docker run -it -p 8080:8080 -v "$env:USERPROFILE\.config\greed:/root/.config/greed" jmsguy/greed-cli $args }
        ```

        - Docker flags breakdown:
            - '-it': Run the docker image in an interactive terminal session, for getting input after the original command
            - '-p 8080:8080': Maps the container's port 8080 to host, allowing the CLI to open a temp server to listen for Link callback
            - '--user': Runs docker container as your user instead of root, granting user correct read/write permissions for volume
            - '-v': Mounts a docker volume to the image, allowing for client-side database use and export functionality

        - To make the alias permanent, add it to your shell profile:
            - **Linux/macOS:**
            ```bash
            echo "alias greed='docker run -it -p 8080:8080 --user $(id -u):$(id -g) -e HOME=/home/user -v ~/.config/greed:/home/user/.config/greed jmsguy/greed-cli'" >> ~/.bashrc
            source ~/.bashrc
            ```

            - **Windows PowerShell:**
            ```powershell
            Add-Content $PROFILE "function greed { docker run -it -p 8080:8080 -v `"`$env:USERPROFILE\.config\greed:/root/.config/greed`" jmsguy/greed-cli `$args }"
            . $PROFILE
            ```

    - Run commands!
    ```bash
    greed register ExampleUser
    greed login ExampleUser
    ```

    - **Note for Docker users:** Link opening is not supported in containers. The CLI will display the link for you to copy and paste into your browser.

3. Install directly
    - Requires Go
    ```bash
    go install github.com/jms-guy/greed@latest
    greed --help
    ```

4. Clone repo
    - Requires Go
    ```bash
    git clone https://github.com/jms-guy/greed
    cd greed
    go build
    ./greed --help
    ```

## Usage

- List of CLI commands found [here](https://github.com/jms-guy/greed/blob/main/docs/CLI_commands.md)
- Help can be found with the command:
```bash
greed --help
```

## To-Do List

- CLI tests + integration tests
- Recurring transaction detection
- Custom transaction tags and filtering
    - Tag certain merchants/transactions with custom labels (fixed expense, variable expense, tax-deductible, vacation fund, etc.)
- Web client

## Contributing & Issues

To contribute, clone the repo as described above in **Installation Options**. Please fork the repository and open a pull request to the `main` branch. If you have an issue, please report it [here](https://github.com/jms-guy/greed/issues).