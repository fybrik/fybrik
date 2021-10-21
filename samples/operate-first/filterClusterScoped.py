#!/usr/bin/env python
import yaml
import os


def splitToYamls(yamlFile):
    with open(yamlFile, 'r') as file:
        fybrik_list = yaml.load_all(file, Loader=yaml.FullLoader)
        for yml in fybrik_list:
            if yml:
                filename = yml['kind'].lower()
                resourceName = yml['metadata']['name']
                apiVersion = yml['apiVersion'].split('/')[0]
                kind = yml['kind'].lower()+'s'
                outputPath = os.path.join(apiVersion, kind, resourceName, filename + ".yaml")
                os.makedirs(os.path.dirname(outputPath), exist_ok = True)
                with open(outputPath,'w') as f:
                    yaml.dump(yml, f)    

def createKustomizations():
    dirs = [name for name in os.listdir(".") if os.path.isdir(name)]
    apiVersions = []
    for folder in dirs:
        apiVersions.append(os.path.join(".", folder))
    for apiVersion in apiVersions:
        resourceKinds = os.listdir(apiVersion)  
        for resourceKind in resourceKinds:
            resourceNames = os.listdir(os.path.join(apiVersion, resourceKind))
            for resource in resourceNames:
                resourcePath = os.path.join(apiVersion, resourceKind, resource)
                kustomizationString = "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization" + "\nresources:\n"
                yamlFile = resourceKind[:-1] + '.yaml'
                kustomizationString += "- " + yamlFile +'\n'
                outputPath = os.path.join(resourcePath,'kustomization.yaml')
                with open (outputPath, 'w') as f:
                    f.write(kustomizationString)

def main():
    splitToYamls('fybrik-crd.yaml')
    print("Successfully split files into yamls and wrote to appropriate directories\n")   
    
    createKustomizations()
    print("Successfully created kustomization files in each subdirectory\n")   

if __name__ == "__main__":
    main()
    