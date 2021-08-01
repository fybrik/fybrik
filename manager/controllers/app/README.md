The controllers are:
* FybrikApplication controller - creates/updates a blueprint specification from the information in the FybrikApplication and from the policy compiler
* Blueprint controller - reconciles the desired cluster state by orchestrating a blueprint
It also configures lower level things like isolation policies.