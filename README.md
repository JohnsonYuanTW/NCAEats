# Go Lunch Order Linebot
Go Lunch Order Linebot is a tool designed to simplify the process of organizing group lunch orders. It allows admins to create, read, update, and delete (CRUD) restaurants and menu items. Admins can start new orders, and other users can add their desired menu items to the order. Once the order is collected, the system generates two detailed reports: one for the restaurant detailing the total count of each item and the total price, and another for the orderer detailing each person's order for easy money collection

## Features
* CRUD operations for restaurants and menu items
* User-friendly interface for starting and participating in orders
* Automated report generation for ease of ordering and payment collection

## Roadmap
- [x] Project Initialization
    - [x] Implement basic functionalities for Linebot
    - [x] Utilize .env for managing environment variables
- [ ] Basic Functionality
    - [x] Implement listing all restaurants and order creation
    - [x] Enable creation of menu items and ordering
    - [x] Generate two reports
    - [ ] Multiple menu import methods
        - [x] linebot
        - [ ] .csv
        - [ ] .xls, .xlsx
- [ ] Refactoring
    - [x] Integrate gorm for Object-Relational Mapping (ORM)
    - [ ] Improve error handling using the `errors` package
    - [ ] Utilize `logrus` for logging
    - [ ] Implement Flex Response
        - [ ] Introduce Flex Response in all functionalities
        - [x] Enhance the approach for loading JSON using DI
    - [x] Enhance the project with DI
- [ ] Docker Integration
    - [ ] Create Dockerfile for building application image
    - [ ] Set up Docker Compose for local development and testing
    - [ ] (Optional) Configure a Docker registry for storing and distributing your application image
- [ ] Testing
    - [ ] Conduct unit tests
    - [ ] Perform integrated tests

## Configuration

The application requires certain env variables to be set for proper operation. These variables can be set in a `.env` file in the root of your project. 

Here is a list of required environment variables:

- `ChannelSecret`: Your Linebot channel secret.
- `ChannelAccessToken`: Your Linebot channel access token.
- `SSLCertfilePath`: Path to your SSL certificate file.
- `SSLKeyPath`: Path to your SSL key file.
- `SITE_URL`: URL of your site.
- `PORT`: Port for your application to run on.
- `DB_USERNAME`: Username for your PostgreSQL database.
- `DB_PASSWORD`: Password for your PostgreSQL database.
- `DB_URL`: URL for your PostgreSQL database.
- `DB_NAME`: Name of your PostgreSQL database.
- `DB_PORT`: Port number for your PostgreSQL database.
- `DB_MAX_IDLE_CONNS`: Maximum number of idle connections in the PostgreSQL connection pool.
- `DB_MAX_OPEN_CONNS`: Maximum number of open connections to the PostgreSQL database.
- `DB_CONN_MAX_LIFETIME`: Maximum lifetime of a connection to the PostgreSQL database.

You can copy the `.env.example` file to a new file named `.env` and fill in the appropriate values. 

```bash
cp .env.example .env
```

## How to Use
WIP

## Requirements
WIP

## Installation
WIP

# Contributing
If you have any suggestions or improvements (better!), please create a pull request! 

# License
This project is licensed under the Apache 2.0 License.

# Contact
[@JohnsonYuanTW](https://github.com/JohnsonYuanTW/)
