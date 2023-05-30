import argparse
import json
import os
import time
from colorama import Fore
from colorama import Style
import subprocess as sp


def startMbg(mbgName, mbgctlName, mbgcPortLocal, mbgcPort, mbgDataPort, dataplane, mbgcrtFlags):
    mbgNodeIp = getNodeIp(mbgName)
    podMbg, podMbgIp    = buildMbg(mbgName)
    destMbgIp          = f"{podMbgIp}:{mbgcPortLocal}"
    runcmd(f"kubectl create service nodeport mbg --tcp={mbgcPortLocal}:{mbgcPortLocal} --node-port={mbgcPort}")
    printHeader(f"\n\nStart {mbgName} (along with PolicyEngine)")
    startcmd= f'{podMbg} -- ./mbg start --id "{mbgName}" --ip {mbgNodeIp} --cport {mbgcPort} --cportLocal {mbgcPortLocal}  --externalDataPortRange {mbgDataPort}\
    --dataplane {dataplane} {mbgcrtFlags} --startPolicyEngine={True} --logFile={True}'
    runcmdb("kubectl exec -i " + startcmd)
    mbgctlPod, _ = buildMbgctl(mbgctlName)
    runcmdb(f'kubectl exec -i {mbgctlPod} -- ./mbgctl create --id {mbgctlName} --mbgIP {destMbgIp}  --dataplane {dataplane} {mbgcrtFlags} ')

def buildMbg(name):
    runcmd(f"kubectl apply -f third_party/MBG/mbg-role.yaml")
    runcmd(f"kubectl create -f third_party/MBG/mbg.yaml")
    waitPod("mbg")
    podMbg, mbgIp= getPodNameIp("mbg")
    return podMbg, mbgIp

def buildMbgctl(name):
    runcmd(f"kubectl create -f third_party/MBG/mbgctl.yaml")
    waitPod("mbgctl")
    name,ip= getPodNameIp("mbgctl")
    return name, ip

def getNodeIp(name):
    clJson=json.loads(sp.getoutput(f' kubectl get nodes -o json'))
    ip = clJson["items"][0]["status"]["addresses"][0]["address"]
    return ip

def getPodNameIp(app):
    podName = getPodNameApp(app)
    podIp   =  getPodIp(podName)  
    return podName, podIp

def getPodNameApp(app):
    cmd=f"kubectl get pods -l app={app} "+'-o jsonpath="{.items[0].metadata.name}"'
    podName=sp.getoutput(cmd)
    return podName

def getPodName(prefix):
    podName=sp.getoutput(f'kubectl get pods -o name | fgrep {prefix}| cut -d\'/\' -f2')
    return podName

def printHeader(msg):
    print(f'{Fore.GREEN}{msg} {Style.RESET_ALL}')

def getPodIp(name):
    name=getPodName(name)
    podIp=sp.getoutput(f"kubectl get pod {name}"+" --template '{{.status.podIP}}'")
    return podIp

def waitPod(name):
    time.sleep(2) #Initial start
    podStatus=""
    while(podStatus != "Running"):
        cmd=f"kubectl get pods -l app={name} "+ '--no-headers -o custom-columns=":status.phase"'
        podStatus =sp.getoutput(cmd)
        if (podStatus != "Running"):
            print (f"Waiting for pod {name} to start current status: {podStatus}")
            time.sleep(7)
        else:
            time.sleep(5)
            break

def runcmd(cmd):
    print(f'{Fore.YELLOW}{cmd} {Style.RESET_ALL}')
    os.system(cmd)

def runcmdb(cmd):
    print(f'{Fore.YELLOW}{cmd} {Style.RESET_ALL}')
    os.system(cmd + ' &')
    time.sleep(7)

############################### MAIN ##########################
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Description of your program')
    parser.add_argument('-d','--dataplane', help='choose which dataplane to use mtls/tcp', required=False, default="mtls")
    parser.add_argument('--certificate', required=False, default="./mtls/mbg1.crt")
    parser.add_argument('--key', required=False, default="./mtls/mbg1.key")
    parser.add_argument('--mbgname', help='MBG name', required=False, default="mbg1")
    parser.add_argument('--mbgctlname', help='MBG control name', required=False, default="mbgctl1")

    args = vars(parser.parse_args())

    printHeader("Start installing MBG")

    dataplane = args["dataplane"]
    # MBG parameters 
    mbgDataPort    = "30001"
    mbgcPort       = "30443"
    mbgcPortLocal  = "8443"
    mbgcrtFlags    = "--rootCa ./mtls/ca.crt " + "--certificate " + args["certificate"] + " --key " + args["key"]
    mbg1Name       = args["mbgname"]
    mbgctlName     = args["mbgctlname"]
    
    startMbg(mbg1Name, mbgctlName, mbgcPortLocal, mbgcPort, mbgDataPort, dataplane ,mbgcrtFlags)

