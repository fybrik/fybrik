#!/usr/bin/python
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
                if not os.path.exists(os.path.dirname(outputPath)):
                    os.mkdir(os.path.dirname(outputPath))
                f = open(outputPath, "w") 
                documents = yaml.dump(yml, f)
                f.close()       

def createKustomizations():
    dirs = [name for name in os.listdir(".") if os.path.isdir(name)]
    for subdir in dirs:
        outputPath = os.path.join(subdir,'m4d-system','kustomization.yaml')
        f = open(outputPath,'w')
        kustomizationString = "---\napiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization" + "\n\nresources:\n"
        subdirPath = os.path.join(".", subdir,'m4d-system')
        for root, subdirs, files in os.walk(subdirPath):
            for file in files:
                if file != 'kustomization.yaml':
                    kustomizationString += "\t- " + file +'\n'
        f.write(kustomizationString)
        f.close()
    
      

def main():
    splitToYamls('m4d.yaml')
    splitToYamls('m4d-crd.yaml')
    print("Successfully split files into yamls and wrote to appropriate directories\n")   
    
    createKustomizations()
    print("Successfully created kustomization files in each subdirectory\n")   



    

if __name__ == "__main__":
    main()
    