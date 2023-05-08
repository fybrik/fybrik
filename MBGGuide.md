Create two clusters.

Install MBG in both clusters:

- Clone the `mbg-agent` repository and switch to `draft-branch` branch https://github.ibm.com/mcnet-research/mbg-agent/tree/draft-branch.
- In the main cluster install and run MBG using the command `python3 tests/iperf3/kind/simple_test.py`.
- In the remote cluster install and run MBG using the command `python3 tests/iperf3/kind/simple_test2.py`.
- In the remote cluster: Get the MBG IP by running `kubectl exec -i {mbg2Pod} -- cat /root/.mbg/mbgApp`, store it as `mbg2IP`.
- In the main cluster: Add a peer using the following command `kubectl exec -i {mbgctl1Pod} -- ./mbgctl add peer --id mbg2 --target {mbg2IP} --port 30443`.
- In the main cluster: Run hello command to connect to the remote MBG using the command `kubectl exec -i {mbgctl1Pod} -- ./mbgctl hello`.

Install fybrik in two clusters with the values of the names of MBG pods: `--set cluster.mbgPodName={mbgPodName} --set cluster.mbgCtlPodName={mbgCtlPodName} --set cluster.mbgNamespace={mbgNamespace}`.

Try the chaining sample of fybrik https://fybrik.io/dev/samples/chaining-sample/.