<div align="center" id="top"> 
  <img src="./.github/app.gif" alt="Grollmus Golang" />

  &#xa0;

  <!-- <a href="https://grollmosgolang.netlify.app">Demo</a> -->
</div>

<h1 align="center">Grollmus Golang</h1>

<p align="center">
  <img alt="Github top language" src="https://img.shields.io/github/languages/top/DerRomtester/testproject-golang-webapi-1">

  <img alt="Github language count" src="https://img.shields.io/github/languages/count/DerRomtester/testproject-golang-webapi-1">

  <img alt="Repository size" src="https://img.shields.io/github/repo-size/DerRomtester/testproject-golang-webapi-1">

  <img alt="License" src="https://img.shields.io/github/license/DerRomtester/testproject-golang-webapi-1">

  <!-- <img alt="Github issues" src="https://img.shields.io/github/issues/DerRomtester/testproject-golang-webapi-1?color=56BEB8" /> -->

  <!-- <img alt="Github forks" src="https://img.shields.io/github/forks/DerRomtester/testproject-golang-webapi-1?color=56BEB8" /> -->

  <!-- <img alt="Github stars" src="https://img.shields.io/github/stars/DerRomtester/testproject-golang-webapi-1?color=56BEB8" /> -->
</p>

<!-- Status -->

<!-- <h4 align="center"> 
	🚧  Grollmos Golang 🚀 Under construction...  🚧
</h4> 

<hr> -->

<p align="center">
  <a href="#dart-about">About</a> &#xa0; | &#xa0; 
  <a href="#sparkles-features">Features</a> &#xa0; | &#xa0;
  <a href="#rocket-technologies">Technologies</a> &#xa0; | &#xa0;
  <a href="#white_check_mark-requirements">Requirements</a> &#xa0; | &#xa0;
  <a href="#checkered_flag-starting">Starting</a> &#xa0; | &#xa0;
  <a href="#memo-license">License</a> &#xa0; | &#xa0;
  <a href="https://github.com/DerRomtester target="_blank">Author</a>
</p>

<br>

## :dart: About ##

A small simple REST API project i found from grollmus

## :rocket: Technologies ##

The following tools were used in this project:

- [Go](https://go.dev/)

## :white_check_mark: Requirements ##

Before starting :checkered_flag:, you need to have [Git](https://git-scm.com) and [Go](https://go.dev) installed.

## API ##
Authenticate
POST http://localhost:23452/v1/auth

Body:
{
    "username" : "test_user",
    "password" : "password123"
}

==> cookie as result

Create User
POST http://localhost:23452/v1/user
{
    "username" : "test_user",
    "password" : "password123"
}

Logout
PUT http://localhost:23452/v1/auth


Get all devices
GET http://localhost:23452/v1/devices


Get device by id
GET http://localhost:23452/v1/device/ID_HERE


Current Session
http://localhost:23452/v1/session


Refresh token and cookie
PUT http://localhost:23452/v1/refresh


Create devices
POST http://localhost:23452/v1/devices

Body:
{
  "devices": [
    {
      "id": "1glmLrTZqf9YZleN",
      "name": "S7-150009",
      "deviceTypeId": "Beweis",
      "failsafe": true,
      "tempMin": 0,
      "tempMax": 60,
      "installationPosition": "horizontal",
      "insertInto19InchCabinet": true,
      "motionEnable": true,
      "siplusCatalog": true,
      "simaticCatalog": true,
      "rotationAxisNumber": 0,
      "positionAxisNumber": 0
    },
  ]
}


Delete all devices
DELETE http://localhost:23452/v1/devices


Delete device by its id
DELETE http://localhost:23452/v1/device/1glmLrTZqf9YZleN


## :checkered_flag: Starting ##

```bash
# Clone this project
$ git clone https://github.com/DerRomtester/testproject-golang-webapi-1.git

# Access
$ cd testproject-golang-webapi-1

# Run the project 
# The server will initialize in the <http://localhost:23452>
$ make start-container

# Build the Binary
$ mkdir bin
$ make build

```

Made with :heart: by <a href="https://github.com/DerRomtester" target="_blank">DerRomtester</a>

&#xa0;

<a href="#top">Back to top</a>
