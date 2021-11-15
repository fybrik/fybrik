# Release Process

The process of creating a release is described in this document. Replace `X.Y.Z` with the version to be released.

## 1. Create a `releases/X.Y.Z` branch 

The `releases/X.Y.Z` branch should be created from a base branch. 

For major and minor releases the base is `master` and for patch releases (fixes) the base is `releases/X.Y.(Z-1)`.
A new patch release should be created before merging pull-requests as described in the next section.

You can do that [directly from GitHub](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-and-deleting-branches-within-your-repository#creating-a-branch).

## 2. Optionally create Pull Requests to `releases/X.Y.Z`

For any change that is required on top of the base you should create a Pull Request targeting the **new** release branch. 

Pull requests to the release branch may contain bug fixes, security fixes and updated documentation.

Collaborators with `write` permissions can cherry-pick a Pull Request that is merged to `master` into the release branch by commenting on the Pull Request with:

```bash
/cherry-pick branch=releases/X.Y.Z
```

You should ensure that all Pull Requests that target the release branch are reviewed and merged before proceeding to the next step.

## 3. Create a [new release](https://github.com/fybrik/fybrik/releases/new) 

Use `vX.Y.Z` tag and set `releases/X.Y.Z` as the target.

Ensure that the release notes explicitly mention upgrade instructions and any breaking change.

## 4. Update Helm charts

An automated Pull Request to update the Helm charts is created in [fybrik/charts](https://github.com/fybrik/charts/pulls). You must manually merge the Pull Request and ensure that a chart release is created after a couple of minutes.

