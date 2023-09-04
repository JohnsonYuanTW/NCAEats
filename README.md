# NCAEats - Go Lunch Order Linebot
NCAEats is a tool designed to simplify the process of organizing group lunch orders. It allows admins to create, read, update, and delete (CRUD) restaurants and menu items. Admins can start new orders, and other users can add their desired menu items to the order. Once the order is collected, the system generates two detailed reports: one for the restaurant detailing the total count of each item and the total price, and another for the orderer detailing each person's order for easy money collection

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
- [x] Refactoring
    - [x] Integrate gorm for Object-Relational Mapping (ORM)
    - [x] Improve error handling using the `errors` package
    - [x] Utilize `logrus` for logging
    - [x] Implement Flex Response
        - [x] Introduce Flex Response in all functionalities
        - [x] Enhance the approach for loading JSON using DI
    - [x] Enhance the project with DI
- [ ] Docker Integration
    - [x] Create Dockerfile for building application image
    - [x] Set up Docker Compose for local development and testing
    - [ ] (Optional) Configure a Docker registry for storing and distributing your application image
- [ ] Testing
    - [ ] Conduct unit tests
    - [ ] Perform integrated tests

## Configuration
This document outlines the environment variables used in [Your Project Name]. Make sure to set these variables appropriately in your environment or within your `.env` file.

### LINE Messaging API Configuration
- **CHANNEL_SECRET**: Your LINE Channel Secret. Obtain it from the LINE Developer Console.
- **CHANNEL_ACCESS_TOKEN**: Your LINE Channel Access Token. This is also obtained from the LINE Developer Console.

### SSL Configuration
Line requires webhook to operate under SSL, so you'll need to configure these paths.
- **USE_SSL**: Whether to use SSL. If your setup includes SSL termination, you can disable SSL here. Default is true. 
- **SSL_CERTIFICATE_HOST_PATH**: The path on your host machine to the SSL certificate.
- **SSL_KEY_HOST_PATH**: The path on your host machine to the SSL private key.
- **SSL_CERTIFICATE_PATH**: The path inside your container or application to the SSL certificate. Default is `"/app/ssl/certs/fullchain.pem"`.
- **SSL_KEY_PATH**: The path inside your container or application to the SSL private key. Default is `"/app/ssl/private/privkey.pem"`.

### Server Configuration
- **SITE_URL**: The base URL of your site, e.g., `"https://example.com"`.
- **PORT**: The port on which your application server runs.

### Database Configuration
Configure your database settings here:
- **DB_USERNAME**: The username for your database.
- **DB_PASSWORD**: The password for your database.
- **DB_URL**: The URL or IP address where your database is hosted.
- **DB_NAME**: The name of your database.
- **DB_PORT**: The port for your database connection. Default is `5432` (standard for PostgreSQL).
- **DB_MAX_IDLE_CONNS**: The maximum number of idle connections that can be simultaneously maintained. Default is `10`.
- **DB_MAX_OPEN_CONNS**: The maximum number of open connections to the database. Default is `100`.
- **DB_CONN_MAX_LIFETIME**: The maximum amount of time a connection may be reused. Default is `1h` (1 hour).

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
