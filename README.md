# Application Manager
​
The application manager handles the operations related to the application lifecycle and the connections between applications. ​​
The application manager receives messages from `public-api` (list/get descriptors and applications, deploy/undeploy instances, add/remove connections, etc.) 
Some of these requests can be satisfied in this component and others are sent to `conductor` and `network-manager`.  

## Getting Started

​
### Prerequisites
​
Before installing this component, we need to have the following deployed:​

* [`system-model`](https://github.com/nalej/system-model): responsible for requesting, updating and deleting application and connection entities. 
* [`conductor`](https://github.com/nalej/conductor): required but not used (this dependency will be removed in the next release).
* [`nalej-bus`](https://github.com/nalej/nalej-bus): necessary for sending the messages related to deploying and undeploying an instance, and to adding and removing connections.
​
### Build and compile
​
In order to build and compile this repository use the provided Makefile:
​
```
make all
```
​
This operation generates the binaries for this repo, downloads the required dependencies, runs existing tests and generates ready-to-deploy Kubernetes files.
​
### Run tests
​
Tests are executed using Ginkgo. To run all the available tests:
​
```
make test
```

### Update dependencies
​
Dependencies are managed using Godep. For an automatic dependencies download use:
​
```
make dep
```
​
In order to have all dependencies up-to-date run:
​
```
dep ensure -update -v
```

### Integration tests

Some integration test are included. To execute them, setup the following environment variables. The execution of 
integration tests may have collateral effects on the state of the platform. DO NOT execute those tests in production.
​

​The following table contains the variables that activate the integration tests
 
 | Variable  | Example Value | Description |
 | ------------- | ------------- |------------- |
 | RUN_INTEGRATION_TEST  | true | Run integration tests |
 | IT_SM_ADDRESS  | localhost:8800 | System Model Address |
 | IT_CONDUCTOR_ADDRESS | localhost:5000 | Conductor Address |
 | IT_BUS_ADDRESS | localhost:6655 | Nalej-Bus Address |

​
## Contributing
​
Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.
​
​
## Versioning
​
We use [SemVer](http://semver.org/) for versioning. For the available versions, see the [tags on this repository](https://github.com/nalej/application-manager/tags). 
​
## Authors
​
See also the list of [contributors](https://github.com/nalej/application-manager/contributors) who participated in this project.
​
## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.



