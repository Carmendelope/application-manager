# Application Manager
​
The application manager is in charge of managing operations related to the application lifecycle and the connections between applications. ​​
The application manager receives messages from public-api as list/get descriptors and applications, deploy/undeploy instances, add and remove connections, etc. 
Some of these requests can be satisfied in this component and other ones should be sent to conductor and network-manager  

## Getting Started

​
### Prerequisites
​
Before installing this component, we need to have the following deployed:​

* system-model: Application manager access to the system model to request, update, delete application and connection entities. 
* conductor: required but not used
* nalej-bus: Application manager needs to connect to nalej-bus to send the messages related to the deploy and undeploy of an instance, 
and messages related to add and remove connections
​
### Build and compile
​
In order to build and compile this repository use the provided Makefile:
​
```
make all
```
​
This operation generates the binaries for this repo, download dependencies,
run existing tests and generate ready-to-deploy Kubernetes files.
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

Some integration test are included. To execute those, setup the following environment variables. The execution of 
integration tests may have collateral effects on the state of the platform. DO NOT execute those tests in production.
​

​The following table contains the variables that activate the integration tests
 
 | Variable  | Example Value | Description |
 | ------------- | ------------- |------------- |
 | RUN_INTEGRATION_TEST  | true | Run integration tests |
 | IT_SM_ADDRESS  | localhost:8800 | System Model Address |
 | IT_CONDUCTOR_ADDRESS | localhost:5000 | Conductor Address |
 | IT_BUS_ADDRESS | localhost:6655 | Nalej-Bus address |

​
## Contributing
​
Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.
​
​
## Versioning
​
We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/nalej/application-manager/tags). 
​
## Authors
​
See also the list of [contributors](https://github.com/nalej/application-manager/contributors) who participated in this project.
​
## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.



