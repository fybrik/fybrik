#!/usr/bin/python
import yaml

def main():
    # Read in the output of helm template . for m4d resources
    with open(r'm4d.yaml') as file:
        m4d_files = []
        for ln in file:
            if ln.startswith("# Source:"):
                filename = ln.split(': ')[1]
                filename = filename.split('/')[-1].strip()
                m4d_files.append(filename)
    
    # Filter out cluster-scoped resources and write to individual yaml files
    with open(r'm4d.yaml') as file:
        i = 0
        m4d_list = yaml.load_all(file, Loader=yaml.FullLoader)
        for yml in m4d_list:
            if yml['kind'] == 'MutatingWebhookConfiguration' or yml['kind'] == 'ClusterRole' or yml['kind'] == 'ClusterRoleBinding' or yml['kind'] == 'Namespace':
                outputFile = m4d_files[i]
                f = open(outputFile, "w")
                documents = yaml.dump(yml, f)
                f.close()
            i+=1

    # Read in the output of helm template . for m4d-crd resources
    with open(r'm4d-crd.yaml') as file:  
        crd_files = []
        for ln in file:
            if ln.startswith("# Source:"):
                filename = ln.split(': ')[1]
                filename = filename.split('/')[-1].strip()
                crd_files.append(filename)
    
     # Write CRDs to individual yaml files
    with open(r'm4d-crd.yaml') as file:
        i = 0
        m4d_list = yaml.load_all(file, Loader=yaml.FullLoader)
        for yml in m4d_list:
            outputFile = crd_files[i]
            f = open(outputFile, "w")
            documents = yaml.dump(yml, f)
            f.close()
            i+=1


if __name__ == "__main__":
    main()