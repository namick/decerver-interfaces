# For

The Legal Integrator is the package (which will be utilized by the deCerver) which will operate as the mechanism for integrating real world contracts with Eris deployed smart contracts.

## Notation

Throughout this repo (and the rest of Eris' work) the term `smart contracts` are used to denote scripted contracts deployed to a client's blockchain while the term `real world contracts` are used to denote a PDF of an electronically signed contract which is held within the client's IPFS layer and is referenced via its SHA1 hash. `lmd` is used to denote a contract template which is built in `legal_markdown` syntax.

## Deploy Sequence

* Signal sent to deCerver to deploy a smart contract to the blockchain.
  * Deploy params to the factories are registered via the API deployment call.
* deCerver signals to eth layer factory to deploy smart contract.
  * Address of smart contract returned to deCerver via eth
* deCerver signals to legal-integrator to deploy real world contract
  * Address of smart contract (along with other deploy params) are sent to the real world `lmd` factory
  * `lmd` factory spins off a real world contract which is "signed" by the company or initiating entity
  * legal-integrator signals to the deCerver that the real contract is deployed.
* deCerver signals to smart contract via eth that the contract status is placed into `offered` status
* deCerver signals to the TFA (two factor authentication) system that the contract is `offered`
  * the following params are sent to the TFA system:
    a. the address of the smart contract
    b. the IPFS reference hash of the real contract

## Acceptance Sequence

* TFA returns with `accepted` signal signed by the private key of the offeree of the contract.
* TFA signals to deCerver that the contract has been accepted
* deCerver signals to legal-integrator to electronically sign the PDF
  * legal-integrator returns the new (updated) reference hash within the IPFS system
* deCerver signals to eth layer with the updated hash of the signed PDF
* eth layer updates the contract status to be placed in `accepted` mode.

At this point the contract is viable or ('live').

## Rejection Sequence

* TFA returns with `rejected` signal signed by the private key of the offeree of the contract.
* TFA signals to deCerver that the contract has been rejected
* deCerver signals to legal-integrator to delete the PDF from the IPFS system
* deCerver signals to eth layer to suicide the contract (via a `rejected` function...?)

At this point the contract is purged.

## Time Out Sequence

Offers should have a TTL (time to live) but how to treat failure to respond with a valid signal from the TFA layer within that time is an open question. Ideas?