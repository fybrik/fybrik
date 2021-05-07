#!/usr/bin/env python
import yaml
import os


def splitToYamls(yamlFile):
    with open(yamlFile, 'r') as file:
        m4d_list = yaml.load_all(file, Loader=yaml.FullLoader)
        for yml in m4d_list:
            if yml:
                filename = yml['kind'].lower()
                resourceName = yml['metadata']['name']
                outputFolder = yml['kind'].lower()+'s'
                outputPath = os.path.join(outputFolder, resourceName, filename + ".yaml")
                os.makedirs(os.path.dirname(outputPath), exist_ok = True)
                with open(outputPath,'w') as f:
                    yaml.dump(yml, f)    

def createKustomizations():
    dirs = [name for name in os.listdir(".") if os.path.isdir(name)]
    subdirPaths = []
    for folder in dirs:
        subdirPaths.append(os.path.join(".", folder))
    for path in subdirPaths:
        resources = os.listdir(path)  
        for resource in resources:
            resourcePath = os.path.join(path, resource)
            kustomizationString = "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization" + "\nresources:\n"
            yamlFile = path[2:-1] + '.yaml'
            kustomizationString += "  - ./" + yamlFile +'\n'
            outputPath = os.path.join(resourcePath,'kustomization.yaml')
            with open (outputPath, 'w') as f:
                f.write(kustomizationString)

def main():
    splitToYamls('m4d.yaml')
    splitToYamls('m4d-crd.yaml')
    print("Successfully split files into yamls and wrote to appropriate directories\n")   
    
    createKustomizations()
    print("Successfully created kustomization files in each subdirectory\n")   



    

if __name__ == "__main__":
    main()
    