################################################################
#Name: Simple iperf3  test
#Desc: create 2 kind clusters :
# 1) MBG and iperf3 client
# 2) MBG and iperf3 server    
###############################################################
#!/usr/bin/env python3
import argparse
import json
import os
import time
from colorama import Fore
from colorama import Style
import subprocess as sp


def startKindClusterMbg(mbgName, mbgctlName, mbgcPortLocal, mbgcPort, mbgDataPort, dataplane, mbgcrtFlags):
    mbgKindIp = getKindIp(mbgName)
    podMbg, podMbgIp    = buildMbg(mbgName)
    destMbgIp          = f"{podMbgIp}:{mbgcPortLocal}"
    runcmd(f"kubectl create service nodeport mbg --tcp={mbgcPortLocal}:{mbgcPortLocal} --node-port={mbgcPort}")
    printHeader(f"\n\nStart {mbgName} (along with PolicyEngine)")
    startcmd= f'{podMbg} -- ./mbg start --id "{mbgName}" --ip {mbgKindIp} --cport {mbgcPort} --cportLocal {mbgcPortLocal}  --externalDataPortRange {mbgDataPort}\
    --dataplane {dataplane} {mbgcrtFlags} --startPolicyEngine={True} --logFile={True}'
    runcmdb("kubectl exec -i " + startcmd)
    mbgctlPod, _ = buildMbgctl(mbgctlName)
    runcmdb(f'kubectl exec -i {mbgctlPod} -- ./mbgctl create --id {mbgctlName} --mbgIP {destMbgIp}  --dataplane {dataplane} {mbgcrtFlags} ')

def buildMbg(name):
    runcmd(f"kubectl apply -f pkg/MBG/mbg-role.yaml")
    runcmd(f"kubectl create -f pkg/MBG/mbg.yaml")
    waitPod("mbg")
    podMbg, mbgIp= getPodNameIp("mbg")
    return podMbg, mbgIp

def buildMbgctl(name):
    runcmd(f"kubectl create -f pkg/MBG/mbgctl.yaml")
    waitPod("mbgctl")
    name,ip= getPodNameIp("mbgctl")
    return name, ip

def getKindIp(name):
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

    printHeader("\n\nStart Kind Test\n\n")
    printHeader("Start pre-setting")

    dataplane = args["dataplane"]
    #MBG1 parameters 
    mbg1DataPort    = "30001"
    mbg1cPort       = "30443"
    mbg1cPortLocal  = "8443"
    mbg1crtFlags    = "--rootCa ./mtls/ca.crt " + "--certificate " + args["certificate"] + " --key " + args["key"]
    mbg1Name        = args["mbgname"]
    mbgctl1Name     = args["mbgctlname"]
    
    ### Build MBG in Kind clusters environment 
    startKindClusterMbg(mbg1Name, mbgctl1Name, mbg1cPortLocal, mbg1cPort, mbg1DataPort, dataplane ,mbg1crtFlags)

