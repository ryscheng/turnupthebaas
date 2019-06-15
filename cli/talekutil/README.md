# TalekUtil

This utility helps with generating configuration files to synchronize a talek
deployment.

## Generating at Talek deployment

1. `talekutil --common --outfile common.json`
  - This will generate a JSON file template for the base of the deployment:
    How big of a database; how frequently reads and writes occur. These
    parameters should be in sync across all replics and clients in order for
    the deployment to work.
  - Modify the `common.json` file that results to suit your experiment / needs.
2. Each trust domain / replica then runs
  a. `talekutil --replica --incommon common.json --private --index <idx> --name <name> --address <addr> --outfile myreplica.json`
    - This generates keying material and the server configuration for that replica
  b. `talekutil --trustdomain --infile myreplica.json --outfile myreplica.pub.json`
    - This derives a sharable version for the replica that are aggregated.
3. The frontend / leader is given each of the `replica.pub.json` files.
4. `talekutil --client --infile common.json --trustdomains replica1.pub.json,replica2.pub.json,... --outfile talek.json`
  - This generates the final configuration distributed to clients and used by the frontend.
  - Edit talek.json to set `FrontendAddr` to the public facing host and port of the frontend.

## running

While the network should fail to make progress until all components are operational,
the most reliable structure for testing will be to start with replica, then the frontent,
and finally allow clients to start imposing load on the system.

Replicas are started using:
    `talekreplica --common common.json --config myreplica.json --listen <local interface:port>`

The frontend uses much the same data as a client:
    `talekfrontend --common common.json --config talek.json --listen <local interface:port>`

The clients only need the final file:
    `talekclient --config talek.json`
