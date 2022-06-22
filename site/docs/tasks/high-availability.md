# High Availability

When having more then one manager instance running Active/Passive high availability method is used: only one of them is being the leader (doing all the exclusive tasks) and other instances waiting in standby mode in case the leader dies to take over the leader role.

The Active/Passive high availability is implemented with [Kubernetes leader election mechanism](https://itnext.io/leader-election-in-kubernetes-using-client-go-a19cbe7a9a85) and it is turned on by default. In this implementation a config-map resource called `fybrik-operator-leader-election` serves as a lock. The config-map also contains information about the chosen leader.

To change Fybrik manager number of replicas the following setting should be added to Fybrik helm chart deployment command `--set manager.replicaCount=<desired replica value>`.


