#!/bin/bash
export git_user=${git_user}
export GH_TOKEN=${GH_TOKEN}
export github_workspace=${github_workspace}
export image_source_repo_username=${image_source_repo_username}
export image_source_repo_password=${ARTIFACTORY_APIKEY}
export run_tkn=${run_tkn:-0}
export skip_tests=${skip_tests:-false}
export cluster_scoped=${cluster_scoped:-false}
export github=${github:-github.ibm.com}
export image_repo="${image_repo:-image-registry.openshift-image-registry.svc:5000}"
export image_source_repo="${image_source_repo:-wcp-ibm-streams-docker-local.artifactory.swg-devops.com}"
export dockerhub_hostname="${dockerhub_hostname:-wcp-ibm-streams-docker-local.artifactory.swg-devops.com/pipelines-tutorial}"
export cpd_url=https://cpd-cpd4.apps.cpstreamsx4.cp.fyre.ibm.com
export git_url=git@${github}:IBM-Data-Fabric/mesh-for-data.git
export wkc_connector_git_url=git@${github}:ngoracke/WKC-connector.git
export cpd_password=password
export cpd_username=admin
export vault_plugin_secrets_wkc_reader_url=git@${github}:data-mesh-research/vault-plugin-secrets-wkc-reader.git
export git_url="git@${github}:IBM-Data-Fabric/mesh-for-data.git"
export wkc_connector_git_url="git@${github}:ngoracke/WKC-connector.git"
export vault_plugin_secrets_wkc_reader_url="git@${github}:data-mesh-research/vault-plugin-secrets-wkc-reader.git"

if [[ ! -z ${GH_TOKEN} ]]; then
    export git_url="https://${github}/IBM-Data-Fabric/mesh-for-data.git"
    export wkc_connector_git_url="https://${github}/ngoracke/WKC-connector.git"
    export vault_plugin_secrets_wkc_reader_url="https://${github}/data-mesh-research/vault-plugin-secrets-wkc-reader.git"
fi

echo "
## Git credentials

For authenticated registries, if you use a git token instead of ssh key, credentials will not be deleted when the run is complete (and therefore, you will not have to regenerate them when restarting tasks).
https://github.ibm.com/settings/tokens

export GH_TOKEN=xxxxxxx
export git_user=user@email.com
"
