### What is this?
This is a small service that exposes a toy RESTful API. It is intended to be a template or example for how we write this sort of code in FleetServices. 
It also contains example configuration for other useful things such as:
1. Creating a Docker image
2. Launching the service image into our [sandbox](https://github.com/metal-toolbox/sandbox) environment. See the [helm](./helm) directory.
3. Typical pull-request workflow (linting, building, code-analysis)
4. Creating a release image of the software for deployment to Kubernetes.

### How do I use it?
That depends on what you want to do. 
- You can clone this repo as a template service, make your modifications, and push it to a new repo on Github.com.
- You can get a sense of how we handle common tasks (like using [gin](https://gin-gonic.com) or [zap](https://pkg.go.dev/go.uber.org/zap)) without being overwhelmed by details of a non-trivial service.
- You can propose new conventions (such as adding a client for [NATS](https://nats.io) or [FleetDB](https://github.com/metal-toolbox/fleetdb))
- You can launch this service into our [kind](https://kind.sigs.k8s.io) sandbox by doing `helm install skeleton-test helm` from the root of this repo. Port-forward to your local environment to test the API by hand, or configure service-to-service tests with other services in kind.

Much of the functionality is encapsulated into `Makefile` targets. On the one hand this is a pretty clear abuse of `make`, but on the other we do it in many other repositories. `make` will tab-prompt the user with potential targets (e.g. `build`, `image`, `push-sandbox-image` et al.)

I hope it serves to reduce the friction of getting a service into production for you.
