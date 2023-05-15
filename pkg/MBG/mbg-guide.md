Create two clusters.

Install MBG in both clusters:

- In the main cluster install and run MBG using the command `python3 pkg/MBG/install-mbg.py --mbgname mbg1 --mbgctlname mbgctl1  --certificate ./mtls/mbg1.crt --key ./mtls/mbg1.key`.
- In the remote cluster install and run MBG using the command `python3 pkg/MBG/install-mbg.py --mbgname mbg2 --mbgctlname mbgctl2  --certificate ./mtls/mbg2.crt --key ./mtls/mbg2.key`.
- In the main cluster Store the name of the MBG and MBG control pod `mbg-deployment-XXXX` and `mbgctl-deployment-XXXX` in `MBG_POD1` and `MBG_CTL_POD1` respectively.
- In the remote cluster Store the name of the MBG and MBG control pod `mbg-deployment-XXXX` and `mbgctl-deployment-XXXX` in `MBG_POD2` and `MBG_CTL_POD2` respectively.
- In the remote cluster: Get the MBG IP by running `kubectl exec -i ${MBG_POD2} -- cat /root/.mbg/mbgApp`, store it in `MBG_IP2`.
- In the main cluster: Add a peer using the following command `kubectl exec -i ${MBG_CTL_POD1} -- ./mbgctl add peer --id mbg2 --target ${MBG_IP2} --port 30443`.
- In the main cluster: Run hello command to connect to the remote MBG using the command `kubectl exec -i ${MBG_CTL_POD1} -- ./mbgctl hello`.

[Install fybrik](https://fybrik.io/dev/tasks/multicluster/) in the two clusters with the values of the names of MBG pods:

- In the main cluster use `--set cluster.mbg.mbgPodName=${MBG_POD1} --set cluster.mbg.mbgCtlPodName=${MBG_CTL_POD1} --set cluster.mbgNamespace=default`.
- In the remote cluster use `--set cluster.mbg.mbgPodName=${MBG_POD2} --set cluster.mbg.mbgCtlPodName=${MBG_CTL_POD2} --set cluster.mbgNamespace=default`.

Try the chaining sample of fybrik https://fybrik.io/dev/samples/chaining-sample/.
