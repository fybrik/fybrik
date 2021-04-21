#!/usr/bin/env python
import yaml
import os


def splitToYamls(yamlFile):
    with open(yamlFile, 'r') as file:
        m4d_list = yaml.load_all(file, Loader=yaml.FullLoader)
        for yml in m4d_list:
            if yml:
                filename = yml['metadata']['name']
                outputFolder = yml['kind'].lower()+'s'
                outputPath = os.path.join(outputFolder, 'm4d-system', filename + ".yaml")
                os.makedirs(os.path.dirname(outputPath), exist_ok = True)
                with open(outputPath,'w') as f:
                    yaml.dump(yml, f)    

def createKustomizations():
    dirs = [name for name in os.listdir(".") if os.path.isdir(name)]
    for subdir in dirs:
        outputPath = os.path.join(subdir,'m4d-system','kustomization.yaml')
        kustomizationString = "---\napiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization" + "\n\nresources:\n"
        subdirPath = os.path.join(".", subdir,'m4d-system')
        for root, subdirs, files in os.walk(subdirPath):
            for file in files:
                if file != 'kustomization.yaml':
                    kustomizationString += "  - " + file +'\n'
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
    