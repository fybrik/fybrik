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
                outputPath = os.path.join(outputFolder, filename + ".yaml")
                if not os.path.exists(os.path.dirname(outputPath)):
                    os.mkdir(os.path.dirname(outputPath))
                f = open(outputPath, "w") 
                documents = yaml.dump(yml, f)
                f.close()          

def main():
    splitToYamls('m4d.yaml')
    splitToYamls('m4d-crd.yaml')

    print("Successfully split files into yamls and wrote to appropriate directories")

if __name__ == "__main__":
    main()
    