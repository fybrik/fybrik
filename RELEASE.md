# Release Process

The process of creating a release is described in this document. Replace `X.Y.Z` with the version to be released.

## 1. Create a `releases/X.Y.Z` branch 

The `releases/X.Y.Z` branch should be created from a base branch. For minor releases the base is `master` and for patch releases the base is `releases/X.Y.(Z-1)`.

## 2. Optionally create Pull Requests to `releases/X.Y.Z`

For any change that are required on top of the base you should create a Pull Requests targeting the release branch. 

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

