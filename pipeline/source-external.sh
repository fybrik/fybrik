export run_tkn=${run_tkn:-0}
export is_kind=${is_kind:-true}
export skip_tests=${skip_tests:-false}
export GH_TOKEN=${GH_TOKEN:-fake}
export image_source_repo_password=fake
export cluster_scoped=${cluster_scoped:-false}
export git_user=${git_user:-fake@fake.com}
export github=${github:-github.com}
export github_workspace=${github_workspace}
export image_source_repo_username=${image_source_repo_username}
export image_repo="${image_repo:-kind-registry:5000}"
export image_source_repo="${image_source_repo:-fake.com}"
export dockerhub_hostname="${dockerhub_hostname:-docker.io}"
export git_url="https://github.com/fybrik/fybrik.git"
echo "
## Git credentials
For authenticated registries, if you use a git token instead of ssh key, credentials will not be deleted when the run is complete (and therefore, you will not have to regenerate them when restarting tasks).
https://github.com/settings/tokens
export GH_TOKEN=xxxxxxx
export git_user=user@email.com
"
