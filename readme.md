# Go API BoilerPlate

Base api setup that i've liked to use now, built in GO
You even get this readme to start with

### RESTful 
There are public endpoints, which require no authentication, and there are private ones
Private calls requires a Bearer Token `user_id:user_token` that get authenticated each time

## Getting Started

This is built to use Cockroach DB, Redis.

### Prerequisites

This requires two (2) environmental variables for handling the file location of the config file as well as the templates directory

```
API_TEMPLATE=/var/app/src/github.com/NathanRThomas/boiler_api/templates/
API_CONFIG=/var/app/config.json
```

### Installing

Build and install the go code

```
go build -o $GOPATH/bin/api -i github.com/NathanRThomas/boiler_api/cmd/api
```

Run the schema against your cockroach database

```
cockroach sql --insecure < api_schema.sql
```

Update the api_config.json file with your application information and make sure it's readable from the API_CONFIG env location

## Running the tests

You can get the version

```
api -v
```

## Deployment

Best to commit changes to the repo, and build on the production machines.  Binary files don't do well in source code control

You're going to want to run this as a service. The following is an example of a /etc/systemd/system/api.service

```
[Unit]
Description=GOLang API
After=network.target

[Service]
Environment=API_CONFIG=/var/app/config.json
Environment=API_TEMPLATE=/var/app/src/github.com/NathanRThomas/boiler_api/templates/

LimitNOFILE=65536
Type=simple
User=root
ExecStart=/usr/bin/api -p=8080
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.targetroot
```

## Built With
GOLang v1.14.1

## Contributing


## Versioning


## Authors

* **Nathan Thomas** - *Initial work* - [LinkedIn](https://www.linkedin.com/in/nathanrthomas1/)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

