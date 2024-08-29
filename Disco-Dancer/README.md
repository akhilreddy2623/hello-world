**payment-administrator:** 
<a href="https://github.com/geico-private/geico-payment-platform/blob/gh-pages/coverages/payment-administrator.html">
    <img width="125" height="20" align="center" src="https://github.com/geico-private/geico-payment-platform/blob/gh-pages/coverages/payment-administrator.svg?raw=true">
</a>

**task-manager:**
<a href="https://github.com/geico-private/geico-payment-platform/blob/gh-pages/coverages/task-manager.html">
    <img width="125" height="20" align="center" src="https://github.com/geico-private/geico-payment-platform/blob/gh-pages/coverages/task-manager.svg?raw=true">
</a>

**payment-executor:**
<a href="https://github.com/geico-private/geico-payment-platform/blob/gh-pages/coverages/payment-executor.html">
    <img width="125" height="20" align="center" src="https://github.com/geico-private/geico-payment-platform/blob/gh-pages/coverages/payment-executor.svg?raw=true">
</a>


# Getting Started
1.	Install and configure Go:  
    * https://go.dev/doc/install
    * https://geico365.sharepoint.com/sites/GC-SIG/SitePages/Go-Setup.aspx
2.	Install a code editor like VS Code and Go extensions:
    * https://code.visualstudio.com/download
3.  Request Avecto Defendpoint Access:

    To temporarily be granted elevated privileges (used during step 5), have your manager create a ticket in Request Access Center (ARC) to add your account to the following Active Directory group:
    * **CN=ENT-EAST-Avecto-U-PowerUser,OU=Plaza,OU=Admin,DC=GEICO,DC=Corp,DC=Net**

    You can check the status of your ticket by using the ARC portal:
    * https://ssm-geico.ssmcloud.net/ECMv6/request/requestHome

    Once complete, you may run a program with temporary elevated privileges by (shift +)'right-clicking' and selecting "Run with Defendpoint". You may have to click "Show more options" for Defenpoint to be visible.
4.  Installing Git:

    If Git is not already installed on your machine, please use the following link to download version 2.43, as versions >=2.44 encounter issues with Zscalar. If Git is installed, it is recommened to downgrade until this issue is resolved.
    * https://github.com/git-for-windows/git/releases/tag/v2.43.0.windows.1
5.	Install a docker GUI:
    * For Mac - https://geico365.sharepoint.com/sites/GC-SIG/SitePages/macOS%20Rancher%20Desktop%20Setup.aspx
    * For Windows - https://geico365.sharepoint.com/sites/GC-SIG/SitePages/Windows%20Rancher%20Desktop%20Setup.aspx 
    * For Windows - https://geico365.sharepoint.com/sites/GC-SIG/SitePages/Windows%20Podman%20Desktop%20Setup.aspx

    Note : You can install either ranch software or podman desktop but not both

    For rancher enable network tunneling under WSL and use the container image dockerd(moby).
6.	Install Make:
    * https://www.gnu.org/software/make/
    You can install 'Make' software using 'Chocolatey' (https://docs.chocolatey.org/en-us/choco/setup/). use this command:  choco install make
    In windows you can install 'Chocolatey' using powershell. Launch powershell by running powershell exe with 'Run with defendpoint'
7.  Setup proxies:
    * $env:GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/core-golang-all,https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
    * $env:GONOSUMDB="artifactory-pd-infra.aks.aze1.cloud.geico.net/geico,dev.azure.com"
8.  GitHub Access:

    If you do not already have GitHub access to the ```geico-payment-platform``` repo, contact your manager to email the Developer Productivity Engineering (DPE) team on your behalf.
    * DPE Team DL: dpereposandpackage@geico.com
    * https://github.com/geico-private/geico-payment-platform

    After DPE team gives the access, it might take some time for the access to propagte and things to work seemlessly 

    Upon completion of your first PR, you will be added as a 'contributor'.
    Steps to setup the code with Github are in the next section
9.  Code Ownership:

    Edit the ```.github/CODEOWNERS``` file by adding your GitHub user ID to be added as a code owner. Create a PR to complete this step (refer to **_PR Standards and Guidelines_** section for more details).
10.  Voltage artifactory Access Request:

    Create ARC request to get Voltage Access, contact a team member to open ARC ticket
    * https://ssm-geico.ssmcloud.net/ECMv6/request/requestHome

    For **_Entitlement_** select:
    * **CN=ENT-ASG-PackageManager-PD-USER,OU=Plaza,OU=Admin,DC=GEICO,DC=corp,DC=net**
11. In the case that you do not have access to create ADO items in ADO, ask your manager to add you to the Payment Platform Team.


As part of local project setup, make sure to complete “Voltage side car setup” also as mentioned in this document.

## If you are on macOS use the links below to setup your workstation and Git
* https://improved-adventure-1w29355.pages.github.io/sig/sig-mac/macos-developer-workstation-setup.html
* https://improved-adventure-1w29355.pages.github.io/sig/sig-mac/macos-git-setup.html 

## Github Setup For Windows
.gitconfig file needs to be present at location C:\Users\{User}. e.g. C:\Users\u59c35

Important things here are **_sslcainfo_** and at the end the **_url_** parameters
```
[user]
    name = Bajpai, Piyush
    email = pbajpai@geico.com
    proxy = http://gpproxy.geico.net:80
[credential]
    helper = manager
    credentialStore = dpapi
[alias]
    credential-manager = credential-manager-core
[http]
    sslVerify = false
    sslBackend = openssl
    sslcainfo = C:\\Dev\\certs\\zscaler.pem
[filter "lfs"]
	smudge = git-lfs smudge -- %f
	process = git-lfs filter-process
	required = true
	clean = git-lfs clean -- %f
[url "https://geico:<ADO-token>>@dev.azure.com/geico"]
	  insteadOf = https://dev.azure.com/geico  
```

* Place the zscaler certificate at some place on your machine and mention that path at **sslcainfo**
  * Zscaler certificate can be fount at location - https://geico365.sharepoint.com/:f:/s/ITBilling/ErlWpOw8DOdHrJjBce04I80BUQ64Xk6M8j_dFgnUNzfiiQ?e=Hh5Jge
* Also in url, please put your ADO token. e.g. https://geico:7s7jxz3qelgyq@dev.azure.com/geico (This is an example token. You will have to generate yours). You can generate a new token here: https://dev.azure.com/geico/_usersSettings/tokens


You might need a restart after this and after that git clone/git push/git fetch/git commits all commands will work

# Build and Test
* To start Individual service:
    * Go inside service folder for example payment-vault/api and run ```go run .```
* To start infrastructure :
    * Go inside project folder and run ```make up```
* To stop infrastructure :
    * Go inside project folder and run ```make down```
* To start all the services (Run ```make build``` first if you are starting the services for the first time) :
    * Go inside project folder and run ```make start```
* To build all the services :
    * Go inside project folder and run ```make build```
* To build and start all the services :
    * Go inside project folder and run ```make build_start```
* To stop all the services :
    * Go inside project folder and run ```make stop```
* To run all the unit test cases in all the services :
    * Go inside project folder and run ```make test```
    * If all the unit test pass the script will output  ```All Unit Tests Passed```
    * Else the script will output ```One or more Test Unit Test failed, please check test_results inside testing folder```
    * All the test results can be found inside ```testing/test_results.out``` file
* To run coverage reports for the application which has integration test 
    * Go inside project folder and run ```make coverage```
    * The coverage reports will be generated in /coverageTmp folder, open the html file in your browser to view which lines are covered.

As a standard all the components should expose REST APIs on port 30000 and all the gRPC APIs on port 5051. 
To have all the services running locally on development computer we need to map all the ports to local network. This mapping will be done in docker-compose as below :
* Administrator
    * REST - 8081
    * gRPC - 9091
* Vault
    * REST - 8082
    * gRPC - 9092
* Executor
    * REST - 8083
    * gRPC - 9093
* Task Manager
    * REST - 8084
    * gRPC - 9094
* Config Manager
    * REST - 8085
    * gRPC - 9095    


# API Server
### Developer Desktop
* If API server has been started using ```go run .``` from command line or using a debugging session the service will start of port 30000 and can be accessed using http://localhost:30000 , only one service can be started using this method
* IF API server has been started using ```make start``` all the services will start and can be accessed at the urls below
    - Administrator : http://localhost:8081
    - Vault : http://localhost:8082
    - Executor : http://localhost:8083
    - Task Manager : http://localhost:8084
    - Config Manager : http://localhost:8085
### DV1 Environment
* API Servers can be reached at the URLs below
    - Administrator : https://bilpmt-admapi-dv.geico.net/
    - Vault : https://bilpmt-vltapi-dv.geico.net/
    - Executor : https://bilpmt-extapi-dv.geico.net/
    - Task Manager : https://bilpmt-tskapi-dv.geico.net/
    - Config Manager : https://bilpmt-cfgapi-dv.geico.net/
* Services that are exposed to external clients are protected by APISIX gateway
* It handles things like Authentication, Authorization and Rate limiting for us
* More details : https://geico365.sharepoint.com/sites/AsynchronousIntegration/SitePages/APISIX.aspx
* Calls to services like Administrator API, Task MAnager API and Vault API need a Bearer Token.
* Use below sample curl command to generate a Bearer Token for Task Manager API, The token should be placed in Authorization Tab of postman
* Command can be imported in postman from File->Import
```
  curl --location --request GET 'https://login.microsoftonline.com/7389d8c0-3607-465c-a69f-7d4426502912/oauth2/v2.0/token' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--header 'Accept: application/json' \
--header 'Cookie: fpc=Aj4r1nydek1BpoxuVSbsgLey9M0WAQAAAH826d0OAAAA; stsservicecookie=estsfd; x-ms-gateway-slice=estsfd; fpc=Aj4r1nydek1BpoxuVSbsgLcuG8pwAgAAAL0g7N0OAAAA; x-ms-gateway-slice=estsfd; fpc=Aj4r1nydek1BpoxuVSbsgLcuG8pwAQAAAKKS8t0OAAAA; stsservicecookie=estsfd; x-ms-gateway-slice=estsfd' \
--data-urlencode 'client_id=f58baa19-b547-490f-954d-f2ff0d8beb6b' \
--data-urlencode 'client_secret=uhi8Q~.LJ9vFzIgkxhgY9uuD__kN8nhcxhbpgcpG' \
--data-urlencode 'scope=api://tskapi-bilpmt-dv-v1/.default' \
--data-urlencode 'grant_type=client_credentials'
  ```
 
# PostgreSQL
### Developer Desktop
* A locally running postgreSQL instance inside a docker container along with pgAdmin can be quickly spun up using ```make up```
* Please do not install a postgreSQL instance locally
* pgAdmin instance can be accessed at http://localhost:8888/
* Use below credentials to log in to pgAdmin
    - email : test@geico.com
    - password : ```admin```
* Once logged-in register the PostgreSQL instance using below details (Servers -> right click -> register)
    - Name : athena
    - Hostname : project-postgres-1
    - Port : 5432
    - Username : nymeria
    - Password : ```zCXbDWt4H7%i^e```
* You will need to create the databases and tables for each service. The schemas can be found in `service-name\common\database-schema\serice_name.sql` e.g. `payment-vault\common\database-schema\payment_vault.sql`.
### DV1 Environment
* PostgreSQL instance is managed by PAAS Team
* They have installed an instance of pgAdmin for us
* pgAdmin instance can be accessed at https://pgadmin.10.250.46.49.sslip.io/login
* Ask someone from DevOps team to give your email id access to DV1 database
* Use below credentials to log in to pgAdmin
    - email : your geico email address
    - password : ```Pga123```

* Use below host names to connect to databases

* Read Only:
    | Component  | Host Name |
    | ------------- | ------------- |
    | administrator  | bilpmt-pmtadmin-dv-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | vault  | bilpmt-pmtvault-dv-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | executor  | bilpmt-pmtexecutor-dv-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | taskmanager  | bilpmt-taskmanager-dv-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | datahub  | bilpmt-datahub-dv-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | congigmanager  | bilpmt-configmanager-dv-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |

* Read and Write:
    | Component  | Host Name |
    | ------------- | ------------- |
    | administrator  | bilpmt-pmtadmin-dv-e1-pgcluster-rw.postgresqldelivery.svc.cluster.local  |
    | vault  | bilpmt-pmtvault-dv-e1-pgcluster-rw.postgresqldelivery.svc.cluster.local  |
    | executor  | bilpmt-pmtexecutor-dv-e1-pgcluster-rw.postgresqldelivery.svc.cluster.local  |
    | taskmanager  | bilpmt-taskmanager-dv-e1-pgcluster-rw.postgresqldelivery.svc.cluster.local  |
    | datahub  | bilpmt-datahub-dv-e1-pgcluster-rw.postgresqldelivery.svc.cluster.local  |
    | congigmanager  | bilpmt-configmanager-dv-e1-pgcluster-rw.postgresqldelivery.svc.cluster.local  |

* Once logged in, register the instance using below details (Servers -> right click -> register)
    - Name - {Can be any meaningful name} e.g. DV1-adm for administrator
    - Hostname - pick the value from above table
    - Username - Your U code or A code in lower case. eg. u59c35 or a12345
    - Password - your network password (the one used to login to your machine)
    - Click on Submit and DV1 database will be setup and you can perform read operations on the tables.
    - Repeat the above steps for all component databases.

### UT1 Environment
* To access UT1 databases, you have to be part of **PL-IT-Billing-PayRec-Payments-Dev** role. Drop an email to <ASDBillingDevOpsL2@geico.com> and they will raise an ARC request to get you added to this role. This access will be READONLY access.
* Follow the same steps as DV1 to log into pgAdmin but use below host names to connect to databases

    | Component  | Host Name |
    | ------------- | ------------- |
    | administrator  | entbill-pmtadmin-ut-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | vault  | bilpmt-pmtvault-ut-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | executor  | bilpmt-pmtexecutor-ut-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | taskmanager  | bilpmt-taskmanager-ut-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |
    | datahub  | bilpmt-datahub-ut-e1-pgcluster-r.postgresqldelivery.svc.cluster.local |
    | congigmanager  | bilpmt-configmanager-ut-e1-pgcluster-r.postgresqldelivery.svc.cluster.local  |

# Kafka
### Developer Desktop
* A locally running Kafka instance inside a docker container along with Kafka UI can be quickly spun up using ```make up```
* Please do not install a kafka instance locally
* Kafka UI instance can be accessed at http://localhost:9085/
* Add the Kafka cluster instance using below details
    - Bootstrap Servers : kafka
    - Port : 9092
### Creating Local Kafka Topics
* Ensure kafka is running locally (```make up```)
* Within the dashboard, click **local** -> **Topics**
* In the upper right, click **Add a Topic**
* Enter in the desired topic name, and use the value ```1``` for **Number of Paritions** (topics for service workers found in the worker/config directory within ```appsettings.json```)
### DV1 Environment
* Kafka instance is managed by PAAS Team
* There is no centrally managed instance of Kafka UI ay the moment
* Use the Kafka UI instance you have running in the docker, it can be accessed at http://localhost:9085/
* Add the DV1 Kafka cluster instance using below details
    - Bootstrap Servers : kafka-bootstrap.fburg-nonprod-01.kafka.geico.net
    - Port : 443
    - Truststore
        - Truststore Location : Use /project/keys/client.truststore
        - Truststore Password : ```123456```
    - Authentication
        - security.protocol : mTLS
        - ssl.keystore.location : Use /project/keys/client.keystore
        - ssl.keystore.password : ```Bilpmt-dv0523@```
* If you are intrested in how above files are being generated, a sparse documentation can be found here : https://geico365.sharepoint.com/sites/PaaS_KB/SitePages/How-to-connect-to-Kafka-Onprem.aspx

# Titan
* Titan is GEICO's Unified Observability Platform built on top of LGTM stack
* If you are intrested more details can be found here : https://geico365.sharepoint.com/sites/GC-OB/SitePages/Titan%20_%20GEICO%27s%20Unified%20Observability%20Platform.aspx
* It can be accessed at  https://titan.geicoddc.net/
* Select 'Sign in with Microsoft'
* To access logs go to Home -> Explore
* Environment can be selected using a dropdown at the top for example for DV1, please select onprem-logs-np
* Use filters like componentid and time range to see the desired logs
* Component Ids for our services
    - Administrator API : admapi
    - Administrator Worker : admwrk
    - Executor API : extapi
    - Executor Worker : extwrk
    - Vault API : vltapi
    - Vault Worker : vltwrk
    - Task Manager API : tskapi
    - Task Manager Worker : tskwrk
    - Configuration Manager API : cfgapi
    - Data Hub Worker : dthwrk

* For creating dashboards in Titan, you need to be part of this AD group - ENT-ASG-TITANBILLING-PD-USER. Please raise an ARC ticket to do so.

# Coding Standards and Guidelines
**Convention**

* Follow general guidelines from here - https://go.dev/doc/effective_go
* Folder and file names should be all lowercase.
* For folder names use - to seprate out words
* For file names use _ to seprate out words
* Package name should be short and match the folder name.
* For package names use - to seprate out words
* Package should not expose anything that is not used outside.
* Keep the public members on the top of the package and privates ones at the bottom.

**Lint**

We are using golangci-lint for code stati check, the CI will warn any lint error from all the modules, so it is suggested to run lint by `make lint` locally before pushing the code, it will install the lint tools and run check locally.
The rules for lint is in the `.golangci.yml` file. To use the same configuration and see lint errors in VSCode while developing, add `"go.lintTool": "golangci-lint",` in your project folder `.vscode/settings.json` file after running linting once.

After pushing the code to the repo, our CI workflow will run same lint check. It will show warnings in the step and PR review page. It is suggested to fix all the issues before merging to the main, or edit rules in the `.golangci.yml`.

# Code Coverage
We are using builtin *go test* and *go tool cover* to generate code coverage report from.
By executing `make coverage`, it will get the generate reports under /coverageTmp for the applications **which have integration test cases**. For detail implementation, read */project/scripts/test/run-coverage.sh*

After merging your code to main branch, the same workflow will be executed and the reports will be stored in specific branch *gh-pages*. You can display the coverage rate badge for each application by putting the image link on top of the README.md

# Database Schema Standards and Guidelines

* ### Naming Conventions
    * Database and Table names will be all in lowercase. For example, Database payment_vault has tables persona, product_details.
    * Table and column names should be kept meaningful and self explanatory.
    * Tables will be created under public scehma ONLY.

* ### Table and Column Names:
    * Separate words with underscores for readability (e.g., product_details).
    * Avoid using reserved keywords as table or column names.
    * Use bigint for all integer type columns
    * Choose appropriate data types for columns to ensure data integrity and efficiency.
    * Column names should be in Pascal Case. To enforce this, use quotes when defining the column names. e.g "UserId"

* ### Primary/Foreign Keys and Constraints:
    * Each table should have a primary key column. 
    * If the PK column is auto-generated and bigint, then always increment the value by 1 and start by 1
    * Use foreign keys to establish relationships between tables.
    * Ensure that foreign keys reference the primary key of the related table.
    * Use NOT NULL constraints for columns that should always have a value.
    * Use constraints to enforce data integrity rules. Use UNIQUE constraints to enforce uniqueness of values in a column.


# PR Standards and Guidelines
* Please use the below summary formal while raising a pull request
```
## Summary
* Use the bullet points to describe the change
* List all the impacted files and what was changed in them

## Testing
* List the ways you used to test the code, if it helps attatch screenshots

```
* Every PR should have at least two approvals before merging
* Every PR should have a appropriate unit test coverage that covers all the major flows.
* PR without unit test cases should only be approved under special circumstances (hotfix in production)
* Please try to raise small atomic PRs to get feedback early and dont wait till the end and raise a single huge PR
* PRs larger than 500 LOC are not allowed
* If the changes need any updates/addition to documentation (HLD, LLD etc). Please update the documentation before raising the PR and mention the same in the PR
* Please link the US or issue with the PR is available.
* If the PR is for a user story create the branch in format of
```alias/US#``` for example ```abtripathi/87654567``` if a user story
* Every PR must include the User Story Number (US#) witihn the PR title and/or body in the format ```AB#US#```, for example ```AB#8867569```. This is required to ensure the **_check-pr_** job will succeed.
* Till Unit Test run is intrigrated in CI pipeline, pleaser attatch a screenshot of successful test run
* Merge a PR only when all the checks pass in github
* When merging a PR, ensure to click **Squash and Merge** to ensure a clean Git history


# Logging
We have created a custom package for logging using Zerolog.
Its located inside pkg/logging

### Guidelines
* Dont use inbuilt go logger, use the logger defined in pkg/logger
* Always pass context as the first parameter to the loggger functions
* I am still not clear how we will manage and pass around context but for now dont pass nil for context please pass context.TODO()

### Usage
* Deffine a logger for your class
```
var log = logging.GetLogger("payment-vault-api")
```
* use the appropriate level to log the message
```
log.Info(context.TODO(), "Starting payment vault service on port %s", webPort)

log.Error(context.TODO(), err, "Error occurred when starting the service")
```
### Viewing Local Container Logs
* Use the following command to view logs for a specific container
```
docker-compose -f your-docker-compose-file.yml logs service-name
```
* example to see kafka logs:
```
docker-compose -f docker-compose-infra.yml logs kafka
```
* Use the following command to view all available service container names from provided docker-compose file
```
docker-compose -f your-docker-compose-file.yml config --services
```
* example to see all container service names from infra file:
```
docker-compose -f docker-compose-infra.yml config --services
```

# Configuration Management

We have created a custom package for configuration management using koanf.

Its located inside at github.com/geico-private/pv-bil-frameworks/config

### Usage

* Create configHandler instance using below code

```
configHandler := commonFunctions.NewConfigHandler()
```

* Once you have configHandler instance call following methods based on the type of config values you would expect to return
```
GetString(key string, defaultValue string) string
GetList(key string) []string
GetInt(key string, defaultValue int) int
GetBool(key string, defaultValue bool) bool
GetDuration(key string, defaultValue time.Duration) time.Duration
```
example to read a config value ```WebServer.Port``` from payment vault 
```
const  defaultWebPort = "80"
webPort := configHandler.GetString("WebServer.Port", defaultWebPort)
```

# Testing
Different features testify offers 
-   [Easy assertions](https://github.com/stretchr/testify?tab=readme-ov-file#assert-package)
-   [Mocking](https://github.com/stretchr/testify?tab=readme-ov-file#mock-package)
-   [Testing suite interfaces and functions](https://github.com/stretchr/testify?tab=readme-ov-file#suite-package)

  To install Testify, use :  go get github.com/stretchr/testify
  To install Mockery, use: go install github.com/vektra/mockery/v2@latest
  
  Sample to create a mock stub using mockery 
- command to run : mockery --all 
OR
mockery --dir C:\Users\a232539\Workspace1\PaymentPlatformMockUp\PaymentVault\db\managers  --name ProductDataManagerInterface --filename ProductDataManagerInterface.go
 -  https://geico.visualstudio.com/Billing/_git/plutus-poc?path=/akamble/PaymentPlatformMockUp/PaymentVault/mocks/ProductDataManagerInterface.go
 
Sample tests using testify
  - https://geico.visualstudio.com/Billing/_git/plutus-poc?path=/akamble/PaymentPlatformMockUp/PaymentVault/core/managers/productmanager_test.go
  - https://geico.visualstudio.com/Billing/_git/plutus-poc?path=/akamble/PaymentPlatformMockUp/PaymentVault/core/managers/productmanager_withoutsuites_test.go

```Recommendation - Do not use test suites feature of testify initially for writing simple test cases```

# Integration Test Guidelines

 Here are the steps to write and run integration test locally in the project:
    Create a normal test file with _test.go suffix, add //go:build integration at the top of the file, so that only this file will be compiled and run when the build tag includes integration.
    The config file for the local/CI environment is located at payment-administrator\worker\config\appsettings.json and payment-administrator\worker\config\secrets.json. Use global variable and init function to connect to the locally running infrastructure, so that we can reuse the connection in all the test cases. For sample integration test file, please refer to payment-administrator\integration-tests\administrator_integration_test.go

    1.Run individual component level integration test
        a)First bring down all the current running infrastructure – “ Make down ” command in the root directory of the project.
        b)Spin up the integration infrastructure locally by running the – “ Make integration_up “ in the root directory of the project that will create the Database & tables for all the components and creates the kafka topics locally
        c)To run all the test cases in that component , run the following command – “ make integration_<ComponentName>”  for example :-  make integration_administrator
        d)Tear down the infrastructure by running “Make integration_down” command in the root directory of the project after the test is done.

    2.Run all (including all the components) the integration tests from the project 
	Spin up the integration infrastructure locally by running the “Make integration_all” command – this command first will create the database and related Kafka topics and then it will start executing all the test cases within that component. Once all the test case execution is completed, it will delete the all the Kafaka topics which were created in that component and will move to next component, next it will start recreating all the topics which are required for the 2nd component (i.e. Executor) and followed by executing all the test cases. This step will keep repeating for all the components . Once all the components in the Test cases are completed, it will automatically bring down the integration infrastructure

    3.Push the code to the user branch, the CI will run the integration test with the above steps, and the test result should match the local test result.




# Database connection
We are using [pgx](https://github.com/jackc/pgx) postgres driver to connect to database
  ### Usage
- You can get a NewDbContext which has Database object from pkg/database/dbcontext.go. 
- This Database object can be used to connect to database and write queries.
-  Sample ping code usage ```database.NewDbContext().Database.Ping(ctx)```.
- Sample [documentation](https://github.com/jackc/pgx/wiki/Getting-started-with-pgx#using-a-connection-pool) to write a select query.
# Message Queue Connection

# Swagger 
* Download Swag for Go by using:  ```go install github.com/swaggo/swag/cmd/swag@v1.8.4```
The swag version 1.16.3 has issues, it is not generating the /HealthCheck endpoint. Things work with version 1.8.4
You can download the latest Swag for Go by using:  ```go install github.com/swaggo/swag/cmd/swag@latest```
* Add comments to your API source code, [See Declarative Comments Format](https://github.com/swaggo/swag#declarative-comments-format)
* Run ```swag init --parseDependency --parseInternal```
* Run it, and browser to http://localhost:30000/swagger/index.html, you can see Swagger 2.0 Api documents
* More [Documentation](https://github.com/swaggo/swag)
# Swagger GRPC
  ### Prerequiste
  - Install following go packages
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
  go install google.golang.org/protobuf/cmd/protoc-gen-go
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
 - [Install buf cli](https://buf.build/docs/installation)

### To create swagger files for GRPC (Example paymentpreference)
- Go to plutus\pkg\proto\paymentpreference path from powershell in case of windows OS and run ```buf generate``` command
### To setup a port for perticular grpc swagger 
- add grpcEntry in boot.yaml as shown here for paymentpreference
 ```
grpc:
  - name: paymentvaultgrpc         
    port: 3051                      
    enabled: true                 
    sw:
      enabled: true               
      jsonPath: "../../../pkg/proto/paymentpreference" 
 ```
 - add corresponding grpcEntry in main.go as shown here for payment preference 
 ```
grpcEntry := boot.GetEntry("paymentvaultgrpc").(*rkgrpc.GrpcEntry)
grpcEntry.AddRegFuncGrpc(func(server *grpc.Server) { proto.RegisterPaymentReferenceServiceServer(server, &internal.Server{}) })
grpcEntry.AddRegFuncGw(proto.RegisterPaymentReferenceServiceHandlerFromEndpoint)
 ```
 - Run the following swagger url for paymentpreference http://localhost:3051/sw
  # Voltage side car setup  
- To set up and run the Voltage sidecar, including the Voltage sidecar in infra setup, ensure the necessary configurations are in the geico-payment-platform\project\config\secrets.properties file. 

Then, start the infrastructure and application with:
* To start infrastructure :
    * Go inside project folder and run ```make up```
* To start integration test infrastructure :
    * Go inside project folder and run ```integration_up```

# Veracode
We are performing veracode scan before deploy to production environment following the enterprise policy. 
Since veracode scan works for one module each time only, we have a script called `project/veracode/prepare.sh` to build and vendor each individual module that can be deployed.
If we have new module to scan, we need to add that module name to $scanningMods in the script.

Current workflow set as follows: 
- Scan will automatically submmited for all modules from master branch everday midnight. Results needs to be check manually in veracode platform.
- Each release branch commit will trigger a scan as well, need to check manually as well.
- We can scan certain modules only by run the pipeline manually by uncheck other modules.

## Troubleshooting

### Voltage Sidecar Account Locked

**Issue:**  
Occasionally, the Voltage Sidecar account may get locked, which will prevent the application from functioning correctly.

**Solution:**  
If you encounter this issue, please follow these steps:

1. Send a message in the designated group channel(support-noc-operations), requesting the account to be unlocked.
2. Provide the necessary details, such as the environment (e.g., non-prod, prod), the application name, and any relevant logs or error messages.
3. Wait for confirmation from the support team that the account has been unlocked.


