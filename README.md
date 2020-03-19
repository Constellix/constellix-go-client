# constellix-go-client
 This repository contains the golang client SDK to interact with Constellix DNS Platform using REST API calls. This SDK is used by [terraform-provider-constellix](https://github.com/Constellix/terraform-provider-constellix).

## Installation ##

Use `go get` to retrieve the SDK to add it to your `GOPATH` workspace, or project's Go module dependencies.


```sh
$go get github.com/Constellix/constellix-go-client
```

There are no additional dependancies needed to be installed.

## Overview ##
  
* <strong>client</strong> :- This package contains the HTTP Client configuration as well as service methods which serves the CRUD operations on the DNS Objects in Constellix DNS platform.

* <strong>models</strong> :- This package contains all the models structs and utility methods for the same.

## How to Use ##

import the client in your go application and retrive the client object by calling client.GetClient() method.
```golang
import github.com/Constellix/constellix-go-client/client
client.GetClient("apikey", "secretkey", client.Insecure(true/false))
```


Use that client object to call the service methods to perform the CRUD operations on the model objects.

Example,

```golang
    client.Save(obj<type interface>,endpoint<type string>)
```