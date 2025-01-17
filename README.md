SGX HVS
=======

SGX Host Verification Service aggregates the platform enablement info from multiple SGX Agent instances.

Key features
------------

-   SHVS saves platform specific information provided by a sgx agent instance, in its own database which will be pulled by integration hub later

System Requirements
-------------------

-   RHEL 8.4 or ubuntu 20.04
-   Epel 8 Repo
-   Proxy settings if applicable

Software requirements
---------------------

-   git
-   makeself
-   docker
-   Go 1.18.8

Step By Step Build Instructions
===============================

Install required shell commands
-------------------------------

### Install tools from `dnf`

``` {.shell}
sudo dnf install -y git wget makeself docker
```

### Install `go 1.18.8`

The `Host Verification Service` requires Go version 1.18 that has
support for `go modules`. The build was validated with version 1.18.8
version of `go`. It is recommended that you use a newer version of `go`
- but please keep in mind that the product has been validated with
1.18.8 and newer versions of `go` may introduce compatibility issues.
You can use the following to install `go`.

``` {.shell}
wget https://dl.google.com/go/go1.18.8.linux-amd64.tar.gz
tar -xzf go1.18.8.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

Build SGX-Host Verification Service
-----------------------------------

-   Git clone the SGX Host Verification Service
-   Run scripts to build the SGX Host Verification Service

``` {.shell}
git clone https://github.com/intel-secl/sgx-hvs.git
cd sgx-hvs
git checkout v5.1.0
make all
```

### Manage service

-   Start service
    -   shvs start
-   Stop service
    -   shvs stop
-   Status of service
    -   shvs status

Third Party Dependencies
========================

Certificate Management Service
------------------------------

Authentication and Authorization Service
----------------------------------------

### Direct dependencies

|  Name       | Repo URL                      | Minimum Version Required  |
|  ---------- | ----------------------------- | :-----------------------: |
|  uuid       | github.com/google/uuid        | v1.2.0                    |
|  errors     | github.com/pkg/errors         | V0.9.1                    |
|  handlers   | github.com/gorilla/handlers   | v1.4.2                    |
|  mux        | github.com/gorilla/mux        | v1.7.4                    |
|  gorm       | github.com/jinzhu/gorm        | v1.9.16                   |
|  logrus     | github.com/sirupsen/logrus    | v1.7.0                    |
|  testify    | github.com/stretchr/testify   | v1.6.1                    |
|  yaml.v3    | gopkg.in/yaml.v3              | v3.0.1                    |
|  common     | github.com/intel-secl/common  | v5.1.0                   |

### Indirect Dependencies

  Repo URL                     Minimum version required
  --------------------------- --------------------------
  https://github.com/lib/pq             1.1.0

*Note: All dependencies are listed in go.mod*
