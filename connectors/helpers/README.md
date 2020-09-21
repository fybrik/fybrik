# Instructions for deployment on Openshift 4 Cluster
* Ensure Openshift Client is installed on the local machine. You can install oc client from this location: https://mirror.openshift.com/pub/openshift-v4/clients/ocp/. Please choose the latest version and install it in your machine. Ensure oc is installed in a location which is present in $PATH variable. Typical location of oc would be in /usr/local/bin/ folder.  

* Ensure Docker Desktop is installed and running on the local machine. This is needed as docker desktop hosts the built docker images in the local docker registry which is maintained by Docker Desktop application.

* In your browser please go to the openshift console and copy the login command. Paste the login command in the terminal. Typically the login command will be in the following form. 

        oc login --token=ACCESS_TOKEN --server=OC4SERVER:PORT

* Once the OC4 login is successful, change the oc project in the terminal to the desired project where the deployment should be carried out. If the project name on Openshift 4 is "irltest1" then run the following command with PROJECTNAME set to irltest1 
        
        oc project PROJECTNAME


* After the above steps, ensure oc is logged in correctly with the correct project by executing the following command. 

        oc status 

* Ensure you change all the yamls in the Openshift4-Deployment folder to have the correct PROJECTNAME in the image tag part of the yaml file 

* Now cd to the Openshift4-Deployment folder and execute the following script command and follow the instructions as given the script. 

        export KUBE_NAMESPACE=irltest1
        make deploy

* Once you are done with working with the cluster, please ensure you logout through the oc commandline using the following command:
        oc logout 

